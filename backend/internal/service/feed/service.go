// Package feed provides the business logic for the social feed.
package feed

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// Service holds the feed business logic.
type Service struct {
	repo postgres.FeedRepository
}

func NewService(repo postgres.FeedRepository) *Service {
	return &Service{repo: repo}
}

// clampPageSize returns ps clamped to [defaultVal, maxVal]. Zero or negative uses defaultVal.
func clampPageSize(ps, defaultVal, maxVal int32) int32 {
	if ps <= 0 {
		return defaultVal
	}
	if ps > maxVal {
		return maxVal
	}
	return ps
}

// normalizeFilter returns the canonical filter: all / signals_only / posts_only / shares.
func normalizeFilter(f string) string {
	switch f {
	case "signals_only", "posts_only", "shares":
		return f
	default:
		return "all"
	}
}

// checkPostAccessible verifies the post exists and currentUserID can view it.
// Returns the appropriate connect error on failure.
func (s *Service) checkPostAccessible(ctx context.Context, postID, currentUserID string) error {
	exists, err := s.repo.CheckPostExists(ctx, postID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("check post: %w", err))
	}
	if !exists {
		return connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	visible, err := s.repo.CheckPostVisibility(ctx, postID, currentUserID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("check visibility: %w", err))
	}
	if !visible {
		return connect.NewError(connect.CodePermissionDenied, errors.New("post not accessible"))
	}
	return nil
}

// ----- Post creation -----

func (s *Service) CreatePost(ctx context.Context, userID string, req *alfqv1.CreatePostRequest) (*alfqv1.Post, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if req.Content == "" && req.PostType != "signal_card" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("content is required"))
	}
	if req.Visibility == "circle" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("circle visibility not yet supported"))
	}
	if req.Visibility != "public" && req.Visibility != "followers" {
		req.Visibility = "public"
	}
	name, _ := s.repo.GetUserName(ctx, userID)

	row, err := s.repo.CreatePost(ctx, &postgres.FeedPostRow{
		AuthorID:         userID,
		AuthorName:       name,
		Content:          req.Content,
		PostType:         req.PostType,
		SignalPair:       req.SignalPair,
		SignalDirection:  req.SignalDirection,
		SignalConfidence: req.SignalConfidence,
		Visibility:       req.Visibility,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create post: %w", err))
	}
	return feedPostRowToProto(row, nil), nil
}

// ----- Feed listing -----

func (s *Service) GetFeed(ctx context.Context, userID string, req *alfqv1.GetFeedRequest) (*alfqv1.FeedResponse, error) {
	filter := normalizeFilter(req.Filter)
	limit := clampPageSize(req.PageSize, 20, 50)
	cursor, err := postgres.DecodeSocialCursor(req.Cursor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, likedByList, nextCursor, err := s.repo.GetFeed(ctx, filter, cursor, limit, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get feed: %w", err))
	}
	return &alfqv1.FeedResponse{
		Posts:      feedPostRowsToProto(rows, likedByList),
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
	}, nil
}

func (s *Service) GetPost(ctx context.Context, userID string, req *alfqv1.GetPostRequest) (*alfqv1.Post, error) {
	row, likedBy, err := s.repo.GetPost(ctx, req.PostId, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("post not found: %w", err))
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	return feedPostRowToProto(row, likedBy), nil
}

func (s *Service) ListUserPosts(ctx context.Context, currentUserID string, req *alfqv1.ListUserPostsRequest) (*alfqv1.FeedResponse, error) {
	if req.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id is required"))
	}
	filter := normalizeFilter(req.Filter)
	limit := clampPageSize(req.PageSize, 20, 50)
	cursor, err := postgres.DecodeSocialCursor(req.Cursor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, likedByList, nextCursor, err := s.repo.ListUserPosts(ctx, req.UserId, filter, cursor, limit, currentUserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list user posts: %w", err))
	}
	return &alfqv1.FeedResponse{
		Posts:      feedPostRowsToProto(rows, likedByList),
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
	}, nil
}

// ----- Likes -----

func (s *Service) LikePost(ctx context.Context, userID string, req *alfqv1.LikePostRequest) (*alfqv1.Post, error) {
	return s.toggleLike(ctx, userID, req.PostId, true)
}

func (s *Service) UnlikePost(ctx context.Context, userID string, req *alfqv1.UnlikePostRequest) (*alfqv1.Post, error) {
	return s.toggleLike(ctx, userID, req.PostId, false)
}

// toggleLike centralizes like/unlike logic. Idempotent for both operations.
func (s *Service) toggleLike(ctx context.Context, userID, postID string, isLike bool) (*alfqv1.Post, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.checkPostAccessible(ctx, postID, userID); err != nil {
		// NotFound from checkPostAccessible is fine; translate appropriately
		return nil, err
	}
	var err error
	if isLike {
		err = s.repo.LikePost(ctx, postID, userID)
	} else {
		err = s.repo.UnlikePost(ctx, postID, userID)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("toggle like: %w", err))
	}
	return s.GetPost(ctx, userID, &alfqv1.GetPostRequest{PostId: postID})
}

