package rpc

import (
	"context"

	"connectrpc.com/connect"
	mt4v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// MT4Handler implements antclawv1connect.MT4ServiceHandler.
// MVP: All methods return UNIMPLEMENTED per P6c.
type MT4Handler struct{}

func NewMT4Handler() *MT4Handler {
	return &MT4Handler{}
}

func (h *MT4Handler) PlaceHolder(ctx context.Context, req *connect.Request[mt4v1.MT4PlaceHolderRequest]) (*connect.Response[mt4v1.MT4PlaceHolderResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

var _ antclawv1connect.MT4ServiceHandler = (*MT4Handler)(nil)
