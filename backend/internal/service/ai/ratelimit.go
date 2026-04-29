// M-F: AI 调用配额追踪。基于 user_quotas 表（M-E 共用）。
//
// 实现选用日级别（ai_calls_today 字段 + reset_at），不引入 Redis token bucket，
// 避免跨服务一致性复杂度。每次调用前 Check + Acquire；超限返 ErrRateLimited。
package ai

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrRateLimited 配额耗尽。
var ErrRateLimited = errors.New("ai: daily rate limit exceeded")

// RateLimiter 日级配额。
type RateLimiter struct{ pool *pgxpool.Pool }

func NewRateLimiter(pool *pgxpool.Pool) *RateLimiter { return &RateLimiter{pool: pool} }

// Status 当前配额状态。
type Status struct {
	UsedToday  int
	MaxPerDay  int
	Remaining  int
	Allowed    bool
}

// Check 不消耗配额，仅查询。userID 为空 / pool 为空 → 视为放行。
func (r *RateLimiter) Check(ctx context.Context, userID string) (*Status, error) {
	if r.pool == nil || userID == "" {
		return &Status{Allowed: true, MaxPerDay: 1000, Remaining: 1000}, nil
	}
	r.maybeReset(ctx, userID)
	var used, maxN int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(ai_calls_today,0), COALESCE(ai_max_per_day,20)
		  FROM user_quotas WHERE user_id = $1`, userID).Scan(&used, &maxN)
	if err != nil {
		// 用户不存在视为新建：写入默认配额。
		_, _ = r.pool.Exec(ctx, `
			INSERT INTO user_quotas(user_id, tier, ai_calls_today, ai_max_per_day, reset_at)
			VALUES ($1, 'free', 0, 20, NOW() + INTERVAL '1 day')
			ON CONFLICT (user_id) DO NOTHING`, userID)
		return &Status{Allowed: true, MaxPerDay: 20, Remaining: 20}, nil
	}
	rem := maxN - used
	if rem < 0 {
		rem = 0
	}
	return &Status{
		UsedToday: used, MaxPerDay: maxN, Remaining: rem, Allowed: rem > 0,
	}, nil
}

// Acquire 消耗 1 次配额；超限返 ErrRateLimited。
func (r *RateLimiter) Acquire(ctx context.Context, userID string) error {
	if r.pool == nil || userID == "" {
		return nil
	}
	r.maybeReset(ctx, userID)
	res, err := r.pool.Exec(ctx, `
		UPDATE user_quotas SET ai_calls_today = ai_calls_today + 1
		 WHERE user_id = $1 AND ai_calls_today < ai_max_per_day`, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		// 没更新 → 已满 或 无记录
		s, _ := r.Check(ctx, userID)
		if s != nil && s.Remaining <= 0 {
			return ErrRateLimited
		}
		// 第一次调用时 Check 已 Insert，再 retry 一次
		_, err = r.pool.Exec(ctx, `
			UPDATE user_quotas SET ai_calls_today = ai_calls_today + 1
			 WHERE user_id = $1 AND ai_calls_today < ai_max_per_day`, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

// maybeReset 若 reset_at 已过期，把 ai_calls_today 归零，新 reset_at = NOW + 1 day。
func (r *RateLimiter) maybeReset(ctx context.Context, userID string) {
	_, _ = r.pool.Exec(ctx, `
		UPDATE user_quotas
		   SET ai_calls_today = 0,
		       reset_at = NOW() + INTERVAL '1 day'
		 WHERE user_id = $1
		   AND (reset_at IS NULL OR reset_at < NOW())`, userID)
	_ = time.Now
}
