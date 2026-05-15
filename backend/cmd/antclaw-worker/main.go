// Package main implements the AntClaw worker process entry point.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	redisv9 "github.com/redis/go-redis/v9"

	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/mql5"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/service/calendar"
	"github.com/antclaw/antclaw/internal/service/datasource"
	"github.com/antclaw/antclaw/internal/service/macro"
)

// 全局 Redis 客户端，用于发布 Job 实时事件。
var globalRedis *redis.Client

type jobEvent struct {
	JobID      string `json:"job_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	StartedAt  int64  `json:"started_at,omitempty"`
	FinishedAt int64  `json:"finished_at,omitempty"`
	Error      string `json:"error,omitempty"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 连接数据库
	logger.Info("Connecting to PostgreSQL...")
	dbpool, err := postgres.NewPoolFromEnv()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	logger.Info("Database connected")

	// 连接Redis
	logger.Info("Connecting to Redis...")
	redisClient := redis.NewClientFromEnv()
	if err := redisClient.Ping(context.Background()); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	logger.Info("Redis connected")
	globalRedis = redisClient

	// 初始化 CredentialResolver：支持运行时热加载
	var secretBox *cryptopkg.SecretBox
	masterKey, err := cryptopkg.LoadOrCreateMasterKey()
	if err != nil {
		logger.Warn("master key load/create failed; datasource encryption disabled", "error", err)
	}
	if masterKey != "" {
		secretBox, err = cryptopkg.NewSecretBox(masterKey)
		if err != nil {
			logger.Warn("SecretBox init failed; datasource encryption disabled", "error", err)
			secretBox = nil
		}
	}
	envFallback := datasource.BuildEnvFallback()
	resolver := datasource.NewCredentialResolver(dbpool, secretBox, envFallback, logger)
	if err := resolver.ReloadAll(context.Background()); err != nil {
		logger.Warn("warm-up credentials failed", "error", err)
	}

	mql5Fetcher := mql5.NewFetcher()

	resolver.OnChange("mql5", func(sourceID, secret, endpoint string) {
		mql5Fetcher.SetBaseURL(endpoint)
		logger.Info("mql5 endpoint hot-reloaded", "endpoint", endpoint)
	})

	// 触发一次回调：把 ReloadAll 已读到的值推入各 client
	resolver.FireAll()

	// 启动订阅：后续管理端修改会触发回调
	stopSub := resolver.StartSubscriber(context.Background(), redisClient.Raw())
	defer stopSub()

	// 创建Repository
	calendarRepo := postgres.NewCalendarRepository(dbpool)
	macroRepo := postgres.NewMacroRepository(dbpool)

	// 创建Service
	calendarSvc := calendar.NewCalendarService(calendarRepo, redisClient, logger)
	fredKey := resolver.GetSecret("fred")
	macroSvc := macro.NewMacroService(macroRepo, fredKey, logger)

	// 热更新 macroSvc 的 FRED key（首次和后续 DB 变更均生效）
	resolver.OnChange("fred", func(sourceID, secret, endpoint string) {
		macroSvc.SetFredKey(secret)
		logger.Info("fred config hot-reloaded", "has_secret", secret != "", "has_endpoint", endpoint != "")
	})
	// 立即应用已加载的 key（FireAll 已在上面执行过，但那时 macroSvc 还未创建）
	macroSvc.SetFredKey(fredKey)

	// 启动定时采集
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 构建 jobID → 执行函数 map（供定时器与手动触发共享）
	jobRunners := buildJobRunners(ctx, dbpool, calendarSvc, macroSvc, logger)

	// 一次性 regime 历史回填（仅在 regime_transitions 为空时生效）
	if err := backfillRegimeHistory(ctx, dbpool, logger); err != nil {
		logger.Warn("regime backfill failed", "error", err)
	}

	// 数据采集循环 - 启用所有数据源
	go runCollectionLoop(ctx, dbpool, calendarSvc, macroSvc, mql5Fetcher, logger)

	// 订阅 jobs:trigger 频道，支持从管理端手动触发任意 job
	go subscribeJobTriggers(ctx, redisClient.Raw(), jobRunners, logger)

	// 客户端智能推送调度循环
	go runPushLoop(ctx, dbpool, redisClient.Raw(), logger)

	// 启动播种：为所有已知 Job 写入 pending 快照，避免页面显示"未运行"
	seedInitialJobSnapshots(ctx, redisClient.Raw(), jobRunners, logger)

	logger.Info("AntClaw Worker started")
	logger.Info("Enabled data collection:")
	logger.Info("  Calendar sync (hourly)")
	logger.Info("  Macro data (every 4 hours)")
	logger.Info("  COT holdings (every 6 hours)")
	logger.Info("  Price data (every 6 hours)")
	logger.Info("  Sentiment data (hourly)")
	logger.Info("  Onchain data (hourly)")
	logger.Info("Enabled push notifications:")
	logger.Info("  Calendar pre/actual/surprise (every 1m)")
	logger.Info("  Daily digest (every 30m)")
	logger.Info("  Macro/Options/Onchain (every 1h)")
	logger.Info("  Carry/Regime/Risk (every 4h)")
	logger.Info("  COT/Calibration (every 6h)")
	logger.Info("Press Ctrl+C to stop.")
	<-sigChan
	logger.Info("Shutting down...")
	cancel()
	time.Sleep(500 * time.Millisecond)
	logger.Info("Worker stopped")
}

