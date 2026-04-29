// M-F: AI 记忆存储（user_id + scope + key → value）。基于 ai_memories 表。
package ai

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MemoryItem 单条记忆。
type MemoryItem struct {
	ID        string
	UserID    string
	Scope     string
	Key       string
	Value     string
	ExpiresAt *time.Time
	CreatedAt time.Time
}

// MemoryStore 记忆 CRUD。
type MemoryStore struct {
	pool *pgxpool.Pool
}

func NewMemoryStore(pool *pgxpool.Pool) *MemoryStore {
	return &MemoryStore{pool: pool}
}

// Remember 保存记忆。ttl=0 表示永久。
func (m *MemoryStore) Remember(ctx context.Context, userID, scope, key, value string, ttlSeconds int64) (string, error) {
	if m.pool == nil {
		return "", errors.New("ai memory: pool not configured")
	}
	if userID == "" || key == "" {
		return "", errors.New("ai memory: user_id and key required")
	}
	if scope == "" {
		scope = "global"
	}
	id := memID(userID, scope, key)
	var expires *time.Time
	if ttlSeconds > 0 {
		t := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
		expires = &t
	}
	_, err := m.pool.Exec(ctx, `
		INSERT INTO ai_memories(id, user_id, scope, key, value, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6, NOW())
		ON CONFLICT (id) DO UPDATE SET
			value = EXCLUDED.value,
			expires_at = EXCLUDED.expires_at,
			created_at = NOW()`,
		id, userID, scope, key, value, expires)
	return id, err
}

// Recall 取记忆；过期返回 not found。
func (m *MemoryStore) Recall(ctx context.Context, userID, scope, key string) (*MemoryItem, error) {
	if m.pool == nil {
		return nil, errors.New("ai memory: pool not configured")
	}
	if scope == "" {
		scope = "global"
	}
	row := m.pool.QueryRow(ctx, `
		SELECT id, user_id, scope, key, value, expires_at, created_at
		  FROM ai_memories
		 WHERE user_id=$1 AND scope=$2 AND key=$3
		   AND (expires_at IS NULL OR expires_at > NOW())
		 LIMIT 1`, userID, scope, key)
	it := &MemoryItem{}
	if err := row.Scan(&it.ID, &it.UserID, &it.Scope, &it.Key, &it.Value, &it.ExpiresAt, &it.CreatedAt); err != nil {
		return nil, err
	}
	return it, nil
}

// Search 当前实现为简单子串匹配（key/value 任一含 query）。
func (m *MemoryStore) Search(ctx context.Context, userID, query string, limit int) ([]MemoryItem, error) {
	if m.pool == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := "%" + strings.ToLower(query) + "%"
	rows, err := m.pool.Query(ctx, `
		SELECT id, user_id, scope, key, value, expires_at, created_at
		  FROM ai_memories
		 WHERE user_id=$1
		   AND (LOWER(key) LIKE $2 OR LOWER(value) LIKE $2)
		   AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY created_at DESC
		 LIMIT $3`, userID, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MemoryItem
	for rows.Next() {
		var it MemoryItem
		if err := rows.Scan(&it.ID, &it.UserID, &it.Scope, &it.Key, &it.Value, &it.ExpiresAt, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

// memID 稳定 hash：sha1(user|scope|key) 前 16 字符。
func memID(userID, scope, key string) string {
	h := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s", userID, scope, key)))
	return hex.EncodeToString(h[:])[:16]
}

// Pool 暴露内部 pool 供 handler 构造 Memory/RateLimiter（M-F）。
// 注：保留方法在 memory.go 内是因为 service.go 已较大，避免无意触发编译重组。
func (s *Service) Pool() *pgxpool.Pool { return s.pool }
