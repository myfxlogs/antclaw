// Package trader provides the business logic for trader profiles and follows.
package trader

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service"
	feedpkg "github.com/antclaw/antclaw/internal/service/feed"
)

// Service holds the trader business logic.
type Service struct {
	repo     postgres.TraderRepository
	limiter  feedpkg.SocialRateLimiter
	eventPub feedpkg.SocialEventPublisher
}

func NewService(repo postgres.TraderRepository) *Service {
	return &Service{repo: repo, limiter: feedpkg.NoopRateLimiter{}, eventPub: feedpkg.NoopEventPublisher{}}
}

// WithRateLimiter attaches a rate limiter for write operations (S12-P0-04).
func (s *Service) WithRateLimiter(limiter feedpkg.SocialRateLimiter) *Service {
	s.limiter = limiter
	return s
}

// WithEventPublisher attaches an event publisher for notification events (S12-P0-05).
func (s *Service) WithEventPublisher(pub feedpkg.SocialEventPublisher) *Service {
	s.eventPub = pub
	return s
}

// ----- Profile -----

func (s *Service) GetProfile(ctx context.Context, currentUserID string, req *alfqv1.GetTraderProfileRequest) (*alfqv1.TraderProfile, error) {
	if !s.userExists(ctx, req.UserId) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	row, err := s.repo.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get profile: %w", err))
	}
	if currentUserID != "" && currentUserID != req.UserId {
		if isFollowing, err := s.repo.IsFollowing(ctx, currentUserID, req.UserId); err != nil {
			feedpkg.NewSocialLogger().Error("is_following", currentUserID, "", err)
		} else {
			row.IsFollowing = isFollowing
		}
	}
	return traderProfileRowToProto(row), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req *alfqv1.UpdateTraderProfileRequest) (*alfqv1.TraderProfile, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	row := &postgres.TraderProfileRow{
		DisplayName:    req.DisplayName,
		Bio:            req.Bio,
		ShowWinRate:    req.ShowWinRate,
		ShowProfitFact: req.ShowProfitFactor,
		ShowSharpe:     req.ShowSharpe,
		ShowTotalTrad:  req.ShowTotalTrades,
	}
	if err := s.repo.UpdateProfile(ctx, userID, row); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update profile: %w", err))
	}
	return s.GetProfile(ctx, userID, &alfqv1.GetTraderProfileRequest{UserId: userID})
}

// ----- Follow -----

func (s *Service) Follow(ctx context.Context, userID string, req *alfqv1.FollowRequest) (*alfqv1.FollowResponse, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.limiter.Allow(ctx, userID, feedpkg.RateLimitFollow); err != nil {
		return nil, err
	}
	if userID == req.TargetUserId {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("cannot follow yourself"))
	}
	if !s.userExists(ctx, req.TargetUserId) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	if err := s.repo.Follow(ctx, userID, req.TargetUserId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("follow: %w", err))
	}
	// S12-P0-05: publish notification event
	followerName, _ := s.repo.GetUserName(ctx, userID)
	_ = s.eventPub.Publish(ctx, feedpkg.SocialEvent{
		Type:      "user_followed",
		ActorID:   userID,
		ActorName: followerName,
		TargetID:  req.TargetUserId,
	})
	cnt, err := s.repo.GetFollowerCount(ctx, req.TargetUserId)
	if err != nil {
		feedpkg.NewSocialLogger().Error("get_follower_count", req.TargetUserId, "", err)
	}
	return &alfqv1.FollowResponse{Success: true, FollowerCount: cnt}, nil
}

func (s *Service) Unfollow(ctx context.Context, userID string, req *alfqv1.UnfollowRequest) (*alfqv1.FollowResponse, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.repo.Unfollow(ctx, userID, req.TargetUserId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unfollow: %w", err))
	}
	cnt, err := s.repo.GetFollowerCount(ctx, req.TargetUserId)
	if err != nil {
		feedpkg.NewSocialLogger().Error("get_follower_count", req.TargetUserId, "", err)
	}
	return &alfqv1.FollowResponse{Success: true, FollowerCount: cnt}, nil
}

