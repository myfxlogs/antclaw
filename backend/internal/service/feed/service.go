// Package feed provides the business logic for the social feed.
package feed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	infraredis "github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service"
)

// Service holds the feed business logic.
type Service struct {
	repo     postgres.FeedRepository
	rdb      *infraredis.Client
	limiter  SocialRateLimiter
	eventPub SocialEventPublisher
}

func NewService(repo postgres.FeedRepository) *Service {
	return &Service{repo: repo, limiter: NoopRateLimiter{}, eventPub: NoopEventPublisher{}}
}

func NewServiceWithRedis(repo postgres.FeedRepository, rdb *infraredis.Client) *Service {
	return &Service{repo: repo, rdb: rdb, limiter: NoopRateLimiter{}, eventPub: NoopEventPublisher{}}
}

// WithRateLimiter attaches a rate limiter for write operations (S12-P0-04).
func (s *Service) WithRateLimiter(limiter SocialRateLimiter) *Service {
	s.limiter = limiter
	return s
}

// WithEventPublisher attaches an event publisher for notification events (S12-P0-05).
func (s *Service) WithEventPublisher(pub SocialEventPublisher) *Service {
	s.eventPub = pub
	return s
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
	if err := s.limiter.Allow(ctx, userID, RateLimitCreatePost); err != nil {
		return nil, err
	}
	// S12-P0-03: unified content validation
	if err := ValidateCreatePostRequest(req.Content, req.PostType, req.SignalDirection, req.SignalConfidence, req.Visibility); err != nil {
		return nil, err
	}
	// S12-P0-01: strict visibility validation — no silent fallback to public
	vis := strings.TrimSpace(req.Visibility)
	if vis == "" {
		vis = "public" // default per product spec
	}
	switch vis {
	case "public", "followers":
		// allowed
	case "circle":
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("circle visibility not yet supported"))
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("visibility must be public or followers"))
	}
	req.Visibility = vis
	name, err := s.repo.GetUserName(ctx, userID)
	if err != nil {
		NewSocialLogger().Error("get_user_name", userID, "", err)
		name = ""
	}

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
	return PostRowToProto(row, nil), nil
}

// ----- Feed listing -----

func (s *Service) GetFeed(ctx context.Context, userID string, req *alfqv1.GetFeedRequest) (*alfqv1.FeedResponse, error) {
	filter := normalizeFilter(req.Filter)
	limit := service.ClampPageSize(req.PageSize, 20, 50)
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

// GetCachedFeed returns feed with Redis caching (2min TTL).
func (s *Service) GetCachedFeed(ctx context.Context, userID, filter, cursor string, limit int32, pg *pgxpool.Pool) ([]*alfqv1.Post, string, error) {
	key := fmt.Sprintf("feed:%s:%s:%s:%d", userID, filter, cursor, limit)
	if s.rdb != nil {
		if cached, err := s.rdb.Get(ctx, key); err == nil && cached != "" {
			var wrapper struct {
				Posts []*alfqv1.Post `json:"posts"`
				Next  string         `json:"next"`
			}
			if json.Unmarshal([]byte(cached), &wrapper) == nil && len(wrapper.Posts) > 0 {
				return wrapper.Posts, wrapper.Next, nil
			}
		}
	}
	// Fetch + rank + filter
	c, _ := postgres.DecodeSocialCursor(cursor)
	rows, likedByList, nextCursor, err := s.repo.GetFeed(ctx, filter, c, limit, userID)
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInternal, fmt.Errorf("get feed: %w", err))
	}
	rows = RankPosts(rows)
	if pg != nil {
		rows = ApplyFilters(ctx, pg, rows)
	}
	posts := feedPostRowsToProto(rows, likedByList)
	next := postgres.EncodeSocialCursor(nextCursor)
	// Write cache
	if s.rdb != nil {
		wrapper, _ := json.Marshal(struct {
			Posts []*alfqv1.Post `json:"posts"`
			Next  string         `json:"next"`
		}{posts, next})
		s.rdb.Set(ctx, key, string(wrapper), 2*time.Minute)
	}
	return posts, next, nil
}

