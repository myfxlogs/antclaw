package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

type CircleHandler struct{ pool *pgxpool.Pool }

func NewCircleHandler(pool *pgxpool.Pool) *CircleHandler { return &CircleHandler{pool: pool} }

func (h *CircleHandler) CreateCircle(ctx context.Context, req *connect.Request[alfqv1.CreateCircleRequest]) (*connect.Response[alfqv1.Circle], error) {
	uid := userIDFromHTTP(ctx, req)
	r := req.Msg
	c := &alfqv1.Circle{Name: r.Name, Description: r.Description, Symbol: r.Symbol, MemberCount: 1, IsMember: true}
	var ts time.Time
	h.pool.QueryRow(ctx, `
		INSERT INTO alfq_circles (name, description, symbol, created_by)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at
	`, r.Name, r.Description, r.Symbol, uid).Scan(&c.Id, &ts)
	c.CreatedAt = ts.Unix()
	h.pool.Exec(ctx, "INSERT INTO alfq_circle_members (circle_id, user_id) VALUES ($1,$2)", c.Id, uid)
	return connect.NewResponse(c), nil
}

func (h *CircleHandler) JoinCircle(ctx context.Context, req *connect.Request[alfqv1.JoinCircleRequest]) (*connect.Response[alfqv1.Circle], error) {
	uid := userIDFromHTTP(ctx, req)
	h.pool.Exec(ctx, "INSERT INTO alfq_circle_members (circle_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", req.Msg.CircleId, uid)
	h.pool.Exec(ctx, "UPDATE alfq_circles SET member_count = (SELECT COUNT(*) FROM alfq_circle_members WHERE circle_id=$1) WHERE id=$1", req.Msg.CircleId)
	return h.getCircle(ctx, req.Msg.CircleId, true)
}

func (h *CircleHandler) LeaveCircle(ctx context.Context, req *connect.Request[alfqv1.LeaveCircleRequest]) (*connect.Response[alfqv1.LeaveCircleResponse], error) {
	uid := userIDFromHTTP(ctx, req)
	h.pool.Exec(ctx, "DELETE FROM alfq_circle_members WHERE circle_id=$1 AND user_id=$2", req.Msg.CircleId, uid)
	return connect.NewResponse(&alfqv1.LeaveCircleResponse{Success: true}), nil
}

func (h *CircleHandler) GetCircleFeed(ctx context.Context, req *connect.Request[alfqv1.GetCircleFeedRequest]) (*connect.Response[alfqv1.CircleFeedResponse], error) {
	// For now, reuse posts with circle_id filter
	return connect.NewResponse(&alfqv1.CircleFeedResponse{}), nil
}

func (h *CircleHandler) ListCircles(ctx context.Context, req *connect.Request[alfqv1.ListCirclesRequest]) (*connect.Response[alfqv1.CircleList], error) {
	rows, _ := h.pool.Query(ctx, `
		SELECT id, name, description, symbol, member_count, created_at FROM alfq_circles ORDER BY member_count DESC LIMIT 20
	`)
	defer rows.Close()
	var circles []*alfqv1.Circle
	for rows.Next() {
		c := &alfqv1.Circle{}
		var ts time.Time
		rows.Scan(&c.Id, &c.Name, &c.Description, &c.Symbol, &c.MemberCount, &ts)
		c.CreatedAt = ts.Unix()
		circles = append(circles, c)
	}
	return connect.NewResponse(&alfqv1.CircleList{Circles: circles}), nil
}

func (h *CircleHandler) getCircle(ctx context.Context, circleID string, isMember bool) (*connect.Response[alfqv1.Circle], error) {
	c := &alfqv1.Circle{IsMember: isMember}
	var ts time.Time
	h.pool.QueryRow(ctx, `
		SELECT id, name, description, symbol, member_count, created_at FROM alfq_circles WHERE id=$1
	`, circleID).Scan(&c.Id, &c.Name, &c.Description, &c.Symbol, &c.MemberCount, &ts)
	c.CreatedAt = ts.Unix()
	return connect.NewResponse(c), nil
}

var _ antclawv1connect.CircleServiceHandler = (*CircleHandler)(nil)
