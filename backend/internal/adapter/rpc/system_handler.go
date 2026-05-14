package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	antclawv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/service/presence"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SystemHandler implements SystemService.
type SystemHandler struct {
	pg       *pgxpool.Pool
	redis    *redis.Client
	boot     time.Time
	presence *presence.Tracker
}

func NewSystemHandler(pg *pgxpool.Pool, r *redis.Client, boot time.Time, pt *presence.Tracker) *SystemHandler {
	return &SystemHandler{pg: pg, redis: r, boot: boot, presence: pt}
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

func (h *SystemHandler) Info(_ context.Context, _ *connect.Request[antclawv1.InfoRequest]) (*connect.Response[antclawv1.InfoResponse], error) {
	return connect.NewResponse(&antclawv1.InfoResponse{
		Version:          "0.1.0",
		GitCommit:        "dev",
		BuiltAt:          timestamppb.New(h.boot),
		ProtoVersion:     "1.0.0",
		MinClientVersion: "1.0.0",
		MaintenanceMode:  false,
		ServerTimezone:   "UTC",
		ServerTime:       time.Now().Unix(),
	}), nil
}

func (h *SystemHandler) GetOnlineUsers(ctx context.Context, _ *connect.Request[antclawv1.GetOnlineUsersRequest]) (*connect.Response[antclawv1.GetOnlineUsersResponse], error) {
	list := h.presence.List()

	// 批量查询 code_id 以替代 UUID 展示
	codeIDMap := make(map[string]string, len(list))
	if len(list) > 0 && h.pg != nil {
		ids := make([]string, 0, len(list))
		for _, u := range list {
			ids = append(ids, u.UserID)
		}
		rows, err := h.pg.Query(ctx,
			`SELECT id::text, COALESCE(code_id, '') FROM users WHERE id::text = ANY($1)`,
			ids,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var uid, cid string
				if err := rows.Scan(&uid, &cid); err == nil {
					codeIDMap[uid] = cid
				}
			}
		}
	}

	users := make([]*antclawv1.OnlineUserInfo, 0, len(list))
	for _, u := range list {
		users = append(users, &antclawv1.OnlineUserInfo{
			UserId:      u.UserID,
			RemoteAddr:  u.RemoteAddr,
			ConnectedAt: u.ConnectedAt.Unix(),
			CodeId:      codeIDMap[u.UserID],
		})
	}
	return connect.NewResponse(&antclawv1.GetOnlineUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}), nil
}

func (h *SystemHandler) GetPushStats(ctx context.Context, _ *connect.Request[antclawv1.GetPushStatsRequest]) (*connect.Response[antclawv1.GetPushStatsResponse], error) {
	var totalNotif, totalPush, recent1h, recent24h int64
	_ = h.pg.QueryRow(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&totalNotif)
	_ = h.pg.QueryRow(ctx, `SELECT COUNT(*) FROM notification_push_state`).Scan(&totalPush)
	_ = h.pg.QueryRow(ctx, `SELECT COUNT(*) FROM notification_push_state WHERE last_sent_at >= NOW() - INTERVAL '1 hour'`).Scan(&recent1h)
	_ = h.pg.QueryRow(ctx, `SELECT COUNT(*) FROM notification_push_state WHERE last_sent_at >= NOW() - INTERVAL '24 hours'`).Scan(&recent24h)

	rows, err := h.pg.Query(ctx, `SELECT push_type, COUNT(*) as cnt FROM notification_push_state GROUP BY push_type ORDER BY cnt DESC LIMIT 20`)
	if err != nil {
		return connect.NewResponse(&antclawv1.GetPushStatsResponse{
			TotalNotifications:     totalNotif,
			TotalPushStateRecords:  totalPush,
			Recent_1H:              recent1h,
			Recent_24H:             recent24h,
		}), nil
	}
	defer rows.Close()

	var byType []*antclawv1.PushTypeStat
	for rows.Next() {
		var pt string
		var cnt int32
		if rows.Scan(&pt, &cnt) == nil {
			byType = append(byType, &antclawv1.PushTypeStat{PushType: pt, Count: cnt})
		}
	}
	return connect.NewResponse(&antclawv1.GetPushStatsResponse{
		TotalNotifications:     totalNotif,
		TotalPushStateRecords:  totalPush,
		ByType:                 byType,
		Recent_1H:              recent1h,
		Recent_24H:             recent24h,
	}), nil
}

var _ antclawv1connect.SystemServiceHandler = (*SystemHandler)(nil)