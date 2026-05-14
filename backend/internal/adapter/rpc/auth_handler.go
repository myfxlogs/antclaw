package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/service/audit"
	authv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/redis/go-redis/v9"
)

// AuthHandler implements antclawv1connect.AuthServiceHandler.
//
// 持久层依赖通过 UserStore 抽象注入，便于在 sqlc 生成完成前用内存实现/mock 通过契约测试。
type AuthHandler struct {
	users UserStore
	rdb   *redis.Client
	audit *audit.AuditService
	pg    *pgxpool.Pool
}

// UserStore is the persistence port required by AuthHandler.
// Real implementation will be wired to sqlc-generated *db.Queries.
//
// 返回值新增 codeID（数字 ID）：登录可凭 email/code_id/username 任一身份。
type UserStore interface {
	CreateUser(ctx context.Context, email, displayName, passwordHash, locale, timezone string) (userID, role, codeID string, passwordVersion int, err error)
	GetUserByEmail(ctx context.Context, email string) (userID, passwordHash, role, locale, status, codeID string, passwordVersion int, err error)
	GetUserByCodeID(ctx context.Context, codeID string) (userID, passwordHash, role, locale, status, outCodeID string, passwordVersion int, err error)
	GetUserByID(ctx context.Context, userID string) (role, locale, codeID string, passwordVersion int, err error)
	CreateSession(ctx context.Context, userID, userAgent, ip string) (sessionID string, err error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeUserSessions(ctx context.Context, userID string) error
	IsSessionRevoked(ctx context.Context, sessionID string) (bool, error)
	StoreRefreshToken(ctx context.Context, jti, sessionID, userID string, expiresAt time.Time) error
	IsRefreshTokenRevoked(ctx context.Context, jti string) (bool, error)
	RevokeRefreshToken(ctx context.Context, jti string) error
	MarkRefreshTokenRotated(ctx context.Context, oldJTI, newJTI string) error
	RevokeUserRefreshTokens(ctx context.Context, userID string) error
}

func NewAuthHandler(users UserStore, rdb *redis.Client, auditSvc *audit.AuditService, pg *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{users: users, rdb: rdb, audit: auditSvc, pg: pg}
}

func (h *AuthHandler) Register(ctx context.Context, req *connect.Request[authv1.RegisterRequest]) (*connect.Response[authv1.RegisterResponse], error) {
	email := normalizeEmail(req.Msg.Email)
	if email == "" || !strings.Contains(email, "@") {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid email"))
	}
	if len(req.Msg.Password) < 10 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password too short"))
	}

	hash, err := auth.HashPassword(req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	locale := req.Msg.Locale.String()
	if locale == "" || locale == "LOCALE_UNSPECIFIED" {
		locale = "zh-CN"
	}
	tz := req.Msg.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	userID, role, codeID, pv, err := h.users.CreateUser(ctx, email, req.Msg.DisplayName, hash, locale, tz)
	if err != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email taken"))
	}

	access, refresh, expiresAt, err := h.issueTokens(ctx, userID, role, locale, pv, clientUA(req.Msg.Client), clientIP(req.Msg.Client))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	h.logAudit(ctx, &userID, audit.ActionRegister, "users", "user registered", req.Msg.Client)
	h.upsertDevice(ctx, userID, req.Msg.Client)

	return connect.NewResponse(&authv1.RegisterResponse{
		UserId:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
		CodeId:       codeID,
	}), nil
}

