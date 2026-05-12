package rpc

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/auth"
)

// userIDFromHTTP extracts authenticated userID.
// Priority: context (if auth interceptor ran) → Authorization header JWT.
func userIDFromHTTP(ctx context.Context, req connect.AnyRequest) string {
	if uid, ok := auth.UserIDFromContext(ctx); ok && uid != "" {
		return uid
	}
	header := req.Header().Get("Authorization")
	if header == "" {
		return ""
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return ""
	}
	claims, err := auth.ParseToken(token)
	if err != nil {
		return ""
	}
	return claims.Subject
}
