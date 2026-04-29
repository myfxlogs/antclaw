package rpc

import (
"context"

"connectrpc.com/connect"
"github.com/antclaw/antclaw/internal/service/ai"
aiv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// AIHandler implements antclawv1connect.AIServiceHandler.
type AIHandler struct {
	svc *ai.Service
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(svc *ai.Service) *AIHandler {
	return &AIHandler{svc: svc}
}

// Chat 把每条 ChatRequest 当作一次问答：读取消息 → 调 systemai → 把回答以单 chunk + done=true 推回。
// 当前不做真正的逐 token 流式（依赖具体 provider SSE 协议），但保证客户端可以读到完整回答和 full_message。
func (h *AIHandler) Chat(ctx context.Context, stream *connect.BidiStream[aiv1.ChatRequest, aiv1.ChatResponse]) error {
	for {
		req, err := stream.Receive()
		if err != nil {
			return nil // EOF or client closed
		}
		reply, err := h.svc.ChatOnce(ctx, req)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if err := stream.Send(reply); err != nil {
			return err
		}
	}
}

func (h *AIHandler) Interpret(ctx context.Context, req *connect.Request[aiv1.InterpretRequest]) (*connect.Response[aiv1.InterpretResponse], error) {
	resp, err := h.svc.Interpret(ctx, req.Msg.DataType, req.Msg.RawData, req.Msg.Question, req.Msg.Locale)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AIHandler) Outlook(ctx context.Context, req *connect.Request[aiv1.OutlookRequest]) (*connect.Response[aiv1.OutlookResponse], error) {
	resp, err := h.svc.Outlook(ctx, req.Msg.Pair, req.Msg.Timeframe, req.Msg.Locale)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *AIHandler) BuildContext(ctx context.Context, req *connect.Request[aiv1.BuildContextRequest]) (*connect.Response[aiv1.BuildContextResponse], error) {
	resp, err := h.svc.BuildContext(ctx, req.Msg.Asset, req.Msg.Scope, req.Msg.Locale)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

// =====================================================
// M-F: 记忆 / 限流 / 工具调用
// =====================================================

func (h *AIHandler) RememberFact(ctx context.Context, req *connect.Request[aiv1.RememberFactRequest]) (*connect.Response[aiv1.RememberFactResponse], error) {
	mem := ai.NewMemoryStore(h.svc.Pool())
	id, err := mem.Remember(ctx, req.Msg.UserId, req.Msg.Scope, req.Msg.Key, req.Msg.Value, req.Msg.TtlSeconds)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&aiv1.RememberFactResponse{Id: id}), nil
}

func (h *AIHandler) RecallFact(ctx context.Context, req *connect.Request[aiv1.RecallFactRequest]) (*connect.Response[aiv1.RecallFactResponse], error) {
	mem := ai.NewMemoryStore(h.svc.Pool())
	it, err := mem.Recall(ctx, req.Msg.UserId, req.Msg.Scope, req.Msg.Key)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&aiv1.RecallFactResponse{
		Id: it.ID, Value: it.Value, CreatedAt: it.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}), nil
}

func (h *AIHandler) SearchMemory(ctx context.Context, req *connect.Request[aiv1.SearchMemoryRequest]) (*connect.Response[aiv1.SearchMemoryResponse], error) {
	mem := ai.NewMemoryStore(h.svc.Pool())
	hits, err := mem.Search(ctx, req.Msg.UserId, req.Msg.Query, int(req.Msg.Limit))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &aiv1.SearchMemoryResponse{}
	for _, h := range hits {
		out.Hits = append(out.Hits, &aiv1.MemoryHit{Id: h.ID, Scope: h.Scope, Key: h.Key, Value: h.Value})
	}
	return connect.NewResponse(out), nil
}

func (h *AIHandler) CheckRateLimit(ctx context.Context, req *connect.Request[aiv1.CheckRateLimitRequest]) (*connect.Response[aiv1.CheckRateLimitResponse], error) {
	rl := ai.NewRateLimiter(h.svc.Pool())
	st, err := rl.Check(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&aiv1.CheckRateLimitResponse{
		UsedToday:  int32(st.UsedToday),
		MaxPerDay:  int32(st.MaxPerDay),
		Remaining:  int32(st.Remaining),
		Allowed:    st.Allowed,
	}), nil
}

func (h *AIHandler) RunWithTools(ctx context.Context, req *connect.Request[aiv1.RunWithToolsRequest]) (*connect.Response[aiv1.RunWithToolsResponse], error) {
	rl := ai.NewRateLimiter(h.svc.Pool())
	if err := rl.Acquire(ctx, req.Msg.UserId); err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, err)
	}
	r, err := h.svc.RunWithTools(ctx, req.Msg.UserId, req.Msg.ThreadId, req.Msg.Message, req.Msg.Tools, int(req.Msg.MaxHops),
		ai.RunOptions{Model: req.Msg.Model, ProviderID: req.Msg.ProviderId})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &aiv1.RunWithToolsResponse{
		ThreadId:         r.ThreadID,
		Answer:           r.Answer,
		CacheHit:         r.CacheHit,
		Model:            r.Model,
		ProviderId:       r.ProviderID,
		PromptTokens:     r.PromptTokens,
		CompletionTokens: r.CompletionTokens,
		TotalTokens:      r.TotalTokens,
	}
	for _, c := range r.Calls {
		out.Calls = append(out.Calls, &aiv1.ToolCall{
			Name: c.Name, ArgsJson: c.ArgsJSON, ResultJson: c.ResultJSON, Error: c.Error,
		})
	}
	return connect.NewResponse(out), nil
}

var _ antclawv1connect.AIServiceHandler = (*AIHandler)(nil)
