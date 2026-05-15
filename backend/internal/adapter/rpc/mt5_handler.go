package rpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	mt5v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"

	"github.com/antclaw/antclaw/internal/domain"
	mtsvc "github.com/antclaw/antclaw/internal/service/mt"
)

// MT5Handler implements antclawv1connect.MT5ServiceHandler.
// Delegates business logic to the mt.Service.
type MT5Handler struct {
	svc *mtsvc.Service
}

// NewMT5Handler creates a new MT5Handler with the given MT service.
func NewMT5Handler(svc *mtsvc.Service) *MT5Handler {
	return &MT5Handler{svc: svc}
}

// AddAccount binds a new MT5 trading account.
func (h *MT5Handler) AddAccount(ctx context.Context, req *connect.Request[mt5v1.AddMT5AccountRequest]) (*connect.Response[mt5v1.MT5Account], error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	svcReq := mtsvc.CreateAccountRequest{
		Login:        req.Msg.Account,
		Password:     req.Msg.InvestorPassword,
		MTType:       domain.MTType5,
		BrokerServer: req.Msg.Server,
		Alias:        req.Msg.Label,
	}

	account, err := h.svc.CreateAccount(ctx, uid, svcReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create mt5 account: %w", err))
	}

	return connect.NewResponse(convertToMT5Account(account)), nil
}

// RemoveAccount deletes an MT5 account binding.
func (h *MT5Handler) RemoveAccount(ctx context.Context, req *connect.Request[mt5v1.RemoveMT5AccountRequest]) (*connect.Response[mt5v1.RemoveMT5AccountResponse], error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := h.svc.DeleteAccount(ctx, uid, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete mt5 account: %w", err))
	}

	return connect.NewResponse(&mt5v1.RemoveMT5AccountResponse{Success: true}), nil
}

// GetAccountInfo returns the current MT5 account summary.
func (h *MT5Handler) GetAccountInfo(ctx context.Context, req *connect.Request[mt5v1.GetMT5AccountInfoRequest]) (*connect.Response[mt5v1.MT5AccountInfo], error) {
	uid, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	account, err := h.svc.GetAccountInfo(ctx, uid, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get mt5 account info: %w", err))
	}

	return connect.NewResponse(convertToMT5AccountInfo(account)), nil
}

// GetPositions returns the current MT5 positions.
func (h *MT5Handler) GetPositions(ctx context.Context, req *connect.Request[mt5v1.GetMT5PositionsRequest]) (*connect.Response[mt5v1.MT5PositionsResponse], error) {
	// Positions require a connected MT gateway session; future enhancement.
	return connect.NewResponse(&mt5v1.MT5PositionsResponse{}), nil
}

// GetHistory returns historical MT5 orders.
func (h *MT5Handler) GetHistory(ctx context.Context, req *connect.Request[mt5v1.GetMT5HistoryRequest]) (*connect.Response[mt5v1.MT5HistoryResponse], error) {
	// History requires a connected MT gateway session; future enhancement.
	return connect.NewResponse(&mt5v1.MT5HistoryResponse{}), nil
}

// ---- Proto conversion helpers ----

func convertToMT5Account(a *domain.MTAccount) *mt5v1.MT5Account {
	connected := a.Status == domain.MTAccountStatusConnected
	return &mt5v1.MT5Account{
		Id:        a.ID,
		Server:    a.BrokerServer,
		Account:   a.Login,
		Label:     a.Alias,
		IsDemo:    a.AccountType == "demo",
		Connected: connected,
		CreatedAt: a.CreatedAt.Unix(),
	}
}

func convertToMT5AccountInfo(a *domain.MTAccount) *mt5v1.MT5AccountInfo {
	return &mt5v1.MT5AccountInfo{
		Id:            a.ID,
		Balance:       a.Balance,
		Equity:        a.Equity,
		Margin:        a.Margin,
		FreeMargin:    a.FreeMargin,
		MarginLevel:   a.MarginLevel,
		TodayPnl:      a.Profit,
		PositionCount: 0,
		UpdatedAt:     a.UpdatedAt.Unix(),
	}
}