// 数据采集循环
func runCollectionLoop(ctx context.Context, dbpool *pgxpool.Pool, calendarSvc *calendar.CalendarService, macroSvc *macro.MacroService, mql5Fetcher *apiclient.MQL5Fetcher, logger *slog.Logger) {
	// 首次运行：立即执行一次全量采集
	logger.Info("Running initial full data collection")
	runAllCollections(ctx, dbpool, calendarSvc, macroSvc, mql5Fetcher, logger)

	// 创建定时器
	calendarTicker := time.NewTicker(1 * time.Hour)   // 每小时采集日历
	macroTicker := time.NewTicker(4 * time.Hour)      // 每4小时采集宏观数据
	actualsTicker := time.NewTicker(30 * time.Minute) // 每30分钟更新实际值
	cotTicker := time.NewTicker(6 * time.Hour)        // 每6小时采集COT
	priceTicker := time.NewTicker(6 * time.Hour)      // 每6小时采集价格
	sentimentTicker := time.NewTicker(1 * time.Hour)  // 每小时采集情绪
	onchainTicker := time.NewTicker(1 * time.Hour)    // 每小时采集链上
	extrasTicker := time.NewTicker(30 * time.Minute)  // Phase 2: 派生与期权数据
	heavyTicker := time.NewTicker(6 * time.Hour)      // Phase 2 重计算: regime/wyckoff/walkforward/calibration

	defer calendarTicker.Stop()
	defer macroTicker.Stop()
	defer actualsTicker.Stop()
	defer cotTicker.Stop()
	defer priceTicker.Stop()
	defer sentimentTicker.Stop()
	defer onchainTicker.Stop()
	defer extrasTicker.Stop()
	defer heavyTicker.Stop()

	logger.Info("Entering scheduled collection mode")
	logger.Info("Collection frequencies:")
	logger.Info("  Calendar sync: hourly")
	logger.Info("  Macro data: every 4 hours")
	logger.Info("  Actuals update: every 30 minutes")
	logger.Info("  COT holdings: every 6 hours")
	logger.Info("  Price data: every 6 hours")
	logger.Info("  Sentiment data: hourly")
	logger.Info("  Onchain data: hourly")

	for {
		select {
		case <-ctx.Done():
			return
		case <-calendarTicker.C:
			logger.Info("Scheduled trigger: calendar sync")
			runWithEvent(ctx, logger, "calendar-sync", "calendar-sync", func() error { return runCalendarCollection(ctx, calendarSvc, logger) })
		case <-macroTicker.C:
			logger.Info("Scheduled trigger: macro sync")
			runWithEvent(ctx, logger, "macro-sync", "macro-sync", func() error { return runMacroCollection(ctx, macroSvc, logger) })
		case <-actualsTicker.C:
			logger.Info("Scheduled trigger: actuals update")
			runWithEvent(ctx, logger, "actuals-update", "actuals-update", func() error { return runActualsUpdate(ctx, calendarSvc, logger) })
		case <-cotTicker.C:
			logger.Info("Scheduled trigger: COT holdings")
			runWithEvent(ctx, logger, "cot-sync", "cot-sync", func() error { return collectCOT(ctx, dbpool, logger) })
		case <-priceTicker.C:
			logger.Info("Scheduled trigger: price data")
			runWithEvent(ctx, logger, "price-sync", "price-sync", func() error { return collectPrices(ctx, dbpool, logger) })
		case <-sentimentTicker.C:
			logger.Info("Scheduled trigger: sentiment data")
			runWithEvent(ctx, logger, "sentiment-sync", "sentiment-sync", func() error { return collectSentiment(ctx, dbpool, logger) })
		case <-onchainTicker.C:
			logger.Info("Scheduled trigger: onchain data")
			runWithEvent(ctx, logger, "onchain-sync", "onchain-sync", func() error { return collectOnchain(ctx, dbpool, logger) })
		case <-extrasTicker.C:
			logger.Info("Scheduled trigger: derived & options extras")
			runWithEvent(ctx, logger, "calendar-titles", "calendar-titles", func() error { return collectCalendarTitles(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "calendar-surprise", "calendar-surprise", func() error { return collectCalendarSurprise(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "event-impact", "event-impact", func() error { return collectEventImpact(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "micro-snapshot", "micro-snapshot", func() error { return collectMicroSnapshots(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "gex-snapshot", "gex-snapshot", func() error { return collectGEX(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "iv-skew", "iv-skew", func() error { return collectIVSkew(ctx, dbpool, logger) })
		case <-heavyTicker.C:
			logger.Info("Scheduled trigger: heavy recompute (regime/wyckoff/walkforward/cot-calibration)")
			runWithEvent(ctx, logger, "cot-calibration", "cot-calibration", func() error { return calibrateCOT(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "wyckoff-events", "wyckoff-events", func() error { return detectWyckoff(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "walkforward", "walkforward", func() error { return runWalkforward(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "regime-overlay", "regime-overlay", func() error { return computeRegimeOverlay(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "transition-matrix", "transition-matrix", func() error { return buildTransitionMatrix(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "cot-signal-outcomes", "cot-signal-outcomes", func() error { return collectCOTSignalOutcomes(ctx, dbpool, logger) })
			runWithEvent(ctx, logger, "data-snapshot", "data-snapshot", func() error { return snapshotCOTIndexToDataSnapshots(ctx, dbpool, logger) })
		}
	}
}

// 执行所有采集任务 - 统一用 runWithEvent 包装以发布实时事件
func runAllCollections(ctx context.Context, dbpool *pgxpool.Pool, calendarSvc *calendar.CalendarService, macroSvc *macro.MacroService, mql5Fetcher *apiclient.MQL5Fetcher, logger *slog.Logger) {
	runWithEvent(ctx, logger, "calendar-sync", "calendar-sync", func() error { return runCalendarCollection(ctx, calendarSvc, logger) })
	runWithEvent(ctx, logger, "macro-sync", "macro-sync", func() error { return runMacroCollection(ctx, macroSvc, logger) })
	runWithEvent(ctx, logger, "actuals-update", "actuals-update", func() error { return runActualsUpdate(ctx, calendarSvc, logger) })
	runWithEvent(ctx, logger, "cot-sync", "cot-sync", func() error { return collectCOT(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "price-sync", "price-sync", func() error { return collectPrices(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "sentiment-sync", "sentiment-sync", func() error { return collectSentiment(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "onchain-sync", "onchain-sync", func() error { return collectOnchain(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "intraday-sync", "intraday-sync", func() error { return collectIntraday(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "defi-sync", "defi-sync", func() error { return collectDefi(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "vix-term-sync", "vix-term-sync", func() error { return collectVIXTermStructure(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "dvol-sync", "dvol-sync", func() error { return collectDVOL(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "cot-analysis", "cot-analysis", func() error { return analyzeCOT(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "macro-regime", "macro-regime", func() error { return analyzeMacroRegime(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "flow-divergence", "flow-divergence", func() error { return analyzeFlowDivergence(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "volume-profile", "volume-profile", func() error { return analyzeVolumeProfile(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "outcome-evaluator", "outcome-evaluator", func() error { return evaluateSignalOutcomes(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "transition-matrix", "transition-matrix", func() error { return buildTransitionMatrix(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "alert-evaluator", "alert-evaluator", func() error { return evaluateAlerts(ctx, dbpool, logger) })

	// === Phase 2: 派生与外部期权数据采集 ===
	runWithEvent(ctx, logger, "calendar-titles", "calendar-titles", func() error { return collectCalendarTitles(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "calendar-surprise", "calendar-surprise", func() error { return collectCalendarSurprise(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "event-impact", "event-impact", func() error { return collectEventImpact(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "cot-calibration", "cot-calibration", func() error { return calibrateCOT(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "wyckoff-events", "wyckoff-events", func() error { return detectWyckoff(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "walkforward", "walkforward", func() error { return runWalkforward(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "gex-snapshot", "gex-snapshot", func() error { return collectGEX(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "iv-skew", "iv-skew", func() error { return collectIVSkew(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "micro-snapshot", "micro-snapshot", func() error { return collectMicroSnapshots(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "regime-overlay", "regime-overlay", func() error { return computeRegimeOverlay(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "cot-signal-outcomes", "cot-signal-outcomes", func() error { return collectCOTSignalOutcomes(ctx, dbpool, logger) })
	runWithEvent(ctx, logger, "data-snapshot", "data-snapshot", func() error { return snapshotCOTIndexToDataSnapshots(ctx, dbpool, logger) })
}

// runWithEvent 包装一个任务执行：先检查启用状态，已禁用则跳过；
// 否则发送 running → succeeded / failed 事件。
// 正常错误通过 fn() 的 error 返回传播；panic 作为最后兜底。
func runWithEvent(ctx context.Context, logger *slog.Logger, jobID, name string, fn func() error) {
	if !isJobEnabled(ctx, jobID) {
		logger.Info("job skipped (disabled)", "job", jobID)
		return
	}
	start := time.Now().Unix()
	publishJobEvent(ctx, logger, jobEvent{JobID: jobID, Name: name, Status: "running", StartedAt: start})

	status := "succeeded"
	errMsg := ""
	defer func() {
		if r := recover(); r != nil {
			status = "failed"
			errMsg = fmt.Sprintf("panic: %v", r)
			logger.Error("job panic", "job", jobID, "panic", r)
		}
		publishJobEvent(ctx, logger, jobEvent{
			JobID: jobID, Name: name, Status: status,
			StartedAt: start, FinishedAt: time.Now().Unix(),
			Error: errMsg,
		})
	}()

	if err := fn(); err != nil {
		status = "failed"
		errMsg = err.Error()
		logger.Error("job failed", "job", jobID, "error", err)
	}
}

// isJobEnabled 查询 Redis 中 jobs:enabled:{jobID} 状态。默认启用。
func isJobEnabled(ctx context.Context, jobID string) bool {
	if globalRedis == nil {
		return true
	}
	val, err := globalRedis.Raw().Get(ctx, fmt.Sprintf("jobs:enabled:%s", jobID)).Result()
	if err != nil {
		return true
	}
	return val != "false"
}

// seedInitialJobSnapshots 为所有已知 Job 写入一条 pending 状态快照（若尚不存在），
// 防止 /jobs 页面在 Worker 重启后、首次执行前出现 "未运行"。
func seedInitialJobSnapshots(ctx context.Context, rdb *redisv9.Client, runners map[string]func(), logger *slog.Logger) {
	if rdb == nil {
		return
	}
	ids := make([]string, 0, len(runners))
	for id := range runners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	seeded := 0
	for _, id := range ids {
		key := fmt.Sprintf("jobs:status:%s", id)
		exists, err := rdb.Exists(ctx, key).Result()
		if err != nil || exists > 0 {
			continue
		}
		publishJobEvent(ctx, logger, jobEvent{JobID: id, Name: id, Status: "pending"})
		seeded++
	}
	logger.Info("seeded initial job snapshots", "count", seeded, "total", len(ids))
}

// 财经日历采集（返回 error 让外层 runWithEvent 统一处理）
func runCalendarCollection(ctx context.Context, svc *calendar.CalendarService, logger *slog.Logger) error {
	result, err := svc.SyncWeek(ctx)
	if err != nil {
		logger.Error("Calendar sync failed", "error", err)
		return fmt.Errorf("calendar sync: %w", err)
	}
	logger.Info("Calendar sync completed", "inserted", result.Inserted, "total", result.Total)
	return nil
}

// 宏观数据采集
func runMacroCollection(ctx context.Context, svc *macro.MacroService, logger *slog.Logger) error {
	series := []string{"GDP", "CPIAUCSL", "UNRATE", "FEDFUNDS", "T10Y2Y", "DGS10", "TB3MS", "T10YIE"}
	result, err := svc.SyncFREDIndicators(ctx, series)
	if err != nil {
		logger.Error("Macro sync failed", "error", err)
		return fmt.Errorf("macro sync: %w", err)
	}
	if result.Inserted == 0 {
		return fmt.Errorf("macro sync inserted 0 records: 可能原因 1) FRED API Key 未在 /datasources 配置 2) 密钥解密失败（master key 漂移，需重新保存密钥） 3) FRED API 网络不可达 4) 无新数据（查看 worker 日志中的 FRED series 警告）")
	}
	logger.Info("Macro sync completed", "inserted", result.Inserted)
	return nil
}

// 实际值更新
func runActualsUpdate(ctx context.Context, svc *calendar.CalendarService, logger *slog.Logger) error {
	result, err := svc.SyncActuals(ctx)
	if err != nil {
		logger.Error("Actuals sync failed", "error", err)
		return fmt.Errorf("actuals update: %w", err)
	}
	logger.Info("Actuals update completed", "updated", result.Updated)
	return nil
}

// publishJobEvent 将 Job 状态写入 Redis Streams，并更新最后状态快照。
func publishJobEvent(ctx context.Context, logger *slog.Logger, evt jobEvent) {
	if globalRedis == nil {
		return
	}
	data, err := json.Marshal(evt)
	if err != nil {
		logger.Error("marshal job event failed", "error", err)
		return
	}
	rdb := globalRedis.Raw()
	// 写入事件流
	_ = rdb.XAdd(ctx, &redisv9.XAddArgs{
		Stream: "stream:jobs_events",
		Values: map[string]interface{}{
			"job_id": evt.JobID,
			"name":   evt.Name,
			"status": evt.Status,
			"data":   string(data),
		},
	}).Err()
	// 更新最后状态快照，便于后续快速查询
	key := fmt.Sprintf("jobs:status:%s", evt.JobID)
	_ = globalRedis.Set(ctx, key, string(data), 0)
}