func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[authv1.LoginRequest]) (*connect.Response[authv1.LoginResponse], error) {
	start := time.Now()
	defer func() {
		// 恒定 200ms 延迟以防止 timing 攻击
		if d := time.Since(start); d < 200*time.Millisecond {
			time.Sleep(200*time.Millisecond - d)
		}
	}()

	// 智能识别 identifier：含 @ 视为邮箱；否则若全数字按 code_id 查询。
	identifier := strings.TrimSpace(req.Msg.Email)
	rateKey := normalizeEmail(identifier)
	if backoff, _ := auth.LoginFailureBackoff(ctx, h.rdb, rateKey); backoff > 0 {
		return nil, connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("rate limited"))
	}

	var (
		userID, hash, role, locale, status, codeID string
		pv                                         int
		err                                        error
	)
	switch {
	case strings.Contains(identifier, "@"):
		userID, hash, role, locale, status, codeID, pv, err = h.users.GetUserByEmail(ctx, normalizeEmail(identifier))
	case auth.IsAllDigits(identifier):
		userID, hash, role, locale, status, codeID, pv, err = h.users.GetUserByCodeID(ctx, identifier)
	default:
		// username 或其它形式暂未支持登录路径，统一走 GetUserByEmail（命中 username UNIQUE 失败也会 401）。
		userID, hash, role, locale, status, codeID, pv, err = h.users.GetUserByEmail(ctx, normalizeEmail(identifier))
	}
	if err != nil {
		_ = auth.RecordLoginFailure(ctx, h.rdb, rateKey)
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid credentials"))
	}
	if status == "banned" {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("account banned"))
	}

	ok, err := auth.VerifyPassword(req.Msg.Password, hash)
	if err != nil || !ok {
		_ = auth.RecordLoginFailure(ctx, h.rdb, rateKey)
		h.logAudit(ctx, &userID, audit.ActionLoginFailed, "users", "invalid credentials", req.Msg.Client)
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid credentials"))
	}

	_ = auth.ClearLoginFailures(ctx, h.rdb, rateKey)

	access, refresh, expiresAt, err := h.issueTokens(ctx, userID, role, locale, pv, clientUA(req.Msg.Client), clientIP(req.Msg.Client))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	h.logAudit(ctx, &userID, audit.ActionLogin, "users", "login success", req.Msg.Client)
	h.upsertDevice(ctx, userID, req.Msg.Client)

	return connect.NewResponse(&authv1.LoginResponse{
		UserId:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    expiresAt,
		CodeId:       codeID,
	}), nil
}

func (h *AuthHandler) Refresh(ctx context.Context, req *connect.Request[authv1.RefreshRequest]) (*connect.Response[authv1.RefreshResponse], error) {
	claims, err := auth.ValidateToken(req.Msg.RefreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if revoked, _ := auth.IsRefreshTokenRevoked(ctx, h.rdb, claims.JTI); revoked {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("refresh revoked"))
	}
	if revoked, _ := h.users.IsRefreshTokenRevoked(ctx, claims.JTI); revoked {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("refresh revoked"))
	}

	// 复用检测
	reused, err := auth.CheckAndRecordRefreshReuse(ctx, h.rdb, claims.JTI, "")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if reused {
		_ = h.users.RevokeUserSessions(ctx, claims.Subject)
		h.logAudit(ctx, &claims.Subject, audit.ActionSessionRevoked, "sessions", "refresh token reuse detected", nil)
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("refresh reused"))
	}

	role, locale, _, pv, err := h.users.GetUserByID(ctx, claims.Subject)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("user not found"))
	}
	if pv != claims.PasswordVersion {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("password changed"))
	}
	if revoked, _ := h.users.IsSessionRevoked(ctx, claims.SessionID); revoked {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("session revoked"))
	}

	// 立即吊销旧 jti
	_ = h.users.RevokeRefreshToken(ctx, claims.JTI)
	_ = auth.RevokeRefreshToken(ctx, h.rdb, claims.JTI, time.Until(time.Unix(claims.ExpiresAt, 0)))

	access, _, err := auth.GenerateAccessToken(claims.Subject, claims.SessionID, role, locale, pv)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	refresh, newJTI, refreshExp, err := auth.GenerateRefreshToken(claims.Subject, claims.SessionID, pv)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := h.users.StoreRefreshToken(ctx, newJTI, claims.SessionID, claims.Subject, refreshExp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	_ = h.users.MarkRefreshTokenRotated(ctx, claims.JTI, newJTI)

	h.logAudit(ctx, &claims.Subject, audit.ActionRefresh, "tokens", "rotated", nil)

	return connect.NewResponse(&authv1.RefreshResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    time.Now().Add(auth.AccessTTL).Unix(),
	}), nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	claims, err := auth.ValidateToken(req.Msg.RefreshToken, auth.TokenTypeRefresh)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if req.Msg.AllDevices {
		_ = h.users.RevokeUserSessions(ctx, claims.Subject)
		_ = h.users.RevokeUserRefreshTokens(ctx, claims.Subject)
	} else {
		_ = h.users.RevokeSession(ctx, claims.SessionID)
		_ = h.users.RevokeRefreshToken(ctx, claims.JTI)
		_ = auth.RevokeRefreshToken(ctx, h.rdb, claims.JTI, time.Until(time.Unix(claims.ExpiresAt, 0)))
	}

	h.logAudit(ctx, &claims.Subject, audit.ActionLogout, "sessions", fmt.Sprintf("all_devices=%v", req.Msg.AllDevices), nil)
	return connect.NewResponse(&authv1.LogoutResponse{}), nil
}

