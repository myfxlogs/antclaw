// userstore.go provides PostgreSQL implementation of rpc.UserStore
package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// codeIDAllocAttempts CreateUser 时为新用户分配 code_id 的最大重试次数。
// 5 位空间 28k；当占用率较低时几乎不会冲突，留 8 次重试足够。
const codeIDAllocAttempts = 8

// UserStore implements rpc.UserStore interface using sqlc queries
type UserStore struct {
	queries *db.Queries
}

// NewUserStore creates a new PostgreSQL-backed UserStore
func NewUserStore(queries *db.Queries) *UserStore {
	return &UserStore{queries: queries}
}

// CreateUser implements rpc.UserStore.
// 注册时立即分配 5 位 code_id；唯一冲突自动重试至 codeIDAllocAttempts 次。
func (s *UserStore) CreateUser(ctx context.Context, email, displayName, passwordHash, locale, timezone string) (userID, role, codeID string, passwordVersion int, err error) {
	var username *string
	var displayNamePtr *string
	if displayName != "" {
		displayNamePtr = &displayName
	}

	for attempt := 0; attempt < codeIDAllocAttempts; attempt++ {
		cid, gerr := auth.GenerateCodeID(auth.CodeIDDefaultDigits)
		if gerr != nil {
			return "", "", "", 0, fmt.Errorf("generate code_id: %w", gerr)
		}
		cidPtr := cid

		user, ierr := s.queries.CreateUser(ctx, db.CreateUserParams{
			Email:        email,
			Username:     username,
			DisplayName:  displayNamePtr,
			PasswordHash: passwordHash,
			Locale:       locale,
			Timezone:     timezone,
			CodeID:       &cidPtr,
		})
		if ierr == nil {
			cidOut := ""
			if user.CodeID != nil {
				cidOut = *user.CodeID
			}
			return user.ID.String(), user.Role, cidOut, int(user.PasswordVersion), nil
		}
		// 唯一冲突 → email 重复或 code_id 重复；email 冲突直接返回，code_id 重试。
		if !isUniqueViolationOnCodeID(ierr) {
			return "", "", "", 0, ierr
		}
	}
	return "", "", "", 0, errors.New("failed to allocate unique code_id after retries")
}

// GetUserByEmail implements rpc.UserStore.
func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (userID, passwordHash, role, locale, status, codeID string, passwordVersion int, err error) {
	user, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", "", "", "", "", 0, err
	}
	return user.ID.String(), user.PasswordHash, user.Role, user.Locale, user.Status, derefStr(user.CodeID), int(user.PasswordVersion), nil
}

// GetUserByCodeID 通过 code_id 查询，签名与 GetUserByEmail 一致便于 Login 复用。
func (s *UserStore) GetUserByCodeID(ctx context.Context, codeID string) (userID, passwordHash, role, locale, status, outCodeID string, passwordVersion int, err error) {
	cid := codeID
	user, err := s.queries.GetUserByCodeID(ctx, &cid)
	if err != nil {
		return "", "", "", "", "", "", 0, err
	}
	return user.ID.String(), user.PasswordHash, user.Role, user.Locale, user.Status, derefStr(user.CodeID), int(user.PasswordVersion), nil
}

// GetUserByID implements rpc.UserStore.
func (s *UserStore) GetUserByID(ctx context.Context, userID string) (role, locale, codeID string, passwordVersion int, err error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", "", "", 0, fmt.Errorf("invalid user id: %w", err)
	}
	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return "", "", "", 0, err
	}
	return user.Role, user.Locale, derefStr(user.CodeID), int(user.PasswordVersion), nil
}

// UpdateUserCodeID 由管理员调用，将 user_id 对应的 code_id 改为指定值。
// 调用方须保证 codeID 已通过 auth.ValidateCodeID 校验。
func (s *UserStore) UpdateUserCodeID(ctx context.Context, userID, codeID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	cid := codeID
	return s.queries.UpdateUserCodeID(ctx, db.UpdateUserCodeIDParams{ID: id, CodeID: &cid})
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// isUniqueViolationOnCodeID 检测错误是否是 code_id 唯一索引冲突。
// pgx 报文形如：`ERROR: duplicate key value violates unique constraint "users_code_id_uq"`。
func isUniqueViolationOnCodeID(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "users_code_id_uq")
}

// CreateSession implements rpc.UserStore
func (s *UserStore) CreateSession(ctx context.Context, userID, userAgent, ip string) (sessionID string, err error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}

	var ipAddr *netip.Addr
	if ip != "" {
		parsed, err := netip.ParseAddr(ip)
		if err == nil {
			ipAddr = &parsed
		}
	}

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    id,
		UserAgent: userAgent,
		Ip:        ipAddr,
	})
	if err != nil {
		return "", err
	}

	return session.ID.String(), nil
}

// RevokeSession implements rpc.UserStore
func (s *UserStore) RevokeSession(ctx context.Context, sessionID string) error {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}

	return s.queries.RevokeSession(ctx, id)
}

// RevokeUserSessions implements rpc.UserStore
func (s *UserStore) RevokeUserSessions(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	return s.queries.RevokeAllUserSessions(ctx, id)
}

// IsSessionRevoked implements rpc.UserStore
func (s *UserStore) IsSessionRevoked(ctx context.Context, sessionID string) (bool, error) {
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return false, fmt.Errorf("invalid session id: %w", err)
	}

	return s.queries.IsSessionRevoked(ctx, id)
}

// StoreRefreshToken implements rpc.UserStore
func (s *UserStore) StoreRefreshToken(ctx context.Context, jti, sessionID, userID string, expiresAt time.Time) error {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	_, err = s.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		SessionID: sessionUUID,
		UserID:    userUUID,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	return err
}

// IsRefreshTokenRevoked implements rpc.UserStore
func (s *UserStore) IsRefreshTokenRevoked(ctx context.Context, jti string) (bool, error) {
	id, err := uuid.Parse(jti)
	if err != nil {
		return false, fmt.Errorf("invalid jti: %w", err)
	}

	return s.queries.IsRefreshTokenRevoked(ctx, id)
}

// RevokeRefreshToken implements rpc.UserStore
func (s *UserStore) RevokeRefreshToken(ctx context.Context, jti string) error {
	id, err := uuid.Parse(jti)
	if err != nil {
		return fmt.Errorf("invalid jti: %w", err)
	}

	return s.queries.RevokeRefreshToken(ctx, id)
}

// MarkRefreshTokenRotated implements rpc.UserStore
func (s *UserStore) MarkRefreshTokenRotated(ctx context.Context, oldJTI, newJTI string) error {
	oldID, err := uuid.Parse(oldJTI)
	if err != nil {
		return fmt.Errorf("invalid old jti: %w", err)
	}

	newID, err := uuid.Parse(newJTI)
	if err != nil {
		return fmt.Errorf("invalid new jti: %w", err)
	}

	return s.queries.MarkRefreshTokenRotated(ctx, db.MarkRefreshTokenRotatedParams{
		Jti:      oldID,
		RotatedTo: pgtype.UUID{Bytes: newID, Valid: true},
	})
}

// RevokeUserRefreshTokens implements rpc.UserStore
func (s *UserStore) RevokeUserRefreshTokens(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	return s.queries.RevokeAllUserRefreshTokens(ctx, id)
}
