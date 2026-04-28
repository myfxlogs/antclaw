package rpc

import (
	"context"

	"connectrpc.com/connect"
	mt5v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// MT5Handler implements antclawv1connect.MT5ServiceHandler.
// MVP: All methods return UNIMPLEMENTED per P6c.
type MT5Handler struct{}

func NewMT5Handler() *MT5Handler {
	return &MT5Handler{}
}

func (h *MT5Handler) PlaceHolder(ctx context.Context, req *connect.Request[mt5v1.MT5PlaceHolderRequest]) (*connect.Response[mt5v1.MT5PlaceHolderResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

var _ antclawv1connect.MT5ServiceHandler = (*MT5Handler)(nil)
