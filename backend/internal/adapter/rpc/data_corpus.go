// Package rpc：管理端采集数据白名单表聚合与预览查询（AdminData Connect 复用）。
package rpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrPreviewMissingJobID 预览请求缺少 job_id。
	ErrPreviewMissingJobID = errors.New("missing job_id")
	// ErrPreviewUnknownJobID 预览 job_id 不在白名单。
	ErrPreviewUnknownJobID = errors.New("unknown job_id")
)

type tableMapping struct {
	JobID   string
	Name    string
	Table   string
	TimeCol string
	Where   string // 可选：附加 WHERE 条件（不带 WHERE 关键字），用于区分共享表的不同数据源
}

// 采集对象 → 数据库表 + 时间列。仅声明已知映射；查询时优雅降级。
// 一个数据库表对应一个 /data 项目；每张表独立展示。
var dataSourceTables = []tableMapping{
	// === 采集类（Collector） ===
	{JobID: "calendar-sync", Name: "财经日历", Table: "calendar_events", TimeCol: "scheduled_at"},
	{JobID: "calendar-titles", Name: "日历事件标题", Table: "calendar_event_titles", TimeCol: "fetched_at"},
	{JobID: "calendar-surprise", Name: "日历意外指数", Table: "calendar_surprise_history", TimeCol: "released_at"},
	{JobID: "macro-sync", Name: "宏观指标 (FRED)", Table: "data_snapshots", TimeCol: "time", Where: "source = 'fred'"},
	{JobID: "cot-sync", Name: "COT 持仓 (CFTC)", Table: "cot_records", TimeCol: "report_date"},
	{JobID: "price-daily", Name: "日 K 线", Table: "price_daily", TimeCol: "time"},
	{JobID: "price-weekly", Name: "周 K 线", Table: "price_weekly", TimeCol: "week"},
	{JobID: "intraday-sync", Name: "5min K 线 (Yahoo)", Table: "price_intraday", TimeCol: "time"},
	{JobID: "sentiment-sync", Name: "情绪快照 (CNN/AAII)", Table: "sentiment_snapshots", TimeCol: "time"},
	{JobID: "onchain-sync", Name: "链上指标 (CoinGecko)", Table: "onchain_metrics", TimeCol: "time"},
	{JobID: "defi-sync", Name: "DeFi TVL (DefiLlama)", Table: "defi_snapshots", TimeCol: "time"},
	{JobID: "vix-term-sync", Name: "VIX 期限结构 (CBOE)", Table: "vix_term_structure", TimeCol: "time"},
	{JobID: "dvol-sync", Name: "DVOL (Deribit)", Table: "dvol_snapshots", TimeCol: "time"},
	{JobID: "event-impact", Name: "事件影响记录", Table: "event_impact_records", TimeCol: "recorded_at"},
	{JobID: "micro-snapshot", Name: "微观结构快照", Table: "micro_snapshots", TimeCol: "time"},
	{JobID: "data-snapshot", Name: "通用数据快照 (其他来源)", Table: "data_snapshots", TimeCol: "time", Where: "source <> 'fred'"},

	// === 分析衍生类（Analyzer） ===
	{JobID: "cot-analysis", Name: "COT 分析 (Index/Z-score)", Table: "cot_analyses", TimeCol: "report_date"},
	{JobID: "cot-calibration", Name: "COT 校准", Table: "cot_calibration", TimeCol: "updated_at"},
	{JobID: "cot-signal-outcomes", Name: "COT 信号结果", Table: "cot_signal_outcomes", TimeCol: "issued_at"},
	{JobID: "macro-regime", Name: "宏观状态分类", Table: "macro_regime_history", TimeCol: "time"},
	{JobID: "regime-overlay", Name: "状态叠加历史", Table: "regime_overlay_history", TimeCol: "time"},
	{JobID: "regime-transitions", Name: "状态转移记录", Table: "regime_transitions", TimeCol: "time"},
	{JobID: "transition-matrix", Name: "状态转移矩阵", Table: "regime_transition_matrix", TimeCol: "asof_date"},
	{JobID: "flow-divergence", Name: "资金流向背离", Table: "flow_divergence_history", TimeCol: "time"},
	{JobID: "volume-profile", Name: "成交量分布 (POC/VAH/VAL)", Table: "volume_profiles", TimeCol: "time"},
	{JobID: "gex-snapshot", Name: "GEX 快照 (期权伽马)", Table: "gex_snapshots", TimeCol: "time"},
	{JobID: "iv-skew", Name: "IV 偏度历史", Table: "iv_skew_history", TimeCol: "time"},
	{JobID: "wyckoff-events", Name: "Wyckoff 事件", Table: "wyckoff_events", TimeCol: "bar_time"},
	{JobID: "walkforward", Name: "Walk-forward 回测", Table: "walkforward_history", TimeCol: "test_to"},
}

