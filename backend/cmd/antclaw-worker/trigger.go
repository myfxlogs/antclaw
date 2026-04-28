// 手动触发：通过 Redis Pub/Sub (channel "jobs:trigger") 接收管理端 RunJob 请求，
// 根据 payload 中的 job_id 调用对应 collector，复用 runWithEvent 进行状态汇报。
package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	redisv9 "github.com/redis/go-redis/v9"

	"github.com/antclaw/antclaw/internal/service/calendar"
	"github.com/antclaw/antclaw/internal/service/macro"
)

// buildJobRunners 构建 jobID → runner 函数。每个 runner 内部仍走 runWithEvent，
// 复用启停开关、状态发布、错误捕获逻辑。
func buildJobRunners(
	ctx context.Context,
	dbpool *pgxpool.Pool,
	calendarSvc *calendar.CalendarService,
	macroSvc *macro.MacroService,
	logger *slog.Logger,
) map[string]func() {
	return map[string]func(){
		"calendar-sync": func() {
			runWithEvent(ctx, logger, "calendar-sync", "calendar-sync",
				func() error { return runCalendarCollection(ctx, calendarSvc, logger) })
		},
		"macro-sync": func() {
			runWithEvent(ctx, logger, "macro-sync", "macro-sync",
				func() error { return runMacroCollection(ctx, macroSvc, logger) })
		},
		"actuals-update": func() {
			runWithEvent(ctx, logger, "actuals-update", "actuals-update",
				func() error { return runActualsUpdate(ctx, calendarSvc, logger) })
		},
		"cot-sync": func() {
			runWithEvent(ctx, logger, "cot-sync", "cot-sync",
				func() error { return collectCOT(ctx, dbpool, logger) })
		},
		"price-sync": func() {
			runWithEvent(ctx, logger, "price-sync", "price-sync",
				func() error { return collectPrices(ctx, dbpool, logger) })
		},
		"sentiment-sync": func() {
			runWithEvent(ctx, logger, "sentiment-sync", "sentiment-sync",
				func() error { return collectSentiment(ctx, dbpool, logger) })
		},
		"onchain-sync": func() {
			runWithEvent(ctx, logger, "onchain-sync", "onchain-sync",
				func() error { return collectOnchain(ctx, dbpool, logger) })
		},
		"intraday-sync": func() {
			runWithEvent(ctx, logger, "intraday-sync", "intraday-sync",
				func() error { return collectIntraday(ctx, dbpool, logger) })
		},
		"defi-sync": func() {
			runWithEvent(ctx, logger, "defi-sync", "defi-sync",
				func() error { return collectDefi(ctx, dbpool, logger) })
		},
		"vix-term-sync": func() {
			runWithEvent(ctx, logger, "vix-term-sync", "vix-term-sync",
				func() error { return collectVIXTermStructure(ctx, dbpool, logger) })
		},
		"dvol-sync": func() {
			runWithEvent(ctx, logger, "dvol-sync", "dvol-sync",
				func() error { return collectDVOL(ctx, dbpool, logger) })
		},
		"cot-analysis": func() {
			runWithEvent(ctx, logger, "cot-analysis", "cot-analysis",
				func() error { return analyzeCOT(ctx, dbpool, logger) })
		},
		"macro-regime": func() {
			runWithEvent(ctx, logger, "macro-regime", "macro-regime",
				func() error { return analyzeMacroRegime(ctx, dbpool, logger) })
		},
		"flow-divergence": func() {
			runWithEvent(ctx, logger, "flow-divergence", "flow-divergence",
				func() error { return analyzeFlowDivergence(ctx, dbpool, logger) })
		},
		"volume-profile": func() {
			runWithEvent(ctx, logger, "volume-profile", "volume-profile",
				func() error { return analyzeVolumeProfile(ctx, dbpool, logger) })
		},
		"outcome-evaluator": func() {
			runWithEvent(ctx, logger, "outcome-evaluator", "outcome-evaluator",
				func() error { return evaluateSignalOutcomes(ctx, dbpool, logger) })
		},
		"transition-matrix": func() {
			runWithEvent(ctx, logger, "transition-matrix", "transition-matrix",
				func() error { return buildTransitionMatrix(ctx, dbpool, logger) })
		},
		"alert-evaluator": func() {
			runWithEvent(ctx, logger, "alert-evaluator", "alert-evaluator",
				func() error { return evaluateAlerts(ctx, dbpool, logger) })
		},
	}
}

// subscribeJobTriggers 监听 Redis "jobs:trigger" 频道，根据 payload.job_id 路由到对应 runner。
// 每次触发开启独立 goroutine，避免阻塞订阅。
func subscribeJobTriggers(ctx context.Context, rdb *redisv9.Client, runners map[string]func(), logger *slog.Logger) {
	if rdb == nil {
		logger.Warn("jobs:trigger subscriber disabled: redis is nil")
		return
	}
	pubsub := rdb.Subscribe(ctx, "jobs:trigger")
	defer pubsub.Close()

	logger.Info("jobs:trigger subscriber started")
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var evt struct {
				JobID string `json:"job_id"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				logger.Warn("invalid jobs:trigger payload", "payload", msg.Payload, "error", err)
				continue
			}
			runner, found := runners[evt.JobID]
			if !found {
				logger.Warn("unknown job_id from trigger", "job_id", evt.JobID)
				continue
			}
			logger.Info("manual trigger received", "job_id", evt.JobID)
			go runner()
		}
	}
}
