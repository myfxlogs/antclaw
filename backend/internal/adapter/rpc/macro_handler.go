package rpc

import (
"context"

"connectrpc.com/connect"
"github.com/antclaw/antclaw/internal/service/macro"
macrov1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// MacroHandler implements antclawv1connect.MacroServiceHandler.
type MacroHandler struct {
	svc *macro.Service
}

// NewMacroHandler creates a new MacroHandler.
func NewMacroHandler(svc *macro.Service) *MacroHandler {
	return &MacroHandler{svc: svc}
}

func (h *MacroHandler) GetFred(ctx context.Context, req *connect.Request[macrov1.GetFredRequest]) (*connect.Response[macrov1.GetFredResponse], error) {
	resp, err := h.svc.GetFred(ctx, req.Msg.SeriesId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetEcb(ctx context.Context, req *connect.Request[macrov1.GetEcbRequest]) (*connect.Response[macrov1.GetEcbResponse], error) {
	resp, err := h.svc.GetEcb(ctx, req.Msg.SeriesKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetSnb(ctx context.Context, req *connect.Request[macrov1.GetSnbRequest]) (*connect.Response[macrov1.GetSnbResponse], error) {
	resp, err := h.svc.GetSnb(ctx, req.Msg.Indicator)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetOecdLeading(ctx context.Context, req *connect.Request[macrov1.GetOecdLeadingRequest]) (*connect.Response[macrov1.GetOecdLeadingResponse], error) {
	resp, err := h.svc.GetOecdLeading(ctx, req.Msg.Country)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetEurostat(ctx context.Context, req *connect.Request[macrov1.GetEurostatRequest]) (*connect.Response[macrov1.GetEurostatResponse], error) {
	resp, err := h.svc.GetEurostat(ctx, req.Msg.Dataset, req.Msg.Geo)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetBis(ctx context.Context, req *connect.Request[macrov1.GetBisRequest]) (*connect.Response[macrov1.GetBisResponse], error) {
	resp, err := h.svc.GetBis(ctx, req.Msg.Dataset, req.Msg.Jurisdiction)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetTradingEconomics(ctx context.Context, req *connect.Request[macrov1.GetTradingEconomicsRequest]) (*connect.Response[macrov1.GetTradingEconomicsResponse], error) {
	resp, err := h.svc.GetTradingEconomics(ctx, req.Msg.Country, req.Msg.Category)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetDtccSwaps(ctx context.Context, req *connect.Request[macrov1.GetDtccSwapsRequest]) (*connect.Response[macrov1.GetDtccSwapsResponse], error) {
	resp, err := h.svc.GetDtccSwaps(ctx, req.Msg.Pair, req.Msg.Tenor)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetSec13F(ctx context.Context, req *connect.Request[macrov1.GetSec13FRequest]) (*connect.Response[macrov1.GetSec13FResponse], error) {
	// Parse quarter from "YYYY-QN" format, default to 0 for latest
	quarterNum := int64(0)
	if len(req.Msg.Quarter) > 1 {
		// Simple parsing: extract last digit if present
		lastChar := req.Msg.Quarter[len(req.Msg.Quarter)-1:]
		switch lastChar {
		case "1":
			quarterNum = 1
		case "2":
			quarterNum = 2
		case "3":
			quarterNum = 3
		case "4":
			quarterNum = 4
		}
	}
	resp, err := h.svc.GetSec13F(ctx, req.Msg.Cik, quarterNum)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetTreasuryAuctions(ctx context.Context, req *connect.Request[macrov1.GetTreasuryAuctionsRequest]) (*connect.Response[macrov1.GetTreasuryAuctionsResponse], error) {
	resp, err := h.svc.GetTreasuryAuctions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetFedWatch(ctx context.Context, req *connect.Request[macrov1.GetFedWatchRequest]) (*connect.Response[macrov1.GetFedWatchResponse], error) {
	resp, err := h.svc.GetFedWatch(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetWorldBank(ctx context.Context, req *connect.Request[macrov1.GetWorldBankRequest]) (*connect.Response[macrov1.GetWorldBankResponse], error) {
	resp, err := h.svc.GetWorldBank(ctx, req.Msg.Indicator, req.Msg.Country)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *MacroHandler) GetImfWeo(ctx context.Context, req *connect.Request[macrov1.GetImfWeoRequest]) (*connect.Response[macrov1.GetImfWeoResponse], error) {
	resp, err := h.svc.GetImfWeo(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.MacroServiceHandler = (*MacroHandler)(nil)
