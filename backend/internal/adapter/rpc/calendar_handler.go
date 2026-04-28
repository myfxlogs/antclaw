package rpc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/service/calendar"
	calendarv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// CalendarHandler implements antclawv1connect.CalendarServiceHandler.
type CalendarHandler struct {
	svc *calendar.Service
}

func NewCalendarHandler(svc *calendar.Service) *CalendarHandler {
	return &CalendarHandler{svc: svc}
}

func (h *CalendarHandler) ListEvents(ctx context.Context, req *connect.Request[calendarv1.ListEventsRequest]) (*connect.Response[calendarv1.ListEventsResponse], error) {
	resp, err := h.svc.ListEvents(ctx, req.Msg.Date, req.Msg.CurrencyFilter, req.Msg.MinImpact)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *CalendarHandler) GetEvent(ctx context.Context, req *connect.Request[calendarv1.GetEventRequest]) (*connect.Response[calendarv1.GetEventResponse], error) {
	resp, err := h.svc.GetEvent(ctx, req.Msg.EventId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *CalendarHandler) GetImpact(ctx context.Context, req *connect.Request[calendarv1.GetImpactRequest]) (*connect.Response[calendarv1.GetImpactResponse], error) {
	resp, err := h.svc.GetImpact(ctx, req.Msg.EventId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

func (h *CalendarHandler) GetImpactHistory(ctx context.Context, req *connect.Request[calendarv1.GetImpactHistoryRequest]) (*connect.Response[calendarv1.GetImpactHistoryResponse], error) {
	resp, err := h.svc.GetImpactHistory(ctx, req.Msg.EventType, req.Msg.Pair, req.Msg.Count)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(resp), nil
}

var _ antclawv1connect.CalendarServiceHandler = (*CalendarHandler)(nil)

// RegisterCalendarService registers the Calendar service with the mux.
func RegisterCalendarService(mux *http.ServeMux, handler antclawv1connect.CalendarServiceHandler) {
	mux.Handle(antclawv1connect.NewCalendarServiceHandler(handler))
}
