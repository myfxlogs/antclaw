// Package datasource 提供数据源配置（API key / 端点）的存取服务。
//
// 安全模型：
//   - 敏感字段（api key 等）以 Argon2id+AES-256-GCM 加密存储；
//   - 列表 / 读取接口永不返回明文密钥；
//   - 仅业务层（如 Worker、内部 service）能通过 GetSecret 获取明文。
package datasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	redisv9 "github.com/redis/go-redis/v9"

	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
)

// Config 是返回给管理端的数据源元信息（不含密文与明文）。
type Config struct {
	SourceID  string    `json:"source_id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`     // api_key / endpoint / custom_json
	Endpoint  string    `json:"endpoint"` // 非敏感
	HasSecret bool      `json:"has_secret"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// Service 数据源配置服务。所有密钥统一通过 /datasources UI 配置后持久化到 DB。
type Service struct {
	pool   *pgxpool.Pool
	secret *cryptopkg.SecretBox
	rdb    *redisv9.Client // 可选：用于 Update 后发布 datasource:changed 通知 worker
}

// NewService 构造服务；secret 不为 nil 时启用加密能力。
// rdb 可为 nil；不为 nil 时将在 Update 成功后向 "datasource:changed" 频道发布消息，
// 由 worker 端 CredentialResolver.subscribeLoop 接收并触发热加载。
func NewService(pool *pgxpool.Pool, secret *cryptopkg.SecretBox, rdb *redisv9.Client) *Service {
	return &Service{pool: pool, secret: secret, rdb: rdb}
}

// List 列出所有数据源（按 source_id 升序）。永不返回密钥。
func (s *Service) List(ctx context.Context) ([]Config, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT source_id, name, kind, endpoint, has_secret, updated_at, updated_by
		FROM data_source_configs
		ORDER BY source_id
	`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	out := make([]Config, 0, 16)
	for rows.Next() {
		var c Config
		if err := rows.Scan(&c.SourceID, &c.Name, &c.Kind, &c.Endpoint, &c.HasSecret, &c.UpdatedAt, &c.UpdatedBy); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get 单条查询，不返回密钥。
func (s *Service) Get(ctx context.Context, sourceID string) (*Config, error) {
	var c Config
	err := s.pool.QueryRow(ctx, `
		SELECT source_id, name, kind, endpoint, has_secret, updated_at, updated_by
		FROM data_source_configs WHERE source_id = $1
	`, sourceID).Scan(&c.SourceID, &c.Name, &c.Kind, &c.Endpoint, &c.HasSecret, &c.UpdatedAt, &c.UpdatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &c, nil
}

// UpdateInput 更新载荷。endpoint 非空则更新；secret 非空则重新加密存储。
// 显式 ClearSecret=true 时清空已存密钥。
type UpdateInput struct {
	Endpoint    *string // nil 表示不更新；空串表示清空
	Secret      *string // nil 表示不更新；非 nil 即重新加密
	ClearSecret bool
	UpdatedBy   string
}

// Update 写入端点 / 重置密钥。事务保证一致性。
func (s *Service) Update(ctx context.Context, sourceID string, in UpdateInput) error {
	if _, err := s.Get(ctx, sourceID); err != nil {
		return err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if in.Endpoint != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE data_source_configs SET endpoint = $1, updated_at = NOW(), updated_by = $2 WHERE source_id = $3
		`, *in.Endpoint, in.UpdatedBy, sourceID); err != nil {
			return fmt.Errorf("update endpoint: %w", err)
		}
	}

	switch {
	case in.ClearSecret:
		if _, err := tx.Exec(ctx, `
			UPDATE data_source_configs
			SET secret_ciphertext = NULL, secret_salt = NULL, secret_nonce = NULL,
			    has_secret = FALSE, updated_at = NOW(), updated_by = $1
			WHERE source_id = $2
		`, in.UpdatedBy, sourceID); err != nil {
			return fmt.Errorf("clear secret: %w", err)
		}
	case in.Secret != nil:
		if s.secret == nil {
			return errors.New("secret encryption is disabled (master key missing)")
		}
		ct, salt, nonce, err := s.secret.Seal([]byte(*in.Secret))
		if err != nil {
			return fmt.Errorf("seal secret: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			UPDATE data_source_configs
			SET secret_ciphertext = $1, secret_salt = $2, secret_nonce = $3,
			    has_secret = TRUE, updated_at = NOW(), updated_by = $4
			WHERE source_id = $5
		`, ct, salt, nonce, in.UpdatedBy, sourceID); err != nil {
			return fmt.Errorf("update secret: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// 通知 worker 热加载。失败不影响 DB 一致性，下次 worker 重启时仍能从 DB 加载。
	if s.rdb != nil {
		payload, _ := json.Marshal(map[string]string{"source_id": sourceID})
		_ = s.rdb.Publish(ctx, "datasource:changed", payload).Err()
	}
	return nil
}

// GetSecret 仅供业务层调用（非 HTTP 直接暴露）。返回明文密钥；不存在返回空。
func (s *Service) GetSecret(ctx context.Context, sourceID string) (string, error) {
	if s.secret == nil {
		return "", errors.New("secret encryption is disabled")
	}
	var (
		ct, salt, nonce []byte
		hasSecret       bool
	)
	err := s.pool.QueryRow(ctx, `
		SELECT secret_ciphertext, secret_salt, secret_nonce, has_secret
		FROM data_source_configs WHERE source_id = $1
	`, sourceID).Scan(&ct, &salt, &nonce, &hasSecret)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query: %w", err)
	}
	if !hasSecret || len(ct) == 0 {
		return "", nil
	}
	plain, err := s.secret.Open(ct, salt, nonce)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	return string(plain), nil
}

// ErrNotFound 数据源不存在。
var ErrNotFound = errors.New("data source not found")
