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

func (h *AIHandler) Chat(ctx context.Context, stream *connect.BidiStream[aiv1.ChatRequest, aiv1.ChatResponse]) error {
	// TODO: Implement streaming chat with session management
	return connect.NewError(connect.CodeUnimplemented, nil)
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

var _ antclawv1connect.AIServiceHandler = (*AIHandler)(nil)
