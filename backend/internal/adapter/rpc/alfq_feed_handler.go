// Package rpc provides the Feed Connect-RPC handler.
// Handler is a thin adapter: extracts user identity, delegates to feed.Service.
package rpc

import (
	"context"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/feed"
)

// FeedHandler implements FeedServiceHandler.
type FeedHandler struct {
	svc *feed.Service
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(svc *feed.Service) *FeedHandler {
	return &FeedHandler{svc: svc}
}

// CreatePost creates a new post.
func (h *FeedHandler) CreatePost(ctx context.Context, req *connect.Request[alfqv1.CreatePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.CreatePost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// GetFeed returns the public feed with cursor pagination and filtering.
func (h *FeedHandler) GetFeed(ctx context.Context, req *connect.Request[alfqv1.GetFeedRequest]) (*connect.Response[alfqv1.FeedResponse], error) {
	userID := userIDFromHTTP(ctx, req)
	resp, err := h.svc.GetFeed(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// GetPost retrieves a single post by ID.
func (h *FeedHandler) GetPost(ctx context.Context, req *connect.Request[alfqv1.GetPostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.GetPost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// LikePost likes a post. Idempotent.
func (h *FeedHandler) LikePost(ctx context.Context, req *connect.Request[alfqv1.LikePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.LikePost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// UnlikePost removes a like. Idempotent.
func (h *FeedHandler) UnlikePost(ctx context.Context, req *connect.Request[alfqv1.UnlikePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.UnlikePost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// CommentOnPost creates a comment on a post.
func (h *FeedHandler) CommentOnPost(ctx context.Context, req *connect.Request[alfqv1.CommentRequest]) (*connect.Response[alfqv1.Comment], error) {
	userID := userIDFromHTTP(ctx, req)
	c, err := h.svc.CommentOnPost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(c), nil
}

// ListComments lists comments for a post with cursor pagination.
func (h *FeedHandler) ListComments(ctx context.Context, req *connect.Request[alfqv1.ListCommentsRequest]) (*connect.Response[alfqv1.ListCommentsResponse], error) {
	userID := userIDFromHTTP(ctx, req)
	resp, err := h.svc.ListComments(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// SharePost shares (quotes) a post.
func (h *FeedHandler) SharePost(ctx context.Context, req *connect.Request[alfqv1.SharePostRequest]) (*connect.Response[alfqv1.Post], error) {
	userID := userIDFromHTTP(ctx, req)
	p, err := h.svc.SharePost(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(p), nil
}

// ListUserPosts returns posts by a specific user.
func (h *FeedHandler) ListUserPosts(ctx context.Context, req *connect.Request[alfqv1.ListUserPostsRequest]) (*connect.Response[alfqv1.FeedResponse], error) {
	userID := userIDFromHTTP(ctx, req)
	resp, err := h.svc.ListUserPosts(ctx, userID, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.FeedServiceHandler = (*FeedHandler)(nil)
