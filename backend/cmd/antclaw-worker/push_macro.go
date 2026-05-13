// 客户端智能推送：宏观 / 期权 / 链上 / 状态迁移风险检测。
//
// 使用 push_util.go 中的 scanUsers / sendIfNotPushed / alertPrefs。
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/antclaw/antclaw/internal/notify"
)

// ---------- 宏观 regime + FRED 异动（每 1 小时） ----------

func pushMacro(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		// 提前查询全局数据，避免对每个用户重复查询
		macroRegime := e.queryMacroRegime(ctx)
		fredAnomalies := e.queryFredAnomalies(ctx)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.MacroAlerts {
				return false
			}
			if macroRegime != "" {
				dateStr := time.Now().Format("2006-01-02")
				ek := fmt.Sprintf("macro:regime:%s:%s", macroRegime, dateStr)
				if e.sendIfNotPushed(ctx, uid, ek, "macro_regime", &notify.Notification{
					UserID: uid, Category: "alert",
					Title: "宏观 regime 更新", Body: macroRegime, Severity: "normal",
					Data:     map[string]string{"kind": "macro_regime", "regime": macroRegime},
					DedupTTL: time.Hour,
				}) {
					sent++
				}
			}
			for _, a := range fredAnomalies {
				if e.sendIfNotPushed(ctx, uid, a.key, "macro_fred", a.notif(uid)) {
					sent++
				}
			}
			return false
		})

		e.log.Info("macro: scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

type fredAnomaly struct {
	key  string
	notif func(uuid.UUID) *notify.Notification
}

func (e *pushEnv) queryMacroRegime(ctx context.Context) string {
	var regime string
	var score *float64
	var ts time.Time
	if err := e.pool.QueryRow(ctx,
		`SELECT regime, score, time FROM macro_regime_history ORDER BY time DESC LIMIT 1`).Scan(&regime, &score, &ts); err != nil {
		return ""
	}
	return fmt.Sprintf("当前宏观 regime：%s（%.2f），更新时间：%s", regime, *score, ts.Format(time.RFC3339))
}

func (e *pushEnv) queryFredAnomalies(ctx context.Context) []fredAnomaly {
	rows, err := e.pool.Query(ctx,
		`SELECT series_id, daily_value FROM fred_daily_agg WHERE day >= NOW() - INTERVAL '3 days' ORDER BY series_id, day DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	latest := map[string]float64{}
	prev := map[string]float64{}
	for rows.Next() {
		var s string
		var v float64
		if err := rows.Scan(&s, &v); err != nil {
			continue
		}
		if _, ok := latest[s]; !ok {
			latest[s] = v
		} else if _, ok := prev[s]; !ok {
			prev[s] = v
		}
	}

	watch := []string{"T10Y2Y", "T10YIE", "DGS10", "FEDFUNDS"}
	var out []fredAnomaly
	for _, series := range watch {
		l, ok1 := latest[series]
		p, ok2 := prev[series]
		if !ok1 || !ok2 || p == 0 {
			continue
		}
		pct := (l - p) / p * 100
		if pct > -2 && pct < 2 {
			continue
		}
		dir := "上升"
		if pct < 0 {
			dir = "下降"
		}
		dateStr := time.Now().Format("2006-01-02")
		lv := l
		pv := pct
		seriesCopy := series
		dirCopy := dir
		out = append(out, fredAnomaly{
			key: fmt.Sprintf("macro:%s:%s:%s", series, dir, dateStr),
			notif: func(uid uuid.UUID) *notify.Notification {
				return &notify.Notification{
					UserID: uid, Category: "alert",
					Title:    fmt.Sprintf("FRED %s 出现异动", seriesCopy),
					Body:     fmt.Sprintf("%s 较前值 %s %.1f%%，当前值：%.4f", seriesCopy, dirCopy, pv, lv),
					Severity: "normal",
					Data:     map[string]string{"kind": "macro_fred", "series": seriesCopy, "direction": dirCopy, "pct_change": fmt.Sprintf("%.1f", pv)},
					DedupTTL: time.Hour,
				}
			},
		})
	}
	return out
}

// ---------- 期权风险（每 1 小时） ----------

func pushOptions(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		// 提前查询高风险品种
		stressed := e.queryStressedSymbols(ctx)
		if len(stressed) == 0 {
			return
		}

		a := stressed[0]
		dateStr := time.Now().Format("2006-01-02")
		ek := fmt.Sprintf("options:stress:%s:%s", a.symbol, dateStr)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.OptionsAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "options_risk", &notify.Notification{
				UserID: uid, Category: "alert",
				Title:    fmt.Sprintf("%s 期权压力偏高", a.symbol),
				Body:     fmt.Sprintf("%s stress_score=%.2f（最新快照），注意期权风险。", a.symbol, a.score),
				Severity: "high",
				Data:     map[string]string{"kind": "options_risk", "symbol": a.symbol, "score": fmt.Sprintf("%.2f", a.score)},
				DedupTTL: time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("options: scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

type stressed struct{ symbol string; score float64 }

func (e *pushEnv) queryStressedSymbols(ctx context.Context) []stressed {
	rows, _ := e.pool.Query(ctx,
		`SELECT symbol, stress_score FROM micro_snapshots WHERE time >= NOW() - INTERVAL '2 hours' AND stress_score > 0.7 ORDER BY stress_score DESC LIMIT 5`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var out []stressed
	for rows.Next() {
		var s stressed
		if rows.Scan(&s.symbol, &s.score) == nil {
			out = append(out, s)
		}
	}
	return out
}

// ---------- 链上风险（每 1 小时） ----------

func pushOnchain(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		var pairA, pairB string
		var z float64
		if err := e.pool.QueryRow(ctx,
			`SELECT pair_a, pair_b, z_score FROM intermarket_correlations WHERE time >= NOW() - INTERVAL '4 hours' AND is_break = TRUE ORDER BY ABS(z_score) DESC LIMIT 1`).Scan(&pairA, &pairB, &z); err != nil {
			return
		}

		dateStr := time.Now().Format("2006-01-02")
		ek := fmt.Sprintf("onchain:correlation:%s:%s", pairA+pairB, dateStr)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.OnchainAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "onchain_risk", &notify.Notification{
				UserID: uid, Category: "alert",
				Title:    "链上相关性断裂",
				Body:     fmt.Sprintf("%s/%s 相关性断裂（z=%.1f），注意链上风险传导。", pairA, pairB, z),
				Severity: "normal",
				Data:     map[string]string{"kind": "onchain_risk", "pair_a": pairA, "pair_b": pairB, "z_score": fmt.Sprintf("%.1f", z)},
				DedupTTL: time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("onchain: scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

// ---------- Carry unwind（每 4 小时） ----------

func pushCarry(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		risky := e.queryRiskyCarry(ctx)
		if len(risky) == 0 {
			return
		}
		dateStr := time.Now().Format("2006-01-02")
		ek := fmt.Sprintf("carry:unwind:%s", dateStr)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.MacroAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "carry_unwind", &notify.Notification{
				UserID: uid, Category: "alert",
				Title:    "Carry Trade 风险提示",
				Body:     fmt.Sprintf("利差异常货币：%s。注意 carry unwind 风险。", strings.Join(risky, ", ")),
				Severity: "high",
				Data:     map[string]string{"kind": "carry_unwind", "currencies": strings.Join(risky, ",")},
				DedupTTL: 4 * time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("carry: scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

func (e *pushEnv) queryRiskyCarry(ctx context.Context) []string {
	rows, _ := e.pool.Query(ctx,
		`SELECT currency, vs_usd_spread FROM carry_rates WHERE date = (SELECT MAX(date) FROM carry_rates) AND ABS(vs_usd_spread) > 3.0 ORDER BY ABS(vs_usd_spread) DESC LIMIT 5`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var cur string
		var sp float64
		if rows.Scan(&cur, &sp) == nil {
			out = append(out, fmt.Sprintf("%s(%.1f%%)", cur, sp))
		}
	}
	return out
}

// ---------- Regime transition（每 4 小时） ----------

func pushRegimeTransition(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		transitions := e.queryRecentTransitions(ctx)
		if len(transitions) == 0 {
			return
		}

		sent := 0
		dateStr := time.Now().Format("2006-01-02")
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.MacroAlerts {
				return false
			}
			for _, t := range transitions {
				ek := fmt.Sprintf("regime:%s:%s:%s:%s", t.symbol, t.from, t.to, dateStr)
				if e.sendIfNotPushed(ctx, uid, ek, "regime_transition", &notify.Notification{
					UserID: uid, Category: "alert",
					Title:    fmt.Sprintf("%s 市场状态发生切换", t.symbol),
					Body:     fmt.Sprintf("%s regime 从 %s 切换为 %s。", t.symbol, t.from, t.to),
					Severity: t.severity,
					Data:     map[string]string{"kind": "regime_transition", "symbol": t.symbol, "from_state": t.from, "to_state": t.to},
					DedupTTL: time.Hour,
				}) {
					sent++
				}
			}
			return false
		})
		e.log.Info("regime: transition scan complete", "sent", sent, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

type regTransition struct{ symbol, from, to, severity string }

func (e *pushEnv) queryRecentTransitions(ctx context.Context) []regTransition {
	rows, _ := e.pool.Query(ctx,
		`SELECT symbol, from_label, to_label, severity FROM regime_transitions WHERE time >= NOW() - INTERVAL '4 hours' ORDER BY time DESC LIMIT 5`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var out []regTransition
	for rows.Next() {
		var t regTransition
		if rows.Scan(&t.symbol, &t.from, &t.to, &t.severity) == nil {
			sev := "normal"
			if t.severity == "CRITICAL" {
				sev = "critical"
			} else if t.severity == "WARN" {
				sev = "high"
			}
			t.severity = sev
			out = append(out, t)
		}
	}
	return out
}

// ---------- 多资产风险共振（每 4 小时） ----------

func pushRiskConfluence(env *pushEnv) func(context.Context, *pushEnv) {
	return func(ctx context.Context, e *pushEnv) {
		started := time.Now()
		count, details := e.assessRiskConfluence(ctx)
		if count < 2 {
			e.log.Debug("risk confluence: below threshold", "count", count)
			return
		}
		dateStr := time.Now().Format("2006-01-02")
		hb := time.Now().Hour() / 4
		ek := fmt.Sprintf("risk_confluence:%s:%d", dateStr, hb)

		sent := 0
		e.scanUsers(ctx, func(uid uuid.UUID, p alertPrefs) bool {
			if !p.MacroAlerts {
				return false
			}
			if e.sendIfNotPushed(ctx, uid, ek, "risk_confluence", &notify.Notification{
				UserID: uid, Category: "signal",
				Title:    "多资产风险共振",
				Body:     fmt.Sprintf("当前 %d 类风险同时触发：%s。建议关注整体市场风险。", count, strings.Join(details, "、")),
				Severity: "high",
				Data:     map[string]string{"kind": "risk_confluence", "risk_count": fmt.Sprint(count), "details": strings.Join(details, ",")},
				DedupTTL: 4 * time.Hour,
			}) {
				sent++
			}
			return false
		})
		e.log.Info("risk: confluence scan complete", "sent", sent, "count", count, "elapsed", time.Since(started).Round(time.Millisecond))
	}
}

func (e *pushEnv) assessRiskConfluence(ctx context.Context) (int, []string) {
	var risk int
	var details []string

	var m string
	if e.pool.QueryRow(ctx, `SELECT regime FROM macro_regime_history WHERE time >= NOW() - INTERVAL '1 hour' AND regime IN ('STAGFLATION','STRESS') ORDER BY time DESC LIMIT 1`).Scan(&m) == nil {
		risk++
		details = append(details, "宏观:"+m)
	}
	var ss float64
	if e.pool.QueryRow(ctx, `SELECT MAX(stress_score) FROM micro_snapshots WHERE time >= NOW() - INTERVAL '1 hour'`).Scan(&ss) == nil && ss > 0.7 {
		risk++
		details = append(details, fmt.Sprintf("期权压力:%.2f", ss))
	}
	var bc int
	if e.pool.QueryRow(ctx, `SELECT COUNT(*) FROM intermarket_correlations WHERE time >= NOW() - INTERVAL '4 hours' AND is_break = TRUE`).Scan(&bc) == nil && bc > 0 {
		risk++
		details = append(details, fmt.Sprintf("相关性断裂:%d对", bc))
	}
	return risk, details
}
