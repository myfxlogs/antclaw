package rpc

import (
	"context"

	"connectrpc.com/connect"
	mt5v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// MT5Handler implements antclawv1connect.MT5ServiceHandler.
// M0.5: 只读接入（账号绑定/余额/持仓/历史订单）。
// 实际 MT5 连接代码由外部成熟模块提供，此处为接口骨架。
type MT5Handler struct{}

func NewMT5Handler() *MT5Handler {
	return &MT5Handler{}
}

func (h *MT5Handler) AddAccount(ctx context.Context, req *connect.Request[mt5v1.AddMT5AccountRequest]) (*connect.Response[mt5v1.MT5Account], error) {
	// TODO: 调用外部 MT5 连接模块验证 server/account/investor_password
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT5Handler) RemoveAccount(ctx context.Context, req *connect.Request[mt5v1.RemoveMT5AccountRequest]) (*connect.Response[mt5v1.RemoveMT5AccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT5Handler) GetAccountInfo(ctx context.Context, req *connect.Request[mt5v1.GetMT5AccountInfoRequest]) (*connect.Response[mt5v1.MT5AccountInfo], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT5Handler) GetPositions(ctx context.Context, req *connect.Request[mt5v1.GetMT5PositionsRequest]) (*connect.Response[mt5v1.MT5PositionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *MT5Handler) GetHistory(ctx context.Context, req *connect.Request[mt5v1.GetMT5HistoryRequest]) (*connect.Response[mt5v1.MT5HistoryResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

var _ antclawv1connect.MT5ServiceHandler = (*MT5Handler)(nil)
