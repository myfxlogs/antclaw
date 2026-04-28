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
	secretBox, err := cryptopkg.NewSecretBox(os.Getenv("ANTCLAW_SECRET_MASTER_KEY"))
	if err != nil {
		logger.Warn("SecretBox init failed; datasource encryption disabled", "error", err)
		secretBox = nil
	}
	envFallback := datasource.BuildEnvFallback()
	resolver := datasource.NewCredentialResolver(dbpool, secretBox, envFallback, logger)
	if err := resolver.ReloadAll(context.Background()); err != nil {
		logger.Warn("warm-up credentials failed", "error", err)
	}

	// 构造 client（先用默认 URL）
	fredClient := apiclient.NewFredClient("")
	mql5Fetcher := apiclient.NewMQL5Fetcher()

	// 注册 OnChange 回调：key/endpoint 变更时热更新 client
	resolver.OnChange("fred", func(sourceID, secret, endpoint string) {
		fredClient.SetAPIKey(secret)
		fredClient.SetBaseURL(endpoint)
		logger.Info("fred config hot-reloaded", "has_secret", secret != "", "has_endpoint", endpoint != "")
	})
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

	// 启动定时采集
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 构建 jobID → 执行函数 map（供定时器与手动触发共享）
	jobRunners := buildJobRunners(ctx, dbpool, calendarSvc, macroSvc, logger)

	// 数据采集循环 - 启用所有数据源
	go runCollectionLoop(ctx, dbpool, calendarSvc, macroSvc, mql5Fetcher, fredClient, logger)

	// 订阅 jobs:trigger 频道，支持从管理端手动触发任意 job
	go subscribeJobTriggers(ctx, redisClient.Raw(), jobRunners, logger)

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
	logger.Info("Press Ctrl+C to stop.")
	<-sigChan
	logger.Info("Shutting down...")
	cancel()
	time.Sleep(500 * time.Millisecond)
	logger.Info("Worker stopped")
}

// 数据采集循环
func runCollectionLoop(ctx context.Context, dbpool *pgxpool.Pool, calendarSvc *calendar.CalendarService, macroSvc *macro.MacroService, mql5Fetcher *apiclient.MQL5Fetcher, fredClient *apiclient.FredClient, logger *slog.Logger) {
	// 首次运行：立即执行一次全量采集
	logger.Info("Running initial full data collection")
	runAllCollections(ctx, dbpool, calendarSvc, macroSvc, mql5Fetcher, fredClient, logger)

	// 创建定时器
	calendarTicker := time.NewTicker(1 * time.Hour)   // 每小时采集日历
	macroTicker := time.NewTicker(4 * time.Hour)      // 每4小时采集宏观数据
	actualsTicker := time.NewTicker(30 * time.Minute) // 每30分钟更新实际值
	cotTicker := time.NewTicker(6 * time.Hour)        // 每6小时采集COT
	priceTicker := time.NewTicker(6 * time.Hour)      // 每6小时采集价格
	sentimentTicker := time.NewTicker(1 * time.Hour)  // 每小时采集情绪
	onchainTicker := time.NewTicker(1 * time.Hour)    // 每小时采集链上

	defer calendarTicker.Stop()
	defer macroTicker.Stop()
	defer actualsTicker.Stop()
	defer cotTicker.Stop()
	defer priceTicker.Stop()
	defer sentimentTicker.Stop()
	defer onchainTicker.Stop()

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
		}
	}
}

// 执行所有采集任务 - 统一用 runWithEvent 包装以发布实时事件
func runAllCollections(ctx context.Context, dbpool *pgxpool.Pool, calendarSvc *calendar.CalendarService, macroSvc *macro.MacroService, mql5Fetcher *apiclient.MQL5Fetcher, fredClient *apiclient.FredClient, logger *slog.Logger) {
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
		return fmt.Errorf("macro sync inserted 0 records: 检查 FRED API Key 是否已在 /datasources 配置")
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

