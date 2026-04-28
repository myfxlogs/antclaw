package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	datasourcev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/datasource"
)

type DataSourceConnectHandler struct {
	svc *datasource.Service
}

func NewDataSourceConnectHandler(svc *datasource.Service) *DataSourceConnectHandler {
	return &DataSourceConnectHandler{svc: svc}
}

func toDataSourceProto(cfg datasource.Config) *datasourcev1.DataSourceConfig {
	return &datasourcev1.DataSourceConfig{
		SourceId:  cfg.SourceID,
		Name:      cfg.Name,
		Kind:      cfg.Kind,
		Endpoint:  cfg.Endpoint,
		HasSecret: cfg.HasSecret,
		UpdatedAt: cfg.UpdatedAt.Format(time.RFC3339),
		UpdatedBy: cfg.UpdatedBy,
	}
}

func (h *DataSourceConnectHandler) ListDataSources(ctx context.Context, _ *connect.Request[datasourcev1.ListDataSourcesRequest]) (*connect.Response[datasourcev1.ListDataSourcesResponse], error) {
	items, err := h.svc.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*datasourcev1.DataSourceConfig, 0, len(items))
	for _, item := range items {
		out = append(out, toDataSourceProto(item))
	}
	return connect.NewResponse(&datasourcev1.ListDataSourcesResponse{Items: out}), nil
}

func (h *DataSourceConnectHandler) UpdateDataSource(ctx context.Context, req *connect.Request[datasourcev1.UpdateDataSourceRequest]) (*connect.Response[datasourcev1.UpdateDataSourceResponse], error) {
	in := datasource.UpdateInput{
		ClearSecret: req.Msg.ClearSecret,
		UpdatedBy:   currentUser(ctx),
	}
	if req.Msg.Endpoint != nil {
		in.Endpoint = req.Msg.Endpoint
	}
	if req.Msg.Secret != nil {
		in.Secret = req.Msg.Secret
	}
	if err := h.svc.Update(ctx, req.Msg.SourceId, in); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	item, err := h.svc.Get(ctx, req.Msg.SourceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&datasourcev1.UpdateDataSourceResponse{Item: toDataSourceProto(*item)}), nil
}

var _ antclawv1connect.DataSourceServiceHandler = (*DataSourceConnectHandler)(nil)
