package rpc

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/domain/apperror"
)

// connectError maps a canonical domain error to a Connect-RPC error.
// Unknown errors default to CodeInternal.
func connectError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, apperror.ErrDataInsufficient):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, apperror.ErrPermissionDenied):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, apperror.ErrProviderNotConfigured):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, apperror.ErrUpstreamUnavailable):
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
