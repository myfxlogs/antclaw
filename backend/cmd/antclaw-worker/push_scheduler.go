// 客户端智能推送：调度循环。
//
// 在 runCollectionLoop 之外独立运行，按文档 4.x 的频率矩阵调度各推送检测器。
// 所有推送最终统一走 notify.Service.Send。
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/notify"
)

// pushEnv 封装推送循环所需依赖。
type pushEnv struct {
	pool  *pgxpool.Pool
	q     *db.Queries
	svc   *notify.Service
	rdb   *redis.Client
	log   *slog.Logger
}

// runPushLoop 启动独立推送调度循环（不阻塞采集循环）。
func runPushLoop(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger) {
	q := db.New(pool)
	svc := notify.NewService(q, rdb)
	env := &pushEnv{
		pool: pool,
		q:    q,
		svc:  svc,
		rdb:  rdb,
		log:  logger.With("component", "push"),
	}

	env.log.Info("Push scheduler starting", "frequencies", "1m/30m/1h/4h/6h")

	// Ticker 频率矩阵（文档 4.x）
	t1m := time.NewTicker(1 * time.Minute)
	t30m := time.NewTicker(30 * time.Minute)
	t1h := time.NewTicker(1 * time.Hour)
	t4h := time.NewTicker(4 * time.Hour)
	t6h := time.NewTicker(6 * time.Hour)
	defer t1m.Stop()
	defer t30m.Stop()
	defer t1h.Stop()
	defer t4h.Stop()
	defer t6h.Stop()

	for {
		select {
		case <-ctx.Done():
			env.log.Info("Push scheduler stopped")
			return
		case <-t1m.C:
			runWithTimeout(ctx, env, 2*time.Minute, "calendar", pushCalendar(env))
		case <-t30m.C:
			runWithTimeout(ctx, env, 3*time.Minute, "daily-digest", pushDailyDigest(env))
		case <-t1h.C:
			runWithTimeout(ctx, env, 3*time.Minute, "macro", pushMacro(env))
			runWithTimeout(ctx, env, 2*time.Minute, "options", pushOptions(env))
			runWithTimeout(ctx, env, 2*time.Minute, "onchain", pushOnchain(env))
			// weekly outlook 窗口检查复用 1h ticker，内部判断时间窗口
			runWithTimeout(ctx, env, 3*time.Minute, "weekly-outlook", pushWeeklyOutlook(env))
		case <-t4h.C:
			runWithTimeout(ctx, env, 3*time.Minute, "carry", pushCarry(env))
			runWithTimeout(ctx, env, 2*time.Minute, "regime-transition", pushRegimeTransition(env))
			runWithTimeout(ctx, env, 2*time.Minute, "risk-confluence", pushRiskConfluence(env))
		case <-t6h.C:
			runWithTimeout(ctx, env, 3*time.Minute, "cot-release", pushCOTRelease(env))
			runWithTimeout(ctx, env, 2*time.Minute, "cot-signal", pushCOTSignal(env))
			runWithTimeout(ctx, env, 2*time.Minute, "calibration", pushCalibration(env))
		}
	}
}

// runWithTimeout 在超时保护下执行推送检测；不会因单个检测器 panic 而崩溃。
func runWithTimeout(ctx context.Context, env *pushEnv, timeout time.Duration, name string, fn func(context.Context, *pushEnv)) {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	func() {
		defer func() {
			if r := recover(); r != nil {
				env.log.Error("push detector panicked", "name", name, "panic", r)
			}
		}()
		start := time.Now()
		fn(c, env)
		env.log.Debug("push detector finished", "name", name, "elapsed", time.Since(start))
	}()
}

// 所有推送检测器已实现：
//   - push_calendar.go   : pushCalendar
//   - push_digest.go     : pushDailyDigest, pushWeeklyOutlook
//   - push_macro.go      : pushMacro, pushOptions, pushOnchain, pushCarry, pushRegimeTransition, pushRiskConfluence
//   - push_cot.go        : pushCOTRelease, pushCOTSignal, pushCalibration
