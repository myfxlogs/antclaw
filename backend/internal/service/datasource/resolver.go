// Package datasource 提供配置热加载能力。
//
// CredentialResolver 封装数据源密钥/端点的运行时解析与热更新。
// Worker 启动时预热，运行时通过 Redis Pub/Sub 接收变更通知。
package datasource

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	redisv9 "github.com/redis/go-redis/v9"

	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
)

// ChangeCallback 是密钥/端点变更时的回调函数签名。
// 用于通知各 collector 更新其持有的客户端实例。
type ChangeCallback func(sourceID, secret, endpoint string)

// CredentialResolver 运行时数据源凭据解析器。
type CredentialResolver struct {
	pool        *pgxpool.Pool
	secretBox   *cryptopkg.SecretBox
	envFallback map[string]string
	logger      *slog.Logger

	mu         sync.RWMutex
	secrets    map[string]string // source_id -> plaintext secret
	endpoints  map[string]string // source_id -> endpoint
	callbacks  map[string][]ChangeCallback // source_id -> callbacks
}

// NewCredentialResolver 构造解析器。
// envFallback 提供环境变量兜底值，如 {"fred": os.Getenv("ANTCLAW_FRED_API_KEY")}
func NewCredentialResolver(pool *pgxpool.Pool, box *cryptopkg.SecretBox, envFallback map[string]string, logger *slog.Logger) *CredentialResolver {
	return &CredentialResolver{
		pool:        pool,
		secretBox:   box,
		envFallback: envFallback,
		logger:      logger,
		secrets:     make(map[string]string),
		endpoints:   make(map[string]string),
		callbacks:   make(map[string][]ChangeCallback),
	}
}

// GetSecret 返回指定数据源的明文密钥；不存在返回空串。
// 热路径，轻量读锁。
func (r *CredentialResolver) GetSecret(sourceID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if v, ok := r.secrets[sourceID]; ok {
		return v
	}
	// 运行时未加载，尝试 fallback
	if v, ok := r.envFallback[sourceID]; ok {
		return v
	}
	return ""
}

// GetEndpoint 返回指定数据源的端点地址；不存在返回空串。
func (r *CredentialResolver) GetEndpoint(sourceID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.endpoints[sourceID]
}

// RegisterCallback 注册变更回调；Reload 成功后会触发。
func (r *CredentialResolver) RegisterCallback(sourceID string, cb ChangeCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callbacks[sourceID] = append(r.callbacks[sourceID], cb)
}

// OnChange 是 RegisterCallback 的别名，语义更清晰。
func (r *CredentialResolver) OnChange(sourceID string, cb ChangeCallback) {
	r.RegisterCallback(sourceID, cb)
}

// FireAll 触发所有已注册回调，把当前缓存值推入各 client。
// Worker 启动时：ReloadAll → 构造 client → 注册 OnChange → FireAll
func (r *CredentialResolver) FireAll() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for id, cbs := range r.callbacks {
		secret := r.secrets[id]
		endpoint := r.endpoints[id]
		for _, cb := range cbs {
			cb(id, secret, endpoint)
		}
	}
}

// Reload 从数据库重新加载单个数据源的密钥与端点。
func (r *CredentialResolver) Reload(ctx context.Context, sourceID string) error {
	svc := NewService(r.pool, r.secretBox, nil)

	// 读取密钥
	secret, err := svc.GetSecret(ctx, sourceID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	// 读取端点
	cfg, _ := svc.Get(ctx, sourceID)
	endpoint := ""
	if cfg != nil {
		endpoint = cfg.Endpoint
	}

	// 更新缓存
	r.mu.Lock()
	r.secrets[sourceID] = secret
	if endpoint != "" {
		r.endpoints[sourceID] = endpoint
	}
	callbacks := r.callbacks[sourceID]
	r.mu.Unlock()

	// 触发回调（在锁外执行，避免阻塞）
	for _, cb := range callbacks {
		cb(sourceID, secret, endpoint)
	}

	if r.logger != nil {
		r.logger.Info("credential reloaded", "source", sourceID, "has_secret", secret != "")
	}
	return nil
}

// ReloadAll 遍历数据库中所有数据源并逐个 Reload。
// 仅 Worker 启动时调用。
func (r *CredentialResolver) ReloadAll(ctx context.Context) error {
	svc := NewService(r.pool, r.secretBox, nil)
	configs, err := svc.List(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := r.Reload(ctx, cfg.SourceID); err != nil {
			if r.logger != nil {
				r.logger.Warn("reload failed during warm-up", "source", cfg.SourceID, "error", err)
			}
		}
	}
	return nil
}

// StartSubscriber 启动 Redis Pub/Sub 订阅，返回 stop 函数。
// 内部自动重连，不会阻塞业务。
func (r *CredentialResolver) StartSubscriber(ctx context.Context, rdb *redisv9.Client) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		for {
			err := r.subscribeLoop(ctx, rdb)
			if ctx.Err() != nil {
				return
			}
			if r.logger != nil {
				r.logger.Warn("subscriber disconnected, reconnecting...", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}
	}()

	return func() {
		cancel()
		<-stopped
	}
}

func (r *CredentialResolver) subscribeLoop(ctx context.Context, rdb *redisv9.Client) error {
	pubsub := rdb.Subscribe(ctx, "datasource:changed")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return errors.New("pubsub channel closed")
			}
			var evt struct {
				SourceID string `json:"source_id"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				if r.logger != nil {
					r.logger.Warn("failed to unmarshal pubsub message", "payload", msg.Payload, "error", err)
				}
				continue
			}
			if evt.SourceID == "" {
				continue
			}
			if err := r.Reload(ctx, evt.SourceID); err != nil {
				if r.logger != nil {
					r.logger.Warn("reload failed", "source", evt.SourceID, "error", err)
				}
			}
		}
	}
}

// ResolveFredKey 解析 FRED key：DB → ENV → 空串
// 返回值：key, 是否来自DB
func ResolveFredKey(r *CredentialResolver) (string, bool) {
	// 1) 尝试 resolver 缓存（来自 DB）
	if key := r.GetSecret("fred"); key != "" {
		return key, true
	}
	// 2) ENV fallback
	if key := r.envFallback["fred"]; key != "" {
		return key, false
	}
	// 3) 无默认值，返回空字符串
	return "", false
}

// BuildEnvFallback 已废弃。所有数据源密钥统一通过 /datasources UI 配置并持久化到 DB。
// 保留空实现以兼容历史调用点。
func BuildEnvFallback() map[string]string {
	return map[string]string{}
}
