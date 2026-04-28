// Package sandbox provides strategy sandbox engine (reserved for later Starlark implementation).
// MVP: directory exists but implementations return ErrNotImplemented.
// See: AntClaw-未决项清单.md §十二.8
package sandbox

import "errors"

// ErrNotImplemented indicates the sandbox functionality is not yet implemented.
var ErrNotImplemented = errors.New("sandbox functionality not implemented")