func (h *AuthHandler) RequestPasswordReset(ctx context.Context, req *connect.Request[authv1.RequestPasswordResetRequest]) (*connect.Response[authv1.RequestPasswordResetResponse], error) {
	// 恒定响应（防止枚举），实际 token 签发在 P4b 邮件链路
	return connect.NewResponse(&authv1.RequestPasswordResetResponse{Sent: true}), nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *connect.Request[authv1.ResetPasswordRequest]) (*connect.Response[authv1.ResetPasswordResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *connect.Request[authv1.VerifyEmailRequest]) (*connect.Response[authv1.VerifyEmailResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("not implemented"))
}

// upsertDevice 在登录/注册成功时将 device_id 写入 devices 表。
// 后续客户端可通过 ReportDeviceInfo RPC 补全 model/os_version 等详细信息。
func (h *AuthHandler) upsertDevice(ctx context.Context, userID string, client *authv1.ClientInfo) {
	if h.pg == nil || client == nil || client.DeviceId == "" {
		return
	}
	_, _ = h.pg.Exec(ctx, `
		INSERT INTO devices (device_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (device_id) DO UPDATE SET user_id=$2, updated_at=NOW()`,
		client.DeviceId, userID,
	)
}

// helpers

func (h *AuthHandler) issueTokens(ctx context.Context, userID, role, locale string, pv int, ua, ip string) (access, refresh string, expiresAt int64, err error) {
	sessionID, err := h.users.CreateSession(ctx, userID, ua, ip)
	if err != nil {
		return "", "", 0, err
	}
	access, _, err = auth.GenerateAccessToken(userID, sessionID, role, locale, pv)
	if err != nil {
		return "", "", 0, err
	}
	refresh, jti, refreshExp, err := auth.GenerateRefreshToken(userID, sessionID, pv)
	if err != nil {
		return "", "", 0, err
	}
	if err := h.users.StoreRefreshToken(ctx, jti, sessionID, userID, refreshExp); err != nil {
		return "", "", 0, err
	}
	return access, refresh, time.Now().Add(auth.AccessTTL).Unix(), nil
}

func (h *AuthHandler) logAudit(ctx context.Context, userID *string, action, resource, details string, client *authv1.ClientInfo) {
	if h.audit == nil {
		return
	}
	entry := audit.AuditEntry{
		Action:   action,
		Resource: resource,
		Details:  details,
	}
	if client != nil {
		entry.IPAddress = client.IpAddress
		entry.UserAgent = client.UserAgent
	}
	// 注：UserID 暂以字符串透传，sqlc 生成接入后再转 uuid
	_ = userID
	_, _ = h.audit.Log(ctx, entry)
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func clientUA(c *authv1.ClientInfo) string {
	if c == nil {
		return ""
	}
	return c.UserAgent
}

func clientIP(c *authv1.ClientInfo) string {
	if c == nil {
		return ""
	}
	return c.IpAddress
}

var _ antclawv1connect.AuthServiceHandler = (*AuthHandler)(nil)
