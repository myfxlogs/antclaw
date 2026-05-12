package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

type TraderHandler struct{ pool *pgxpool.Pool }

func NewTraderHandler(pool *pgxpool.Pool) *TraderHandler { return &TraderHandler{pool: pool} }

func (h *TraderHandler) GetProfile(ctx context.Context, req *connect.Request[alfqv1.GetTraderProfileRequest]) (*connect.Response[alfqv1.TraderProfile], error) {
	uid := req.Msg.UserId
	p := &alfqv1.TraderProfile{UserId: uid, Tier: "normal"}
	var name, bio string
	var ts time.Time
	h.pool.QueryRow(ctx, "SELECT COALESCE(display_name, email), '' FROM users WHERE id=$1", uid).Scan(&name, &bio)
	p.DisplayName = name
	p.Bio = bio
	h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alfq_follows WHERE following_id=$1", uid).Scan(&p.FollowerCount)
	h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alfq_follows WHERE follower_id=$1", uid).Scan(&p.FollowingCount)
	p.CreatedAt = ts.Unix()
	return connect.NewResponse(p), nil
}

func (h *TraderHandler) UpdateProfile(ctx context.Context, req *connect.Request[alfqv1.UpdateTraderProfileRequest]) (*connect.Response[alfqv1.TraderProfile], error) {
	uid := userIDFromHTTP(ctx, req)
	r := req.Msg
	_, _ = h.pool.Exec(ctx, "UPDATE users SET display_name=COALESCE(NULLIF($1,''), display_name) WHERE id=$2", r.DisplayName, uid)
	return h.GetProfile(ctx, connect.NewRequest(&alfqv1.GetTraderProfileRequest{UserId: uid}))
}

func (h *TraderHandler) Follow(ctx context.Context, req *connect.Request[alfqv1.FollowRequest]) (*connect.Response[alfqv1.FollowResponse], error) {
	uid := userIDFromHTTP(ctx, req)
	_, _ = h.pool.Exec(ctx, "INSERT INTO alfq_follows (follower_id, following_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", uid, req.Msg.TargetUserId)
	var cnt int32
	h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alfq_follows WHERE following_id=$1", req.Msg.TargetUserId).Scan(&cnt)
	return connect.NewResponse(&alfqv1.FollowResponse{Success: true, FollowerCount: cnt}), nil
}

func (h *TraderHandler) Unfollow(ctx context.Context, req *connect.Request[alfqv1.UnfollowRequest]) (*connect.Response[alfqv1.FollowResponse], error) {
	uid := userIDFromHTTP(ctx, req)
	h.pool.Exec(ctx, "DELETE FROM alfq_follows WHERE follower_id=$1 AND following_id=$2", uid, req.Msg.TargetUserId)
	var cnt int32
	h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM alfq_follows WHERE following_id=$1", req.Msg.TargetUserId).Scan(&cnt)
	return connect.NewResponse(&alfqv1.FollowResponse{Success: true, FollowerCount: cnt}), nil
}

func (h *TraderHandler) GetFollowers(ctx context.Context, req *connect.Request[alfqv1.GetFollowersRequest]) (*connect.Response[alfqv1.UserList], error) {
	rows, _ := h.pool.Query(ctx, `
		SELECT u.id, COALESCE(u.display_name, u.email), 'normal', 
		       (SELECT COUNT(*) FROM alfq_follows WHERE following_id=u.id)
		FROM alfq_follows f JOIN users u ON u.id=f.follower_id
		WHERE f.following_id=$1 LIMIT $2
	`, req.Msg.UserId, req.Msg.PageSize)
	defer rows.Close()
	var users []*alfqv1.UserInfo
	for rows.Next() {
		u := &alfqv1.UserInfo{}
		rows.Scan(&u.UserId, &u.DisplayName, &u.Tier, &u.FollowerCount)
		users = append(users, u)
	}
	return connect.NewResponse(&alfqv1.UserList{Users: users}), nil
}

func (h *TraderHandler) GetFollowing(ctx context.Context, req *connect.Request[alfqv1.GetFollowingRequest]) (*connect.Response[alfqv1.UserList], error) {
	rows, _ := h.pool.Query(ctx, `
		SELECT u.id, COALESCE(u.display_name, u.email), 'normal',
		       (SELECT COUNT(*) FROM alfq_follows WHERE following_id=u.id)
		FROM alfq_follows f JOIN users u ON u.id=f.following_id
		WHERE f.follower_id=$1 LIMIT $2
	`, req.Msg.UserId, req.Msg.PageSize)
	defer rows.Close()
	var users []*alfqv1.UserInfo
	for rows.Next() {
		u := &alfqv1.UserInfo{}
		rows.Scan(&u.UserId, &u.DisplayName, &u.Tier, &u.FollowerCount)
		users = append(users, u)
	}
	return connect.NewResponse(&alfqv1.UserList{Users: users}), nil
}

var _ antclawv1connect.TraderServiceHandler = (*TraderHandler)(nil)
