// Package rpc provides the Trader Connect-RPC handler.
// Handler is a thin adapter: extracts user identity, delegates to trader.Service.
package rpc

import (
	"context"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/trader"
)

// TraderHandler implements TraderServiceHandler.
type TraderHandler struct {
	svc *trader.Service
}

// NewTraderHandler creates a new TraderHandler.
func NewTraderHandler(svc *trader.Service) *TraderHandler {
	return &TraderHandler{svc: svc}
}

// GetProfile retrieves a trader profile.
func (h *TraderHandler) GetProfile(ctx context.Context, req *connect.Request[alfqv1.GetTraderProfileRequest]) (*connect.Response[alfqv1.TraderProfile], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.GetProfile(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// UpdateProfile updates the trader's display name.
func (h *TraderHandler) UpdateProfile(ctx context.Context, req *connect.Request[alfqv1.UpdateTraderProfileRequest]) (*connect.Response[alfqv1.TraderProfile], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.UpdateProfile(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// Follow creates a follow relationship.
func (h *TraderHandler) Follow(ctx context.Context, req *connect.Request[alfqv1.FollowRequest]) (*connect.Response[alfqv1.FollowResponse], error) {
	userID := userIDFromHTTP(ctx, req)
	resp, err := h.svc.Follow(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// Unfollow removes a follow relationship.
func (h *TraderHandler) Unfollow(ctx context.Context, req *connect.Request[alfqv1.UnfollowRequest]) (*connect.Response[alfqv1.FollowResponse], error) {
	userID := userIDFromHTTP(ctx, req)
	resp, err := h.svc.Unfollow(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetFollowers returns followers with cursor pagination.
func (h *TraderHandler) GetFollowers(ctx context.Context, req *connect.Request[alfqv1.GetFollowersRequest]) (*connect.Response[alfqv1.UserList], error) {
	resp, err := h.svc.GetFollowers(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetFollowing returns users this user follows with cursor pagination.
func (h *TraderHandler) GetFollowing(ctx context.Context, req *connect.Request[alfqv1.GetFollowingRequest]) (*connect.Response[alfqv1.UserList], error) {
	resp, err := h.svc.GetFollowing(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.TraderServiceHandler = (*TraderHandler)(nil)