// ----- Comments -----

func (s *Service) CommentOnPost(ctx context.Context, userID string, req *alfqv1.CommentRequest) (*alfqv1.Comment, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if req.Content == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("content is required"))
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	name, _ := s.repo.GetUserName(ctx, userID)
	row := &postgres.FeedCommentRow{
		PostID:     req.PostId,
		AuthorID:   userID,
		AuthorName: name,
		Content:    req.Content,
	}
	if req.ParentCommentId != "" {
		pid := req.ParentCommentId
		row.ParentCommentID = &pid
	}
	row, err := s.repo.CreateComment(ctx, row)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create comment: %w", err))
	}
	return feedCommentRowToProto(row), nil
}

func (s *Service) ListComments(ctx context.Context, userID string, req *alfqv1.ListCommentsRequest) (*alfqv1.ListCommentsResponse, error) {
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	limit := clampPageSize(req.PageSize, 50, 100)
	cursor, err := postgres.DecodeSocialCursor(req.Cursor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, nextCursor, err := s.repo.ListComments(ctx, req.PostId, cursor, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list comments: %w", err))
	}
	comments := make([]*alfqv1.Comment, len(rows))
	for i, r := range rows {
		comments[i] = feedCommentRowToProto(r)
	}
	return &alfqv1.ListCommentsResponse{
		Comments:   comments,
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
	}, nil
}

// ----- Share -----

func (s *Service) SharePost(ctx context.Context, userID string, req *alfqv1.SharePostRequest) (*alfqv1.Post, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	name, _ := s.repo.GetUserName(ctx, userID)
	opid := req.PostId
	row, err := s.repo.CreatePost(ctx, &postgres.FeedPostRow{
		AuthorID:       userID,
		AuthorName:     name,
		Content:        req.Comment,
		PostType:       "share",
		Visibility:     "public",
		OriginalPostID: &opid,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("share post: %w", err))
	}
	return feedPostRowToProto(row, nil), nil
}

// ----- Proto mapping -----

func feedPostRowToProto(row *postgres.FeedPostRow, likedBy []string) *alfqv1.Post {
	p := &alfqv1.Post{
		Id:               row.ID,
		AuthorId:         row.AuthorID,
		AuthorName:       row.AuthorName,
		Content:          row.Content,
		PostType:         row.PostType,
		SignalPair:       row.SignalPair,
		SignalDirection:  row.SignalDirection,
		SignalConfidence: row.SignalConfidence,
		Visibility:       row.Visibility,
		LikeCount:        row.LikeCount,
		CommentCount:     row.CommentCount,
		ShareCount:       row.ShareCount,
		CreatedAt:        row.CreatedAt.Unix(),
	}
	if row.CircleID != nil {
		p.CircleId = *row.CircleID
	}
	if row.OriginalPostID != nil {
		p.OriginalPostId = *row.OriginalPostID
	}
	if likedBy != nil {
		p.LikedBy = likedBy
	}
	return p
}

func feedPostRowsToProto(rows []*postgres.FeedPostRow, likedByList [][]string) []*alfqv1.Post {
	posts := make([]*alfqv1.Post, len(rows))
	for i, r := range rows {
		var lb []string
		if i < len(likedByList) {
			lb = likedByList[i]
		}
		posts[i] = feedPostRowToProto(r, lb)
	}
	return posts
}

func feedCommentRowToProto(row *postgres.FeedCommentRow) *alfqv1.Comment {
	c := &alfqv1.Comment{
		Id:         row.ID,
		PostId:     row.PostID,
		AuthorId:   row.AuthorID,
		AuthorName: row.AuthorName,
		Content:    row.Content,
		CreatedAt:  row.CreatedAt.Unix(),
	}
	if row.ParentCommentID != nil {
		c.ParentCommentId = *row.ParentCommentID
	}
	return c
}
