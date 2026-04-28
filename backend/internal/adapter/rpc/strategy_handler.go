package rpc

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	strategyv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/service/strategy"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

// StrategyHandler implements antclawv1connect.StrategyServiceHandler.
type StrategyHandler struct {
	svc *strategy.Service
}

func NewStrategyHandler(svc *strategy.Service) *StrategyHandler {
	return &StrategyHandler{svc: svc}
}

func strategyToProto(s strategy.Strategy) *strategyv1.Strategy {
	params, _ := structpb.NewStruct(s.Params)
	out := &strategyv1.Strategy{
		Id:            s.ID.String(),
		Name:          s.Name,
		Kind:          s.Kind,
		Symbol:        s.Symbol,
		Timeframe:     s.Timeframe,
		Params:        params,
		ScheduleCron:  s.ScheduleCron,
		Enabled:       s.Enabled,
		Status:        s.Status,
		Description:   s.Description,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
		CreatedBy:     s.CreatedBy,
		UpdatedBy:     s.UpdatedBy,
		LastRunStatus: s.LastRunStatus,
	}
	if s.LastRunAt != nil {
		out.LastRunAt = s.LastRunAt.Format(time.RFC3339)
	}
	return out
}

func runToProto(r strategy.RunResult) *strategyv1.StrategyRun {
	metrics, _ := structpb.NewStruct(r.Metrics)
	return &strategyv1.StrategyRun{
		RunId:        r.RunID.String(),
		StrategyId:   r.StrategyID.String(),
		StartedAt:    r.StartedAt.Format(time.RFC3339),
		FinishedAt:   r.FinishedAt.Format(time.RFC3339),
		Status:       r.Status,
		Metrics:      metrics,
		Mock:         r.Mock,
		ErrorMessage: r.ErrorMessage,
	}
}

func parseStrategyID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid id: %w", err))
	}
	return parsed, nil
}

func currentUser(_ context.Context) string {
	return "admin-001"
}

func (h *StrategyHandler) ListStrategies(ctx context.Context, req *connect.Request[strategyv1.ListStrategiesRequest]) (*connect.Response[strategyv1.ListStrategiesResponse], error) {
	items, total, err := h.svc.List(ctx, int(req.Msg.Offset), int(req.Msg.Limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*strategyv1.Strategy, 0, len(items))
	for _, it := range items {
		out = append(out, strategyToProto(it))
	}
	return connect.NewResponse(&strategyv1.ListStrategiesResponse{
		Items: out,
		Total: int32(total),
	}), nil
}

func (h *StrategyHandler) GetStrategy(ctx context.Context, req *connect.Request[strategyv1.GetStrategyRequest]) (*connect.Response[strategyv1.GetStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	st, err := h.svc.Get(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&strategyv1.GetStrategyResponse{Item: strategyToProto(*st)}), nil
}

func (h *StrategyHandler) CreateStrategy(ctx context.Context, req *connect.Request[strategyv1.CreateStrategyRequest]) (*connect.Response[strategyv1.CreateStrategyResponse], error) {
	params := map[string]any{}
	if req.Msg.Params != nil {
		params = req.Msg.Params.AsMap()
	}
	st := &strategy.Strategy{
		Name:         req.Msg.Name,
		Kind:         req.Msg.Kind,
		Symbol:       req.Msg.Symbol,
		Timeframe:    req.Msg.Timeframe,
		Params:       params,
		ScheduleCron: req.Msg.ScheduleCron,
		Description:  req.Msg.Description,
		Status:       "draft",
	}
	if st.ScheduleCron == "" {
		st.ScheduleCron = "@hourly"
	}
	if err := h.svc.Create(ctx, st, currentUser(ctx)); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&strategyv1.CreateStrategyResponse{Item: strategyToProto(*st)}), nil
}

func (h *StrategyHandler) UpdateStrategy(ctx context.Context, req *connect.Request[strategyv1.UpdateStrategyRequest]) (*connect.Response[strategyv1.UpdateStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	params := map[string]any{}
	if req.Msg.Params != nil {
		params = req.Msg.Params.AsMap()
	}
	st := &strategy.Strategy{
		Name:         req.Msg.Name,
		Kind:         req.Msg.Kind,
		Symbol:       req.Msg.Symbol,
		Timeframe:    req.Msg.Timeframe,
		Params:       params,
		ScheduleCron: req.Msg.ScheduleCron,
		Enabled:      req.Msg.Enabled,
		Status:       req.Msg.Status,
		Description:  req.Msg.Description,
	}
	if err := h.svc.Update(ctx, id, st, currentUser(ctx)); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&strategyv1.UpdateStrategyResponse{Id: id.String()}), nil
}

func (h *StrategyHandler) DeleteStrategy(ctx context.Context, req *connect.Request[strategyv1.DeleteStrategyRequest]) (*connect.Response[strategyv1.DeleteStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	if err := h.svc.SoftDelete(ctx, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&strategyv1.DeleteStrategyResponse{}), nil
}

func (h *StrategyHandler) EnableStrategy(ctx context.Context, req *connect.Request[strategyv1.EnableStrategyRequest]) (*connect.Response[strategyv1.EnableStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	if err := h.svc.Enable(ctx, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&strategyv1.EnableStrategyResponse{Id: id.String(), Enabled: true}), nil
}

func (h *StrategyHandler) DisableStrategy(ctx context.Context, req *connect.Request[strategyv1.DisableStrategyRequest]) (*connect.Response[strategyv1.DisableStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	if err := h.svc.Disable(ctx, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&strategyv1.DisableStrategyResponse{Id: id.String(), Enabled: false}), nil
}

func (h *StrategyHandler) RunStrategy(ctx context.Context, req *connect.Request[strategyv1.RunStrategyRequest]) (*connect.Response[strategyv1.RunStrategyResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	result, err := h.svc.Run(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&strategyv1.RunStrategyResponse{Item: runToProto(*result)}), nil
}

func (h *StrategyHandler) ListStrategyRuns(ctx context.Context, req *connect.Request[strategyv1.ListStrategyRunsRequest]) (*connect.Response[strategyv1.ListStrategyRunsResponse], error) {
	id, err := parseStrategyID(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	runs, err := h.svc.ListRuns(ctx, id, int(req.Msg.Limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	items := make([]*strategyv1.StrategyRun, 0, len(runs))
	for _, run := range runs {
		items = append(items, runToProto(run))
	}
	return connect.NewResponse(&strategyv1.ListStrategyRunsResponse{Items: items}), nil
}

var _ antclawv1connect.StrategyServiceHandler = (*StrategyHandler)(nil)
