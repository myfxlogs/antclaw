package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	antclawv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SystemHandler implements SystemService.
type SystemHandler struct {
	pg    *pgxpool.Pool
	redis *redis.Client
	boot  time.Time
}

func NewSystemHandler(pg *pgxpool.Pool, r *redis.Client, boot time.Time) *SystemHandler {
	return &SystemHandler{pg: pg, redis: r, boot: boot}
}

func (h *SystemHandler) Healthz(ctx context.Context, _ *connect.Request[antclawv1.HealthzRequest]) (*connect.Response[antclawv1.HealthzResponse], error) {
	status := "healthy"
	components := map[string]*antclawv1.ComponentHealth{}
	// postgres
	pgStatus := &antclawv1.ComponentHealth{Name: "postgres", Status: "up"}
	if err := h.pg.Ping(ctx); err != nil {
		pgStatus.Status = "down"
		pgStatus.Detail = err.Error()
		status = "unhealthy"
	}
	components["postgres"] = pgStatus
	// redis
	rStatus := &antclawv1.ComponentHealth{Name: "redis", Status: "up"}
	if err := h.redis.Ping(ctx); err != nil {
		rStatus.Status = "down"
		rStatus.Detail = err.Error()
		status = "unhealthy"
	}
	components["redis"] = rStatus
	resp := &antclawv1.HealthzResponse{Status: status, Components: components, CheckedAt: timestamppb.Now()}
	return connect.NewResponse(resp), nil
}

func (h *SystemHandler) Readyz(ctx context.Context, _ *connect.Request[antclawv1.ReadyzRequest]) (*connect.Response[antclawv1.ReadyzResponse], error) {
	ready := time.Since(h.boot) > 5*time.Second
	return connect.NewResponse(&antclawv1.ReadyzResponse{Ready: ready}), nil
}

func (h *SystemHandler) Info(ctx context.Context, _ *connect.Request[antclawv1.InfoRequest]) (*connect.Response[antclawv1.InfoResponse], error) {
	_ = ctx
	// 版本信息可从环境注入，这里以占位符返回
	return connect.NewResponse(&antclawv1.InfoResponse{
		Version:   "0.1.0",
		GitCommit: "dev",
		BuiltAt:   timestamppb.New(h.boot),
	}), nil
}

var _ antclawv1connect.SystemServiceHandler = (*SystemHandler)(nil)
