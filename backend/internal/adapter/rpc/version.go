// Package rpc provides adapter-layer version variables injectable via ldflags.
package rpc

// Build-time variables injected via -ldflags.
// Example: go build -ldflags "-X github.com/antclaw/antclaw/internal/adapter/rpc.Version=1.0.0 -X ..."
var (
	// Version is the API server version. Default "0.1.0".
	Version = "0.1.0"
	// GitCommit is the git SHA at build time. Default "dev".
	GitCommit = "dev"
	// ProtoVersion is the protocol version. Default "1.0.0".
	ProtoVersion = "1.0.0"
	// MinClientVersion is the minimum compatible client version.
	MinClientVersion = "1.0.0"
)
