// Package trader provides the business logic for trader profiles and follows.
package trader

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// Service holds the trader business logic.
type Service struct {
	repo postgres.TraderRepository
}

func NewService(repo postgres.TraderRepository) *Service {
	return &Service{repo: repo}
}

// clampPageSize returns ps clamped to [defaultVal, maxVal].
func clampPageSize(ps, defaultVal, maxVal int32) int32 {
	if ps <= 0 {
		return defaultVal
	}
	if ps > maxVal {
		return maxVal
	}
	return ps
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
		row.IsFollowing, _ = s.repo.IsFollowing(ctx, currentUserID, req.UserId)
	}
	return traderProfileRowToProto(row), nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req *alfqv1.UpdateTraderProfileRequest) (*alfqv1.TraderProfile, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.repo.UpdateProfile(ctx, userID, req.DisplayName); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("update profile: %w", err))
	}
	return s.GetProfile(ctx, userID, &alfqv1.GetTraderProfileRequest{UserId: userID})
}

// ----- Follow -----

func (s *Service) Follow(ctx context.Context, userID string, req *alfqv1.FollowRequest) (*alfqv1.FollowResponse, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
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
	cnt, _ := s.repo.GetFollowerCount(ctx, req.TargetUserId)
	return &alfqv1.FollowResponse{Success: true, FollowerCount: cnt}, nil
}

func (s *Service) Unfollow(ctx context.Context, userID string, req *alfqv1.UnfollowRequest) (*alfqv1.FollowResponse, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("login required"))
	}
	if err := s.repo.Unfollow(ctx, userID, req.TargetUserId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unfollow: %w", err))
	}
	cnt, _ := s.repo.GetFollowerCount(ctx, req.TargetUserId)
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

// listUsers centralizes "decode cursor → paginate → map rows" for follower/following lists.
func (s *Service) listUsers(ctx context.Context, userID, rawCursor string, pageSize int32, fn listUsersFunc) (*alfqv1.UserList, error) {
	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user_id is required"))
	}
	limit := clampPageSize(pageSize, 20, 50)
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
