// Package apperror defines canonical domain errors used across services.
// Handlers map these to Connect-RPC codes via connectError().
package apperror

import "errors"

var (
	// ErrDataInsufficient is returned when database has no data for the query.
	ErrDataInsufficient = errors.New("data insufficient")

	// ErrProviderNotConfigured is returned when an upstream API is not configured.
	ErrProviderNotConfigured = errors.New("provider not configured")

	// ErrUpstreamUnavailable is returned when an upstream API call fails.
	ErrUpstreamUnavailable = errors.New("upstream unavailable")

	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrPermissionDenied is returned when the caller lacks required permissions.
	ErrPermissionDenied = errors.New("permission denied")
)
