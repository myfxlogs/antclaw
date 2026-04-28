package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyRole     contextKey = "role"
	ContextKeyClaims   contextKey = "claims"
	ContextKeySession  contextKey = "session_id"
)

// AuthInterceptor returns a Connect-RPC interceptor for authentication
func AuthInterceptor(requireAuth bool) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if !requireAuth {
				return next(ctx, req)
			}

			token, err := extractToken(req.Header())
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			claims, err := ValidateToken(token, TokenTypeAccess)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			// Context enriched with user info
			ctx = WithUserID(ctx, claims.Subject)
			ctx = WithRole(ctx, claims.Role)
			ctx = WithSessionID(ctx, claims.SessionID)
			ctx = WithClaims(ctx, claims)

			return next(ctx, req)
		}
	}
}

func extractToken(header http.Header) (string, error) {
	auth := header.Get("Authorization")
	if auth == "" {
		// Try cookie
		cookie := header.Get("Cookie")
		if cookie != "" {
			for _, c := range strings.Split(cookie, ";") {
				c = strings.TrimSpace(c)
				if strings.HasPrefix(c, "antclaw_at=") {
					return strings.TrimPrefix(c, "antclaw_at="), nil
				}
			}
		}
		return "", fmt.Errorf("missing authorization")
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization format")
	}

	return parts[1], nil
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextKeyUserID).(string)
	return v, ok
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ContextKeyRole, role)
}

func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextKeyRole).(string)
	return v, ok
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, ContextKeySession, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextKeySession).(string)
	return v, ok
}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ContextKeyClaims, claims)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	v, ok := ctx.Value(ContextKeyClaims).(*Claims)
	return v, ok
}
