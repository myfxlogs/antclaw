// 客户端智能推送：COT / 信号 / 校准状态。
//
// 使用 push_util.go 中的 scanUsers / sendIfNotPushed / latestCOTReportDate。
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- COT 新数据发布（每 6 小时） ----------

func pushCOTRelease(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		reportDate, err := e.latestCOTReportDate(ctx)
		if err != nil {
			e.log.Debug("COT: no analysis data yet")
			return
		}
		if time.Since(reportDate) > 6*time.Hour {
			e.log.Debug("COT: no new release", "report_date", reportDate.Format("2006-01-02"))
			return
		}

		dateStr := reportDate.Format("2006-01-02")
		ek := fmt.Sprintf("cot:release:%s", dateStr)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.CotAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "cot_release", &notify.Notification{
				UserID: uid, Category: "alert",
				Title:    "COT 新数据已更新",
				Body:     fmt.Sprintf("最新报告日期：%s。系统已完成持仓分析与信号刷新。", dateStr),
				Severity: "normal",
				Data:     map[string]string{"kind": "cot_release", "report_date": dateStr},
				DedupTTL: 7 * 24 * time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("COT: release scan complete", "sent", sent, "report_date", dateStr, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

// ---------- COT 强信号（每 6 小时） ----------

func pushCOTSignal(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		signals, reportDate, ok := e.queryCOTSignals(ctx)
		if !ok {
			return
		}
		dateStr := reportDate.Format("2006-01-02")

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.CotAlerts {
				return false
			}
			for _, s := range signals {
				ek := fmt.Sprintf("cot:signal:%s:%s:%s", s.contract, s.biasType, dateStr)
				if e.sendIfNotPushed(ctx, uid, ek, "cot_signal", &notify.Notification{
					UserID: uid, Category: "signal",
					Title:    fmt.Sprintf("COT %s 强%s信号", s.contract, s.direction),
					Body:     fmt.Sprintf("%s commercial=%s noncomm=%s signal=%.2f。", s.contract, s.comm, s.nonc, s.strength),
					Severity: s.severity,
					Data:     map[string]string{"kind": "cot_signal", "contract": s.contract, "direction": s.direction, "signal_strength": fmt.Sprintf("%.2f", s.strength), "report_date": dateStr},
					DedupTTL: 7 * 24 * time.Hour,
				}) {
					sent++
				}
			}
			return false
		})
		e.log.Info("COT: signal scan complete", "signals", len(signals), "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

type cotSig struct{ contract, comm, nonc, direction, biasType, severity string; strength float64 }

func (e *pushEnv) queryCOTSignals(ctx context.Context) ([]cotSig, time.Time, bool) {
	rows, err := e.pool.Query(ctx, `
		SELECT contract_code, commercial_bias, noncomm_bias, signal_strength
		  FROM cot_analyses
		 WHERE report_date = (SELECT MAX(report_date) FROM cot_analyses)
		   AND signal_strength IS NOT NULL
		   AND ABS(signal_strength) >= 0.7
		 ORDER BY ABS(signal_strength) DESC
		 LIMIT 10`)
	if err != nil {
		return nil, time.Time{}, false
	}
	defer rows.Close()

	var reportDate time.Time
	_ = e.pool.QueryRow(ctx, `SELECT MAX(report_date) FROM cot_analyses`).Scan(&reportDate)

	var out []cotSig
	for rows.Next() {
		var contract, comm, nonc string
		var strength float64
		if err := rows.Scan(&contract, &comm, &nonc, &strength); err != nil {
			continue
		}
		dir := "看涨"
		if strength < 0 {
			dir = "看跌"
		}
		bias := ""
		sev := "normal"
		if comm == "EXTREME_LONG" || nonc == "EXTREME_SHORT" {
			bias = "极端"
			sev = "high"
		}
		out = append(out, cotSig{contract, comm, nonc, dir, bias, sev, strength})
	}
	return out, reportDate, len(out) > 0
}

// ---------- Walk-forward / Calibration（每 6 小时） ----------

func pushCalibration(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		// 检查最近 24h 内是否有新的校准
		rows, _ := e.pool.Query(ctx,
			`SELECT type, brier, n_samples, fitted_at FROM signal_calibrations WHERE fitted_at >= NOW() - INTERVAL '24 hours' ORDER BY fitted_at DESC LIMIT 5`)
		if rows == nil {
			return
		}
		rows.Close()

		dateStr := time.Now().Format("2006-01-02")
		ek := fmt.Sprintf("calibration:update:%s", dateStr)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.CotAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "calibration", &notify.Notification{
				UserID: uid, Category: "digest",
				Title:    "策略校准已更新",
				Body:     "信号校准模型已完成最新拟合，回测质量监控结果可查看。",
				Severity: "low",
				Data:     map[string]string{"kind": "calibration_update", "date": dateStr},
				DedupTTL: 12 * time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("calibration: scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}
