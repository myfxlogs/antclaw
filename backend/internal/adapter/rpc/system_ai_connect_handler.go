package rpc

import (
	"context"
	"errors"
	"log"
	"time"

	"connectrpc.com/connect"
	systemaiv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/systemai"
)

type SystemAIConnectHandler struct {
	svc *systemai.Service
}

func NewSystemAIConnectHandler(svc *systemai.Service) *SystemAIConnectHandler {
	return &SystemAIConnectHandler{svc: svc}
}

func toSystemAIProto(cfg systemai.Config) *systemaiv1.SystemAIConfig {
	return &systemaiv1.SystemAIConfig{
		ProviderId:     cfg.ProviderID,
		Name:           cfg.Name,
		BaseUrl:        cfg.BaseURL,
		Organization:   cfg.Organization,
		Models:         cfg.Models,
		DefaultModel:   cfg.DefaultModel,
		Temperature:    cfg.Temperature,
		TimeoutSeconds: int32(cfg.TimeoutSeconds),
		MaxTokens:      int32(cfg.MaxTokens),
		Purposes:       cfg.Purposes,
		PrimaryFor:     cfg.PrimaryFor,
		Enabled:        cfg.Enabled,
		HasSecret:      cfg.HasSecret,
		CreatedAt:      cfg.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      cfg.UpdatedAt.Format(time.RFC3339),
		UpdatedBy:      cfg.UpdatedBy,
		DocsUrl:        cfg.DocsURL,
		ApplyUrl:       cfg.ApplyURL,
	}
}

func (h *SystemAIConnectHandler) ListConfigs(ctx context.Context, _ *connect.Request[systemaiv1.ListSystemAIConfigsRequest]) (*connect.Response[systemaiv1.ListSystemAIConfigsResponse], error) {
	configs, err := h.svc.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*systemaiv1.SystemAIConfig, 0, len(configs))
	for _, cfg := range configs {
		items = append(items, toSystemAIProto(cfg))
	}
	return connect.NewResponse(&systemaiv1.ListSystemAIConfigsResponse{Items: items}), nil
}

func (h *SystemAIConnectHandler) GetConfig(ctx context.Context, req *connect.Request[systemaiv1.GetSystemAIConfigRequest]) (*connect.Response[systemaiv1.GetSystemAIConfigResponse], error) {
	cfg, err := h.svc.Get(ctx, req.Msg.ProviderId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&systemaiv1.GetSystemAIConfigResponse{Item: toSystemAIProto(*cfg)}), nil
}

func (h *SystemAIConnectHandler) UpdateConfig(ctx context.Context, req *connect.Request[systemaiv1.UpdateSystemAIConfigRequest]) (*connect.Response[systemaiv1.UpdateSystemAIConfigResponse], error) {
	cfg := &systemai.Config{
		ProviderID:     req.Msg.ProviderId,
		Name:           req.Msg.Name,
		BaseURL:        req.Msg.BaseUrl,
		Organization:   req.Msg.Organization,
		Models:         req.Msg.Models,
		DefaultModel:   req.Msg.DefaultModel,
		Temperature:    req.Msg.Temperature,
		TimeoutSeconds: int(req.Msg.TimeoutSeconds),
		MaxTokens:      int(req.Msg.MaxTokens),
		Purposes:       req.Msg.Purposes,
		PrimaryFor:     req.Msg.PrimaryFor,
		Enabled:        req.Msg.Enabled,
	}
	if err := h.svc.UpdateConfig(ctx, cfg, currentUser(ctx)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&systemaiv1.UpdateSystemAIConfigResponse{ProviderId: req.Msg.ProviderId}), nil
}

func (h *SystemAIConnectHandler) UpdateSecret(ctx context.Context, req *connect.Request[systemaiv1.UpdateSystemAISecretRequest]) (*connect.Response[systemaiv1.UpdateSystemAISecretResponse], error) {
	if err := h.svc.UpdateSecret(ctx, req.Msg.ProviderId, req.Msg.Secret, currentUser(ctx)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&systemaiv1.UpdateSystemAISecretResponse{
		ProviderId:    req.Msg.ProviderId,
		SecretUpdated: true,
	}), nil
}

func (h *SystemAIConnectHandler) DiscoverModels(ctx context.Context, req *connect.Request[systemaiv1.DiscoverSystemAIModelsRequest]) (*connect.Response[systemaiv1.DiscoverSystemAIModelsResponse], error) {
	models, err := h.svc.DiscoverModels(ctx, req.Msg.ProviderId)
	if err != nil {
		log.Printf("systemai: DiscoverModels provider=%s raw_err=%v", req.Msg.ProviderId, err)
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(systemai.FriendlyDiscoverError(err)))
	}
	if len(models) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no models discovered from provider"))
	}
	cfg, err := h.svc.Get(ctx, req.Msg.ProviderId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	cfg.Models = models
	cfg.DefaultModel = models[0]
	if err := h.svc.UpdateConfig(ctx, cfg, currentUser(ctx)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&systemaiv1.DiscoverSystemAIModelsResponse{
		ProviderId:   req.Msg.ProviderId,
		Models:       models,
		DefaultModel: models[0],
	}), nil
}

func (h *SystemAIConnectHandler) ValidateConnection(ctx context.Context, req *connect.Request[systemaiv1.ValidateSystemAIConnectionRequest]) (*connect.Response[systemaiv1.ValidateSystemAIConnectionResponse], error) {
	models, err := h.svc.DiscoverModels(ctx, req.Msg.ProviderId)
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(systemai.FriendlyDiscoverError(err)))
	}
	return connect.NewResponse(&systemaiv1.ValidateSystemAIConnectionResponse{
		ProviderId: req.Msg.ProviderId,
		Ok:         true,
		ModelCount: int32(len(models)),
	}), nil
}

var _ antclawv1connect.SystemAIServiceHandler = (*SystemAIConnectHandler)(nil)
