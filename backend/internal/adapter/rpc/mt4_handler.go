package rpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	mt4v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"

	"github.com/antclaw/antclaw/internal/domain"
	mtsvc "github.com/antclaw/antclaw/internal/service/mt"
)

// MT4Handler implements antclawv1connect.MT4ServiceHandler.
// Delegates business logic to the mt.Service.
type MT4Handler struct {
	svc *mtsvc.Service
}

// NewMT4Handler creates a new MT4Handler with the given MT service.
func NewMT4Handler(svc *mtsvc.Service) *MT4Handler {
	return &MT4Handler{svc: svc}
}

// AddAccount binds a new MT4 trading account.
func (h *MT4Handler) AddAccount(ctx context.Context, req *connect.Request[mt4v1.AddMT4AccountRequest]) (*connect.Response[mt4v1.MT4Account], error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Map MT4 proto request to Service request
	svcReq := mtsvc.CreateAccountRequest{
		Login:        req.Msg.Account,
		Password:     req.Msg.InvestorPassword,
		MTType:       domain.MTType4,
		BrokerServer: req.Msg.Server,
		Alias:        req.Msg.Label,
	}

	account, err := h.svc.CreateAccount(ctx, userID, svcReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create mt4 account: %w", err))
	}

	return connect.NewResponse(convertToMT4Account(account)), nil
}

// RemoveAccount deletes an MT4 account binding.
func (h *MT4Handler) RemoveAccount(ctx context.Context, req *connect.Request[mt4v1.RemoveMT4AccountRequest]) (*connect.Response[mt4v1.RemoveMT4AccountResponse], error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := h.svc.DeleteAccount(ctx, userID, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete mt4 account: %w", err))
	}

	return connect.NewResponse(&mt4v1.RemoveMT4AccountResponse{Success: true}), nil
}

// GetAccountInfo returns the current MT4 account summary.
func (h *MT4Handler) GetAccountInfo(ctx context.Context, req *connect.Request[mt4v1.GetMT4AccountInfoRequest]) (*connect.Response[mt4v1.MT4AccountInfo], error) {
	userID, err := userID(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	account, err := h.svc.GetAccountInfo(ctx, userID, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get mt4 account info: %w", err))
	}

	return connect.NewResponse(convertToMT4AccountInfo(account)), nil
}

// GetPositions returns the current MT4 positions.
func (h *MT4Handler) GetPositions(ctx context.Context, req *connect.Request[mt4v1.GetMT4PositionsRequest]) (*connect.Response[mt4v1.MT4PositionsResponse], error) {
	// Positions require a connected MT gateway session.
	// This is a future enhancement; return empty for now.
	return connect.NewResponse(&mt4v1.MT4PositionsResponse{}), nil
}

// GetHistory returns historical MT4 orders.
func (h *MT4Handler) GetHistory(ctx context.Context, req *connect.Request[mt4v1.GetMT4HistoryRequest]) (*connect.Response[mt4v1.MT4HistoryResponse], error) {
	// History requires a connected MT gateway session.
	// This is a future enhancement; return empty for now.
	return connect.NewResponse(&mt4v1.MT4HistoryResponse{}), nil
}

// ---- Proto conversion helpers ----

func convertToMT4Account(a *domain.MTAccount) *mt4v1.MT4Account {
	connected := a.Status == domain.MTAccountStatusConnected
	return &mt4v1.MT4Account{
		Id:        a.ID,
		Server:    a.BrokerServer,
		Account:   a.Login,
		Label:     a.Alias,
		IsDemo:    a.AccountType == "demo",
		Connected: connected,
		CreatedAt: a.CreatedAt.Unix(),
	}
}

func convertToMT4AccountInfo(a *domain.MTAccount) *mt4v1.MT4AccountInfo {
	return &mt4v1.MT4AccountInfo{
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
