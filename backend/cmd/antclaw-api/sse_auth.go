package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/antclaw/antclaw/internal/auth"
)

// extractUserIDFromRequest 从 Authorization: Bearer 或 Cookie antclaw_at= 中提取 access_token，
// 校验后返回 user_id 和 role。SSE 在浏览器端依赖 cookie（EventSource 无法自定义请求头）。
func extractUserIDFromRequest(r *http.Request) (userID, role string, err error) {
	token := ""
	if a := r.Header.Get("Authorization"); a != "" {
		parts := strings.SplitN(a, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = parts[1]
		}
	}
	if token == "" {
		if c, err := r.Cookie("antclaw_at"); err == nil {
			token = c.Value
		}
	}
	// 兼容 EventSource 无 Authorization 时的查询参数兜底（仅在前端确实需要时使用）。
	if token == "" {
		token = r.URL.Query().Get("access_token")
	}
	if token == "" {
		return "", "", errors.New("missing token")
	}
	claims, err := auth.ValidateToken(token, auth.TokenTypeAccess)
	if err != nil {
		return "", "", err
	}
	if claims.Subject == "" {
		return "", "", errors.New("invalid claims")
	}
	return claims.Subject, claims.Role, nil
}