// ----- Paginated lists -----

func (s *Service) GetFollowers(ctx context.Context, req *alfqv1.GetFollowersRequest) (*alfqv1.UserList, error) {
	return s.listUsers(ctx, req.UserId, req.Cursor, req.PageSize, s.repo.GetFollowers)
}

func (s *Service) GetFollowing(ctx context.Context, req *alfqv1.GetFollowingRequest) (*alfqv1.UserList, error) {
	return s.listUsers(ctx, req.UserId, req.Cursor, req.PageSize, s.repo.GetFollowing)
}

// listUsersFactor is the function shape for GetFollowers/GetFollowing repository calls.
type listUsersFunc func(ctx context.Context, userID string, cursor *postgres.SocialCursor, limit int32) ([]*postgres.UserInfoRow, *postgres.SocialCursor, error)

// ListRecommendedTraders returns traders ranked by follower count (P2, algorithm option 1).
func (s *Service) ListRecommendedTraders(ctx context.Context, req *alfqv1.ListRecommendedTradersRequest) (*alfqv1.UserList, error) {
	limit := service.ClampPageSize(req.PageSize, 20, 50)
	cursor, err := postgres.DecodeSocialCursor(req.Cursor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, nextCursor, err := s.repo.ListRecommendedTraders(ctx, cursor, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list recommended traders: %w", err))
	}
	users := make([]*alfqv1.UserInfo, len(rows))
	for i, r := range rows {
		users[i] = &alfqv1.UserInfo{
			UserId:        r.UserID,
			DisplayName:   r.DisplayName,
			Tier:          r.Tier,
			FollowerCount: r.FollowerCount,
		}
	}
	return &alfqv1.UserList{
		Users:      users,
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
	}, nil
}

// listUsers centralizes "decode cursor → paginate → map rows" for follower/following lists.
func (s *Service) listUsers(ctx context.Context, userID, rawCursor string, pageSize int32, fn listUsersFunc) (*alfqv1.UserList, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id is required"))
	}
	limit := service.ClampPageSize(pageSize, 20, 50)
	cursor, err := postgres.DecodeSocialCursor(rawCursor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid cursor: %w", err))
	}
	rows, nextCursor, err := fn(ctx, userID, cursor, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list users: %w", err))
	}
	users := make([]*alfqv1.UserInfo, len(rows))
	for i, r := range rows {
		users[i] = &alfqv1.UserInfo{
			UserId:        r.UserID,
			DisplayName:   r.DisplayName,
			Tier:          r.Tier,
			FollowerCount: r.FollowerCount,
		}
	}
	return &alfqv1.UserList{
		Users:      users,
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
	}, nil
}

// ----- helpers -----

func (s *Service) userExists(ctx context.Context, userID string) bool {
	ok, err := s.repo.CheckUserExists(ctx, userID)
	return err == nil && ok
}

func traderProfileRowToProto(row *postgres.TraderProfileRow) *alfqv1.TraderProfile {
	return &alfqv1.TraderProfile{
		UserId:           row.UserID,
		DisplayName:      row.DisplayName,
		Bio:              row.Bio,
		Tier:             row.Tier,
		ShowWinRate:      row.ShowWinRate,
		ShowProfitFactor: row.ShowProfitFact,
		ShowSharpe:       row.ShowSharpe,
		ShowTotalTrades:  row.ShowTotalTrad,
		WinRate:          row.WinRate,
		ProfitFactor:     row.ProfitFactor,
		SharpeRatio:      row.SharpeRatio,
		TotalTrades:      row.TotalTrades,
		FollowerCount:    row.FollowerCount,
		FollowingCount:   row.FollowingCount,
		CreatedAt:        row.CreatedAt.Unix(),
		IsFollowing:      row.IsFollowing,
	}
}