// DataSourceSummary 单个采集对象的统计快照。
type DataSourceSummary struct {
	JobID      string
	Name       string
	Table      string
	Count      int64
	LatestTime int64
	Error      string
}

func quoteIdent(name string) string {
	return `"` + name + `"`
}

func querySummary(ctx context.Context, pool *pgxpool.Pool, m tableMapping) DataSourceSummary {
	out := DataSourceSummary{JobID: m.JobID, Name: m.Name, Table: m.Table}

	whereClause := ""
	if m.Where != "" {
		whereClause = " WHERE " + m.Where
	}

	var count int64
	if err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s%s`, quoteIdent(m.Table), whereClause)).Scan(&count); err != nil {
		out.Error = err.Error()
		return out
	}
	out.Count = count

	if m.TimeCol != "" && count > 0 {
		var latest time.Time
		err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT MAX(%s) FROM %s%s`, quoteIdent(m.TimeCol), quoteIdent(m.Table), whereClause)).Scan(&latest)
		if err == nil {
			out.LatestTime = latest.Unix()
		} else {
			out.Error = err.Error()
		}
	}
	return out
}

// CollectDataSummaries 顺序查询各表汇总，单表失败写入 Error 不中断。
func CollectDataSummaries(ctx context.Context, pool *pgxpool.Pool) []DataSourceSummary {
	out := make([]DataSourceSummary, 0, len(dataSourceTables))
	for _, m := range dataSourceTables {
		out = append(out, querySummary(ctx, pool, m))
	}
	return out
}

func findMapping(jobID string) (*tableMapping, bool) {
	for i := range dataSourceTables {
		if dataSourceTables[i].JobID == jobID {
			return &dataSourceTables[i], true
		}
	}
	return nil, false
}

// DataPreviewRows 预览查询结果。
type DataPreviewRows struct {
	JobID        string
	Table        string
	TimeCol      string
	Columns      []string
	Rows         []map[string]interface{}
	TotalSampled int
}

// FetchDataPreview 拉取最近 limit 条（上限 200），job_id 须在白名单内。
func FetchDataPreview(ctx context.Context, pool *pgxpool.Pool, jobID string, limit int) (*DataPreviewRows, error) {
	if jobID == "" {
		return nil, ErrPreviewMissingJobID
	}
	mapping, ok := findMapping(jobID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPreviewUnknownJobID, jobID)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 兜底：若表不存在（采集器尚未落地建表），直接返回空预览，避免前端 500。
	var exists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
		mapping.Table).Scan(&exists); err == nil && !exists {
		return &DataPreviewRows{
			JobID:   mapping.JobID,
			Table:   mapping.Table,
			TimeCol: mapping.TimeCol,
		}, nil
	}

	orderBy := quoteIdent(mapping.TimeCol)
	if mapping.TimeCol == "" {
		orderBy = "id"
	}

	whereClause := ""
	if mapping.Where != "" {
		whereClause = " WHERE " + mapping.Where
	}

	sql := fmt.Sprintf(`SELECT * FROM %s%s ORDER BY %s DESC NULLS LAST LIMIT $1`,
		quoteIdent(mapping.Table), whereClause, orderBy)

	rows, err := pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = f.Name
	}

	items, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, fmt.Errorf("collect rows failed: %w", err)
	}

	sanitizePreviewRows(items)

	return &DataPreviewRows{
		JobID:        mapping.JobID,
		Table:        mapping.Table,
		TimeCol:      mapping.TimeCol,
		Columns:      columns,
		Rows:         items,
		TotalSampled: len(items),
	}, nil
}

func sanitizePreviewRows(rows []map[string]interface{}) {
	for _, row := range rows {
		for k, v := range row {
			switch val := v.(type) {
			case []byte:
				hex := fmt.Sprintf("%x", val)
				if len(hex) > 64 {
					hex = hex[:64] + "..."
				}
				row[k] = hex
			case time.Time:
				row[k] = val.Format(time.RFC3339)
			case string:
				if len(val) > 1024 {
					row[k] = val[:1024] + "..."
				}
			}
		}
	}
}
