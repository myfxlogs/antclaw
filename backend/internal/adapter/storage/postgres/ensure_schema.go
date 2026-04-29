// Package postgres bootstrap ensures required tables exist at API startup.
// 采用幂等 SQL，避免依赖 goose/sqlc 等外部迁移工具；可安全多次调用。
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureAdminSchema 为 Admin 相关功能（审计日志、数据源配置等）创建必需的表结构。
// 所有语句都使用 IF NOT EXISTS，幂等安全；在 API 启动时调用。
func EnsureAdminSchema(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		// ===== 用户表（必须在 sessions/refresh_tokens 之前） =====
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT NOT NULL UNIQUE,
			email_verified_at TIMESTAMPTZ,
			username TEXT,
			display_name TEXT,
			password_hash TEXT NOT NULL,
			password_version INTEGER NOT NULL DEFAULT 1,
			role TEXT NOT NULL DEFAULT 'user',
			status TEXT NOT NULL DEFAULT 'active',
			locale TEXT NOT NULL DEFAULT 'en-US',
			timezone TEXT NOT NULL DEFAULT 'UTC',
			totp_secret_enc BYTEA,
			totp_enabled BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL`,

		// ===== 认证相关：sessions / refresh_tokens / password_resets =====
		// 这些表登录流程必需；用 IF NOT EXISTS 兼容已通过 goose 跑过迁移的环境。
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			user_agent TEXT NOT NULL DEFAULT '',
			ip INET,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			revoked_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_lastseen ON sessions(user_id, last_seen_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(revoked_at) WHERE revoked_at IS NULL`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			jti UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			rotated_to UUID
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL`,

		`CREATE TABLE IF NOT EXISTS password_resets (
			token_hash BYTEA PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			purpose TEXT NOT NULL CHECK (purpose IN ('password_reset', 'email_verify')),
			issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '15 minutes'),
			consumed_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_password_resets_expires ON password_resets(expires_at) WHERE consumed_at IS NULL`,

		// ===== 审计日志 =====
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id UUID,
			action TEXT NOT NULL,
			resource TEXT NOT NULL DEFAULT '',
			details TEXT NOT NULL DEFAULT '',
			ip_address TEXT,
			user_agent TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			hash_prev BYTEA NOT NULL DEFAULT '\x',
			hash_self BYTEA NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC)`,

		// ===== 数据源配置（按数据源粒度，敏感字段使用 Argon2id+AES-256-GCM 加密） =====
		`CREATE TABLE IF NOT EXISTS data_source_configs (
			source_id        TEXT PRIMARY KEY,
			name             TEXT NOT NULL,
			kind             TEXT NOT NULL,
			endpoint         TEXT NOT NULL DEFAULT '',
			secret_ciphertext BYTEA,
			secret_salt       BYTEA,
			secret_nonce      BYTEA,
			has_secret       BOOLEAN NOT NULL DEFAULT FALSE,
			updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by       TEXT NOT NULL DEFAULT ''
		)`,

		// ===== 回测策略管理 =====
		`CREATE TABLE IF NOT EXISTS strategies (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name            TEXT NOT NULL,
			kind            TEXT NOT NULL,
			symbol          TEXT NOT NULL,
			timeframe       TEXT NOT NULL,
			params          JSONB NOT NULL DEFAULT '{}'::jsonb,
			schedule_cron   TEXT NOT NULL DEFAULT '@hourly',
			enabled         BOOLEAN NOT NULL DEFAULT FALSE,
			status          TEXT NOT NULL DEFAULT 'draft',
			description     TEXT NOT NULL DEFAULT '',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by      TEXT NOT NULL DEFAULT '',
			updated_by      TEXT NOT NULL DEFAULT '',
			last_run_at     TIMESTAMPTZ,
			last_run_status TEXT,
			deleted_at      TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_strategies_enabled ON strategies(enabled) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_strategies_kind ON strategies(kind) WHERE deleted_at IS NULL`,

		// ===== 策略运行历史 =====
		`CREATE TABLE IF NOT EXISTS strategy_runs (
			run_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			strategy_id     UUID NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
			started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			finished_at     TIMESTAMPTZ,
			status          TEXT NOT NULL DEFAULT 'running',
			metrics         JSONB NOT NULL DEFAULT '{}'::jsonb,
			mock            BOOLEAN NOT NULL DEFAULT TRUE,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_strategy_runs_strategy ON strategy_runs(strategy_id, created_at DESC)`,

		// ===== 系统 AI 配置 =====
		`CREATE TABLE IF NOT EXISTS system_ai_configs (
			provider_id        TEXT PRIMARY KEY,
			name               TEXT NOT NULL,
			base_url           TEXT NOT NULL DEFAULT '',
			organization       TEXT NOT NULL DEFAULT '',
			models             TEXT[] NOT NULL DEFAULT '{}',
			default_model      TEXT NOT NULL DEFAULT '',
			temperature        DOUBLE PRECISION NOT NULL DEFAULT 0.2,
			timeout_seconds    INTEGER NOT NULL DEFAULT 60,
			max_tokens         INTEGER NOT NULL DEFAULT 4096,
			purposes           TEXT[] NOT NULL DEFAULT '{}',
			primary_for        TEXT[] NOT NULL DEFAULT '{}',
			secret_ciphertext  BYTEA,
			secret_salt        BYTEA,
			secret_nonce       BYTEA,
			has_secret         BOOLEAN NOT NULL DEFAULT FALSE,
			enabled            BOOLEAN NOT NULL DEFAULT FALSE,
			created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by         TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_system_ai_enabled ON system_ai_configs(enabled)`,
		// 兼容旧库：增量补齐 docs_url / apply_url 列。
		`ALTER TABLE system_ai_configs ADD COLUMN IF NOT EXISTS docs_url  TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE system_ai_configs ADD COLUMN IF NOT EXISTS apply_url TEXT NOT NULL DEFAULT ''`,
		// AI 调用结果缓存（fingerprint 主键）。
		`CREATE TABLE IF NOT EXISTS ai_cache (
			fingerprint TEXT PRIMARY KEY,
			operation   TEXT NOT NULL,
			model       TEXT,
			result      TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at  TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_cache_expires ON ai_cache(expires_at)`,

		`CREATE TABLE IF NOT EXISTS regime_transition_matrix (
			asof_date DATE NOT NULL,
			symbol VARCHAR(32) NOT NULL,
			timeframe VARCHAR(8) NOT NULL,
			from_label VARCHAR(16) NOT NULL,
			to_label VARCHAR(16) NOT NULL,
			probability DOUBLE PRECISION NOT NULL,
			sample_size INT NOT NULL,
			PRIMARY KEY (asof_date, symbol, timeframe, from_label, to_label)
		)`,

		`CREATE TABLE IF NOT EXISTS onchain_metrics (
			time TIMESTAMPTZ NOT NULL,
			asset VARCHAR(16) NOT NULL,
			metric VARCHAR(32) NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			source VARCHAR(32),
			PRIMARY KEY (time, asset, metric)
		)`,

		`CREATE TABLE IF NOT EXISTS quant_strategy_perf (
			asof_date DATE NOT NULL,
			symbol VARCHAR(32) NOT NULL,
			strategy VARCHAR(64) NOT NULL,
			sharpe DOUBLE PRECISION,
			sortino DOUBLE PRECISION,
			drawdown DOUBLE PRECISION,
			win_rate DOUBLE PRECISION,
			sample_trades INT,
			PRIMARY KEY (asof_date, symbol, strategy)
		)`,

		`CREATE TABLE IF NOT EXISTS user_signal_alerts (
			id BIGSERIAL PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL,
			alert_type VARCHAR(32) NOT NULL,
			symbol VARCHAR(32) NOT NULL,
			params JSONB NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			last_fired_at TIMESTAMPTZ,
			cooldown_seconds INT NOT NULL DEFAULT 3600,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE INDEX IF NOT EXISTS idx_user_alerts_user ON user_signal_alerts(user_id) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_user_alerts_active ON user_signal_alerts(enabled, alert_type) WHERE deleted_at IS NULL`,

		`CREATE TABLE IF NOT EXISTS signal_calibration_isotonic (
			signal_type VARCHAR(64) NOT NULL,
			bucket_idx INT NOT NULL,
			x_lower DOUBLE PRECISION,
			x_upper DOUBLE PRECISION,
			p_calibrated DOUBLE PRECISION,
			sample_size INT,
			updated_at TIMESTAMPTZ,
			PRIMARY KEY (signal_type, bucket_idx)
		)`,
		// M-B / M-C / M-E / M-F 增量
		`CREATE TABLE IF NOT EXISTS backtest_trades (
			job_id TEXT NOT NULL, seq INT NOT NULL,
			opened_at TIMESTAMPTZ, closed_at TIMESTAMPTZ,
			side TEXT NOT NULL, entry DOUBLE PRECISION, exit DOUBLE PRECISION,
			pnl DOUBLE PRECISION, pnl_pct DOUBLE PRECISION,
			mfe DOUBLE PRECISION, mae DOUBLE PRECISION,
			cost DOUBLE PRECISION, regime TEXT,
			PRIMARY KEY (job_id, seq)
		)`,
		`CREATE TABLE IF NOT EXISTS backtest_metrics_by_regime (
			job_id TEXT NOT NULL, regime TEXT NOT NULL,
			n_trades INT, sharpe DOUBLE PRECISION, sortino DOUBLE PRECISION,
			max_drawdown DOUBLE PRECISION, win_rate DOUBLE PRECISION,
			PRIMARY KEY (job_id, regime)
		)`,
		`CREATE TABLE IF NOT EXISTS signal_calibrations (
			model_id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			params JSONB NOT NULL,
			n_samples INT,
			brier DOUBLE PRECISION,
			fitted_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_preferences (
			user_id TEXT PRIMARY KEY,
			pairs TEXT[] DEFAULT '{}',
			high_impact_only BOOLEAN DEFAULT FALSE,
			quiet_hours_start INT,
			quiet_hours_end INT,
			timezone TEXT DEFAULT 'UTC'
		)`,
		`CREATE TABLE IF NOT EXISTS user_quotas (
			user_id TEXT PRIMARY KEY,
			tier TEXT NOT NULL DEFAULT 'free',
			ai_calls_today INT DEFAULT 0,
			ai_max_per_day INT DEFAULT 20,
			reset_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS alert_log (
			id BIGSERIAL PRIMARY KEY,
			user_id TEXT, alert_type TEXT, severity TEXT,
			payload JSONB, sent BOOLEAN, reason TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_log_user_time ON alert_log(user_id, created_at DESC)`,
		`CREATE TABLE IF NOT EXISTS ai_memories (
			id TEXT PRIMARY KEY,
			user_id TEXT, scope TEXT, key TEXT, value TEXT,
			expires_at TIMESTAMPTZ, created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ai_memories_user_scope_key ON ai_memories(user_id, scope, key)`,
		`CREATE TABLE IF NOT EXISTS ai_conversations (
			thread_id TEXT PRIMARY KEY, user_id TEXT,
			started_at TIMESTAMPTZ DEFAULT NOW(),
			last_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_messages (
			thread_id TEXT REFERENCES ai_conversations(thread_id) ON DELETE CASCADE,
			seq INT, role TEXT, content TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			PRIMARY KEY (thread_id, seq)
		)`,
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			return fmt.Errorf("ensure admin schema: %w", err)
		}
	}
	// 种子数据：声明已知数据源，便于前端列表展示。
	// ON CONFLICT DO NOTHING 保证不覆盖已配置的密钥。
	if err := ensureDataSourceSeeds(ctx, pool); err != nil {
		return fmt.Errorf("seed data sources: %w", err)
	}
	if err := ensureSystemAIConfigSeeds(ctx, pool); err != nil {
		return fmt.Errorf("seed system ai configs: %w", err)
	}
	if err := ensureStrategySeeds(ctx, pool); err != nil {
		return fmt.Errorf("seed strategies: %w", err)
	}
	return nil
}