func (s *Service) GetFollowingFeed(ctx context.Context, userID string, cursor string, pageSize int32) ([]*alfqv1.Post, string, error) {
	if userID == "" {
		return nil, "", connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	limit := service.ClampPageSize(pageSize, 20, 50)
	c, err := postgres.DecodeSocialCursor(cursor)
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, likedByList, nextCursor, err := s.repo.GetFollowingFeed(ctx, userID, c, limit, userID)
	if err != nil {
		return nil, "", connect.NewError(connect.CodeInternal, fmt.Errorf("get following feed: %w", err))
	}
	return feedPostRowsToProto(rows, likedByList), postgres.EncodeSocialCursor(nextCursor), nil
}

func (s *Service) GetPost(ctx context.Context, userID string, req *alfqv1.GetPostRequest) (*alfqv1.Post, error) {
	row, likedBy, err := s.repo.GetPost(ctx, req.PostId, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("post not found: %w", err))
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	return PostRowToProto(row, likedBy), nil
}

func (s *Service) ListUserPosts(ctx context.Context, currentUserID string, req *alfqv1.ListUserPostsRequest) (*alfqv1.FeedResponse, error) {
	if req.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id is required"))
	}
	filter := normalizeFilter(req.Filter)
	limit := service.ClampPageSize(req.PageSize, 20, 50)
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
	if err := s.limiter.Allow(ctx, userID, RateLimitLikePost); err != nil {
		return nil, err
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
	if err := s.limiter.Allow(ctx, userID, RateLimitCommentOnPost); err != nil {
		return nil, err
	}
	if err := ValidateCommentRequest(req.Content); err != nil {
		return nil, err
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	name, err := s.repo.GetUserName(ctx, userID)
	if err != nil {
		NewSocialLogger().Error("get_user_name", userID, req.PostId, err)
		name = ""
	}
	row := &postgres.FeedCommentRow{
		PostID:     req.PostId,
		AuthorID:   userID,
		AuthorName: name,
		Content:    req.Content,
	}
	if req.ParentCommentId != "" {
		ok, err := s.repo.CheckCommentPostID(ctx, req.ParentCommentId, req.PostId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("check parent comment: %w", err))
		}
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("parent_comment_id does not belong to the same post"))
		}
		pid := req.ParentCommentId
		row.ParentCommentID = &pid
	}
	row, err = s.repo.CreateComment(ctx, row)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create comment: %w", err))
	}
	// S12-P0-05: publish notification event (non-blocking, failure logged)
	// Don't notify when commenting on own post
	postRow, _, _ := s.repo.GetPost(ctx, req.PostId, "")
	if postRow != nil && postRow.AuthorID != userID {
		_ = s.eventPub.Publish(ctx, SocialEvent{
			Type:       "post_commented",
			ActorID:    userID,
			ActorName:  name,
			TargetID:   postRow.AuthorID,
			PostID:     req.PostId,
			PostTitle:  postRow.Content,
			CommentID:  row.ID,
		})
	}
	return feedCommentRowToProto(row), nil
}

func (s *Service) ListComments(ctx context.Context, userID string, req *alfqv1.ListCommentsRequest) (*alfqv1.ListCommentsResponse, error) {
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	limit := service.ClampPageSize(req.PageSize, 50, 100)
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
	if err := s.limiter.Allow(ctx, userID, RateLimitSharePost); err != nil {
		return nil, err
	}
	if err := ValidateSharePostRequest(req.Comment); err != nil {
		return nil, err
	}
	if err := s.checkPostAccessible(ctx, req.PostId, userID); err != nil {
		return nil, err
	}
	name, err := s.repo.GetUserName(ctx, userID)
	if err != nil {
		NewSocialLogger().Error("get_user_name", userID, req.PostId, err)
		name = ""
	}
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
	// S12-P0-05: notify original post author (skip if sharing own post)
	origRow, _, _ := s.repo.GetPost(ctx, req.PostId, "")
	if origRow != nil && origRow.AuthorID != userID {
		_ = s.eventPub.Publish(ctx, SocialEvent{
			Type:       "post_shared",
			ActorID:    userID,
			ActorName:  name,
			TargetID:   origRow.AuthorID,
			PostID:     req.PostId,
			PostTitle:  origRow.Content,
		})
	}
	return PostRowToProto(row, nil), nil
}

// ----- Proto mapping -----

// PostRowToProto maps a repository row to a proto Post message.
func PostRowToProto(row *postgres.FeedPostRow, likedBy []string) *alfqv1.Post {
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
		posts[i] = PostRowToProto(r, lb)
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
