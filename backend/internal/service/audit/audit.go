package audit

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	redisv9 "github.com/redis/go-redis/v9"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuditService 写入 append-only 审计日志，含 SHA256 哈希链。
type AuditService struct {
	queries  *db.Queries
	lastHash []byte
	redis    *redis.Client
}

// AuditEntry 单条业务侧审计载荷。UserID 可为空。
type AuditEntry struct {
	UserID    *uuid.UUID
	Action    string
	Resource  string
	Details   string
	IPAddress string
	UserAgent string
}

func NewAuditService(queries *db.Queries, rds *redis.Client) *AuditService {
	return &AuditService{queries: queries, lastHash: []byte{}, redis: rds}
}

// Log 计算 hash_prev || payload 的 SHA256 后写入审计表。
func (s *AuditService) Log(ctx context.Context, entry AuditEntry) (int64, error) {
	payload := buildPayload(entry, time.Now().Unix())
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}

	h := sha256.New()
	h.Write(s.lastHash)
	h.Write(payloadJSON)
	hashSelf := h.Sum(nil)

	id, err := s.queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		UserID:    toPgUUID(entry.UserID),
		Action:    entry.Action,
		Resource:  entry.Resource,
		Details:   entry.Details,
		IpAddress: nullableStr(entry.IPAddress),
		UserAgent: nullableStr(entry.UserAgent),
		HashPrev:  s.lastHash,
		HashSelf:  hashSelf,
	})
	if err != nil {
		return 0, fmt.Errorf("create audit log: %w", err)
	}
	s.lastHash = hashSelf

	// 将审计事件写入 Redis Streams，供 SSE 实时消费
	if s.redis != nil {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		payloadJSON, _ := json.Marshal(payload)
		_ = s.redis.Raw().XAdd(ctx, &redisv9.XAddArgs{
			Stream: "stream:audit_events",
			Values: map[string]interface{}{
				"action":   entry.Action,
				"resource": entry.Resource,
				"data":     string(payloadJSON),
			},
		}).Err()
	}
	return id, nil
}

// VerifyChain 全表回放校验哈希链完整性。
func (s *AuditService) VerifyChain(ctx context.Context) error {
	entries, err := s.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:  1_000_000,
		Offset: 0,
	})
	if err != nil {
		return fmt.Errorf("fetch audit logs: %w", err)
	}

	// ListAuditLogs ORDER BY id DESC; 反向遍历得到时间序
	var prevHash []byte
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		ts := int64(0)
		if e.CreatedAt.Valid {
			ts = e.CreatedAt.Time.Unix()
		}
		payload := buildPayload(AuditEntry{
			UserID:    fromPgUUID(e.UserID),
			Action:    e.Action,
			Resource:  e.Resource,
			Details:   e.Details,
			IPAddress: derefStr(e.IpAddress),
			UserAgent: derefStr(e.UserAgent),
		}, ts)
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("entry %d marshal: %w", e.ID, err)
		}

		if !equalBytes(e.HashPrev, prevHash) {
			return fmt.Errorf("hash chain broken at entry %d: hash_prev mismatch", e.ID)
		}
		h := sha256.New()
		h.Write(e.HashPrev)
		h.Write(payloadJSON)
		if !equalBytes(e.HashSelf, h.Sum(nil)) {
			return fmt.Errorf("hash chain broken at entry %d: hash_self mismatch", e.ID)
		}
		prevHash = e.HashSelf
	}
	return nil
}

type auditPayload struct {
	UserID    string `json:"user_id,omitempty"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Details   string `json:"details"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Timestamp int64  `json:"timestamp"`
}

func buildPayload(e AuditEntry, ts int64) auditPayload {
	p := auditPayload{
		Action:    e.Action,
		Resource:  e.Resource,
		Details:   e.Details,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
		Timestamp: ts,
	}
	if e.UserID != nil {
		p.UserID = e.UserID.String()
	}
	return p
}

func toPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func fromPgUUID(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes)
	return &v
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// 预定义 action 类型
const (
	ActionRegister       = "register"
	ActionLogin          = "login"
	ActionLoginFailed    = "login_failed"
	ActionRefresh        = "refresh"
	ActionLogout         = "logout"
	ActionPasswordReset  = "password_reset"
	ActionEmailVerified  = "email_verified"
	ActionRoleChanged    = "role_changed"
	ActionUserBanned     = "user_banned"
	ActionUserUnbanned   = "user_unbanned"
	ActionSessionRevoked = "session_revoked"
	ActionForceLogout    = "force_logout"
)
