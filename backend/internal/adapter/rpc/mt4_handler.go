package rpc

import (
	"context"

	"connectrpc.com/connect"
	mt4v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// MT4Handler implements antclawv1connect.MT4ServiceHandler.
type MT4Handler struct{}

func NewMT4Handler() *MT4Handler {
	return &MT4Handler{}
}

func (h *MT4Handler) AddAccount(ctx context.Context, req *connect.Request[mt4v1.AddMT4AccountRequest]) (*connect.Response[mt4v1.MT4Account], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT4Handler) RemoveAccount(ctx context.Context, req *connect.Request[mt4v1.RemoveMT4AccountRequest]) (*connect.Response[mt4v1.RemoveMT4AccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT4Handler) GetAccountInfo(ctx context.Context, req *connect.Request[mt4v1.GetMT4AccountInfoRequest]) (*connect.Response[mt4v1.MT4AccountInfo], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT4Handler) GetPositions(ctx context.Context, req *connect.Request[mt4v1.GetMT4PositionsRequest]) (*connect.Response[mt4v1.MT4PositionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT4Handler) GetHistory(ctx context.Context, req *connect.Request[mt4v1.GetMT4HistoryRequest]) (*connect.Response[mt4v1.MT4HistoryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

var _ antclawv1connect.MT4ServiceHandler = (*MT4Handler)(nil)
