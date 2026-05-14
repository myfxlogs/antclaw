package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

type FeedHandler struct {
	pool *pgxpool.Pool
}

func NewFeedHandler(pool *pgxpool.Pool) *FeedHandler {
	return &FeedHandler{pool: pool}
}

func (h *FeedHandler) CreatePost(ctx context.Context, req *connect.Request[alfqv1.CreatePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}
	// Get user name
	var userName string
	h.pool.QueryRow(ctx, "SELECT COALESCE(display_name, email) FROM users WHERE id=$1", userID).Scan(&userName)

	r := req.Msg
	var id string
	row := h.pool.QueryRow(ctx, `
		INSERT INTO alfq_posts (author_id, author_name, content, post_type, signal_pair, signal_direction, signal_confidence, visibility)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at
	`, userID, userName, r.Content, r.PostType, r.SignalPair, r.SignalDirection, r.SignalConfidence, r.Visibility)

	p := &alfqv1.Post{
		AuthorId: userID, AuthorName: userName,
		Content: r.Content, PostType: r.PostType,
		SignalPair: r.SignalPair, SignalDirection: r.SignalDirection,
		SignalConfidence: r.SignalConfidence, Visibility: r.Visibility,
	}
	var ts time.Time
	if err := row.Scan(&id, &ts); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	p.Id = id
	p.CreatedAt = ts.Unix()
	return connect.NewResponse(p), nil
}

func (h *FeedHandler) GetFeed(ctx context.Context, req *connect.Request[alfqv1.GetFeedRequest]) (*connect.Response[alfqv1.FeedResponse], error) {
	r := req.Msg
	rows, err := h.pool.Query(ctx, `
		SELECT id, author_id, author_name, content, post_type, signal_pair, signal_direction, signal_confidence, visibility,
		       (SELECT COUNT(*) FROM alfq_likes WHERE post_id=p.id) as like_count,
		       (SELECT COUNT(*) FROM alfq_comments WHERE post_id=p.id) as comment_count,
		       created_at
		FROM alfq_posts p
		WHERE visibility='public' AND ($1='' OR created_at < (SELECT created_at FROM alfq_posts WHERE id=$1::uuid))
		ORDER BY created_at DESC LIMIT $2
	`, r.Cursor, minInt(r.PageSize, 20))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()

	var posts []*alfqv1.Post
	for rows.Next() {
		p := &alfqv1.Post{}
		var ts time.Time
		if err := rows.Scan(&p.Id, &p.AuthorId, &p.AuthorName, &p.Content, &p.PostType, &p.SignalPair, &p.SignalDirection, &p.SignalConfidence, &p.Visibility, &p.LikeCount, &p.CommentCount, &ts); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		p.CreatedAt = ts.Unix()
		posts = append(posts, p)
	}
	nextCursor := ""
	if len(posts) > 0 {
		nextCursor = posts[len(posts)-1].Id
	}
	return connect.NewResponse(&alfqv1.FeedResponse{Posts: posts, NextCursor: nextCursor}), nil
}

func (h *FeedHandler) GetPost(ctx context.Context, req *connect.Request[alfqv1.GetPostRequest]) (*connect.Response[alfqv1.Post], error) {
	// simplified: return from DB
	p := &alfqv1.Post{}
	var ts time.Time
	err := h.pool.QueryRow(ctx, `
		SELECT id, author_id, author_name, content, post_type, signal_pair, signal_direction, signal_confidence, visibility, created_at
		FROM alfq_posts WHERE id=$1
	`, req.Msg.PostId).Scan(&p.Id, &p.AuthorId, &p.AuthorName, &p.Content, &p.PostType, &p.SignalPair, &p.SignalDirection, &p.SignalConfidence, &p.Visibility, &ts)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	p.CreatedAt = ts.Unix()
	return connect.NewResponse(p), nil
}

func (h *FeedHandler) LikePost(ctx context.Context, req *connect.Request[alfqv1.LikePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	_, err := h.pool.Exec(ctx, "INSERT INTO alfq_likes (post_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", req.Msg.PostId, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return h.GetPost(ctx, connect.NewRequest(&alfqv1.GetPostRequest{PostId: req.Msg.PostId}))
}

func (h *FeedHandler) UnlikePost(ctx context.Context, req *connect.Request[alfqv1.UnlikePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	h.pool.Exec(ctx, "DELETE FROM alfq_likes WHERE post_id=$1 AND user_id=$2", req.Msg.PostId, userID)
	return h.GetPost(ctx, connect.NewRequest(&alfqv1.GetPostRequest{PostId: req.Msg.PostId}))
}

func (h *FeedHandler) CommentOnPost(ctx context.Context, req *connect.Request[alfqv1.CommentRequest]) (*connect.Response[alfqv1.Comment], error) {
	userID := userIDFromHTTP(ctx, req)
	var userName string
	h.pool.QueryRow(ctx, "SELECT COALESCE(display_name, email) FROM users WHERE id=$1", userID).Scan(&userName)
	r := req.Msg
	c := &alfqv1.Comment{AuthorId: userID, AuthorName: userName, Content: r.Content, PostId: r.PostId}
	var ts time.Time
	err := h.pool.QueryRow(ctx, `
		INSERT INTO alfq_comments (post_id, author_id, author_name, content)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at
	`, r.PostId, userID, userName, r.Content).Scan(&c.Id, &ts)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	c.CreatedAt = ts.Unix()
	return connect.NewResponse(c), nil
}

func (h *FeedHandler) SharePost(ctx context.Context, req *connect.Request[alfqv1.SharePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	var userName string
	h.pool.QueryRow(ctx, "SELECT COALESCE(display_name, email) FROM users WHERE id=$1", userID).Scan(&userName)
	p := &alfqv1.Post{AuthorId: userID, AuthorName: userName, Content: req.Msg.Comment, PostType: "share"}
	var ts time.Time
	var id string
	err := h.pool.QueryRow(ctx, `
		INSERT INTO alfq_posts (author_id, author_name, content, post_type, visibility)
		VALUES ($1,$2,$3,'share','public') RETURNING id, created_at
	`, userID, userName, req.Msg.Comment).Scan(&id, &ts)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	p.Id = id
	p.CreatedAt = ts.Unix()
	return connect.NewResponse(p), nil
}

func minInt(a, b int32) int32 {
	if a == 0 || a > b { return b }
	return a
}

var _ antclawv1connect.FeedServiceHandler = (*FeedHandler)(nil)
