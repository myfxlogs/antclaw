package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ensureDataSourceSeeds 写入业务里已知数据源的占位记录。
// 不写入 secret_*；初始 has_secret=false，前端可见但状态为"未配置"。
// 数据源清单参考：docs/数据采集源密钥清单.md
func ensureDataSourceSeeds(ctx context.Context, pool *pgxpool.Pool) error {
	type seed struct{ id, name, kind, endpoint string }
	seeds := []seed{
		// ===== 必需 API Key 的数据源 =====
		{"twelve_data", "TwelveData (行情主力)", "api_key", "https://api.twelvedata.com"},
		{"alpha_vantage", "Alpha Vantage (行情备援)", "api_key", "https://www.alphavantage.co"},
		{"fred", "FRED 宏观数据", "api_key", "https://api.stlouisfed.org/fred"},
		{"eia", "EIA 能源数据", "api_key", "https://api.eia.gov"},
		{"coingecko", "CoinGecko (Demo Key)", "api_key", "https://api.coingecko.com"},
		{"firecrawl", "Firecrawl 网页爬取", "api_key", "https://api.firecrawl.dev"},
		{"massive", "Massive (历史面板)", "api_key", "https://api.massive.com"},
		{"cftc_socrata", "CFTC Socrata COT", "api_key", "https://publicreporting.cftc.gov"},
		{"bybit", "Bybit 私有 API", "api_key", "https://api.bybit.com"},

		// ===== 公共数据源（无需 Key） =====
		// 持仓与衍生品
		{"cftc_legacy", "CFTC Legacy CSV", "endpoint", "https://www.cftc.gov/dea/newcot"},
		{"dtcc_ppd", "DTCC PPD (FX 衍生)", "endpoint", "https://pddata.dtcc.com"},

		// 经济/政府公开 API
		{"mql5", "MQL5 财经日历", "endpoint", "https://www.mql5.com/en/economic-calendar/content"},
		{"worldbank", "World Bank 数据", "endpoint", "https://api.worldbank.org"},
		{"ecb", "ECB 欧洲央行", "endpoint", "https://data-api.ecb.europa.eu"},
		{"eurostat", "Eurostat 欧盟统计", "endpoint", "https://ec.europa.eu/eurostat"},
		{"oecd", "OECD CLI 指数", "endpoint", "https://sdmx.oecd.org"},
		{"snb", "SNB 瑞士央行", "endpoint", "https://data.snb.ch"},
		{"us_treasury", "US Treasury 收益率", "endpoint", "https://home.treasury.gov"},
		{"treasury_direct", "TreasuryDirect 拍卖", "endpoint", "https://www.treasurydirect.gov"},
		{"bis", "BIS 国际清算银行", "endpoint", "https://stats.bis.org"},
		{"imf", "IMF DataMapper", "endpoint", "https://www.imf.org/external/datamapper"},
		{"fed_rss", "Federal Reserve RSS", "endpoint", "https://www.federalreserve.gov/feeds"},
		{"sec_edgar", "SEC EDGAR 13F", "endpoint", "https://data.sec.gov"},

		// 价格/量化指标（公共行情）
		{"yahoo", "Yahoo Finance", "endpoint", "https://query1.finance.yahoo.com"},
		{"stooq", "Stooq 价格备援", "endpoint", "https://stooq.com"},
		{"cboe", "CBOE 波动率指数", "endpoint", "https://cdn.cboe.com"},
		{"coinmetrics", "Coin Metrics Community", "endpoint", "https://community-api.coinmetrics.io"},
		{"blockchain_info", "Blockchain.info", "endpoint", "https://api.blockchain.info"},
		{"defillama", "DefiLlama TVL", "endpoint", "https://api.llama.fi"},
		{"cryptocompare", "CryptoCompare", "endpoint", "https://min-api.cryptocompare.com"},
		{"deribit", "Deribit DVOL", "endpoint", "https://www.deribit.com"},

		// 情绪/风险（公共）
		{"cnn_fear_greed", "CNN Fear & Greed", "endpoint", "https://production.dataviz.cnn.io"},
		{"alternative_me", "Alternative.me 情绪", "endpoint", "https://api.alternative.me"},
	}
	for _, s := range seeds {
		_, err := pool.Exec(ctx, `
			INSERT INTO data_source_configs (source_id, name, kind, endpoint)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (source_id) DO NOTHING
		`, s.id, s.name, s.kind, s.endpoint)
		if err != nil {
			return err
		}
	}
	// 清理已迁移到 system_ai_configs 或不再属于数据源的过时记录
	if _, err := pool.Exec(ctx, `
		DELETE FROM data_source_configs
		WHERE source_id IN ('gemini', 'claude', 'telegram')
	`); err != nil {
		return err
	}
	return nil
}

// ensureSystemAIConfigSeeds 写入已知 AI Provider 占位记录。
func ensureSystemAIConfigSeeds(ctx context.Context, pool *pgxpool.Pool) error {
	type aiSeed struct {
		id, name     string
		baseURL      string
		models       []string
		defaultModel string
		purposes     []string
		docsURL      string
		applyURL     string
	}
	seeds := []aiSeed{
		{"openai", "OpenAI", "https://api.openai.com/v1", []string{}, "", []string{"chat", "embedding", "summarizer"},
			"https://platform.openai.com/docs/api-reference", "https://platform.openai.com/api-keys"},
		{"openai_compatible", "OpenAI Compatible", "", []string{}, "", []string{"chat", "reasoning", "summarizer"},
			"https://platform.openai.com/docs/api-reference", ""},
		{"anthropic", "Anthropic Claude", "https://api.anthropic.com/v1", []string{}, "", []string{"chat", "reasoning"},
			"https://docs.anthropic.com/en/api", "https://console.anthropic.com/settings/keys"},
		// Gemini 默认使用官方 OpenAI 兼容端点 /v1beta/openai。
		{"gemini", "Google Gemini", "https://generativelanguage.googleapis.com/v1beta/openai", []string{}, "", []string{"chat", "reasoning"},
			"https://ai.google.dev/gemini-api/docs/openai", "https://aistudio.google.com/apikey"},
		{"deepseek", "DeepSeek", "https://api.deepseek.com/v1", []string{}, "", []string{"chat", "reasoning"},
			"https://api-docs.deepseek.com", "https://platform.deepseek.com/api_keys"},
		{"moonshot", "月之暗面 Kimi", "https://api.moonshot.cn/v1", []string{}, "", []string{"chat", "summarizer"},
			"https://platform.moonshot.cn/docs", "https://platform.moonshot.cn/console/api-keys"},
		{"qwen", "通义千问", "https://dashscope.aliyuncs.com/compatible-mode/v1", []string{}, "", []string{"chat", "embedding"},
			"https://help.aliyun.com/zh/model-studio/developer-reference/use-qwen-by-calling-api", "https://bailian.console.aliyun.com/?apiKey=1"},
		{"zhipu", "智谱 GLM", "https://open.bigmodel.cn/api/paas/v4", []string{}, "", []string{"chat"},
			"https://www.bigmodel.cn/dev/api", "https://www.bigmodel.cn/usercenter/apikeys"},
	}
	for _, s := range seeds {
		_, err := pool.Exec(ctx, `
			INSERT INTO system_ai_configs (provider_id, name, base_url, models, default_model, purposes, enabled, docs_url, apply_url)
			VALUES ($1, $2, $3, $4, $5, $6, FALSE, $7, $8)
			ON CONFLICT (provider_id) DO NOTHING
		`, s.id, s.name, s.baseURL, s.models, s.defaultModel, s.purposes, s.docsURL, s.applyURL)
		if err != nil {
			return err
		}
		// 幂等补齐 docs_url / apply_url（只在当前为空时写入，不覆盖运维手改）
		if _, err := pool.Exec(ctx, `
			UPDATE system_ai_configs
			   SET docs_url  = COALESCE(NULLIF(docs_url, ''),  $2),
			       apply_url = COALESCE(NULLIF(apply_url, ''), $3)
			 WHERE provider_id = $1`, s.id, s.docsURL, s.applyURL); err != nil {
			return err
		}
		if s.baseURL != "" {
			if _, err := pool.Exec(ctx, `
				UPDATE system_ai_configs
				SET base_url = $2
				WHERE provider_id = $1
				  AND COALESCE(base_url, '') = ''
			`, s.id, s.baseURL); err != nil {
				return err
			}
		}
		// 清空模板中的硬编码模型，用户通过 Base URL 自动发现模型列表。
		if _, err := pool.Exec(ctx, `
			UPDATE system_ai_configs
			SET models = '{}'::text[], default_model = ''
			WHERE provider_id = $1
			  AND COALESCE(base_url, '') = ''
		`, s.id); err != nil {
			return err
		}
	}
	return nil
}

// ensureStrategySeeds 写入预设策略模板（默认停用，需管理员手动启用）。
func ensureStrategySeeds(ctx context.Context, pool *pgxpool.Pool) error {
	type strategySeed struct {
		name        string
		kind        string
		symbol      string
		timeframe   string
		params      string
		schedule    string
		description string
	}
	seeds := []strategySeed{
		{name: "量化-动量突破", kind: "quant_momentum", symbol: "EURUSD", timeframe: "1d", params: `{"lookback":20,"threshold":0.03}`, schedule: "@hourly", description: "量化信号策略1：20日动量突破。"},
		{name: "量化-趋势跟随", kind: "quant_trend", symbol: "GBPUSD", timeframe: "4h", params: `{"ema_fast":21,"ema_slow":55}`, schedule: "@hourly", description: "量化信号策略2：EMA 趋势跟随。"},
		{name: "量化-均值回归", kind: "quant_mean_reversion", symbol: "EURUSD", timeframe: "1h", params: `{"zscore_window":48,"entry_z":2.0,"exit_z":0.5}`, schedule: "@daily", description: "量化信号策略3：价差 zscore 回归。"},
		{name: "量化-利差Carry", kind: "quant_carry", symbol: "AUDJPY", timeframe: "1d", params: `{"carry_weight":0.7,"vol_target":0.1}`, schedule: "@daily", description: "量化信号策略4：Carry + 波动目标。"},
		{name: "量化-多因子合成", kind: "quant_multi_factor", symbol: "USDJPY", timeframe: "1d", params: `{"momentum":0.3,"trend":0.25,"carry":0.2,"residual":0.25}`, schedule: "@daily", description: "量化信号策略5：多因子加权合成。"},
		{name: "VP-轮廓旋转", kind: "vp_rotation", symbol: "EURUSD", timeframe: "1h", params: `{"value_area_width":0.7,"rotation_bars":12}`, schedule: "@hourly", description: "VP回测变体1：价值区轮廓旋转。"},
		{name: "VP-POC突破", kind: "vp_breakout", symbol: "GBPUSD", timeframe: "1h", params: `{"poc_break_threshold":0.0015}`, schedule: "@hourly", description: "VP回测变体2：POC 突破跟随。"},
		{name: "VP-吸收反转", kind: "vp_absorption_reversal", symbol: "XAUUSD", timeframe: "30m", params: `{"absorption_ratio":1.8,"reversal_bars":6}`, schedule: "@hourly", description: "VP回测变体3：吸收信号反转。"},
		{name: "CTA-Dual EMA", kind: "cta_dual_ema", symbol: "EURUSD", timeframe: "4h", params: `{"fast":20,"slow":60}`, schedule: "@hourly", description: "CTA策略1：双均线交叉。"},
		{name: "CTA-Donchian Breakout", kind: "cta_donchian_breakout", symbol: "USDJPY", timeframe: "1d", params: `{"channel":55}`, schedule: "@daily", description: "CTA策略2：唐奇安通道突破。"},
		{name: "CTA-ATR Trail", kind: "cta_atr_trail", symbol: "XAUUSD", timeframe: "4h", params: `{"atr_period":14,"atr_mult":2.5}`, schedule: "@hourly", description: "CTA策略3：ATR 跟踪止损。"},
		{name: "CTA-Multi TF", kind: "cta_multi_tf", symbol: "BTCUSDT", timeframe: "1h", params: `{"primary_tf":"1h","confirm_tf":"4h","target_vol":0.12}`, schedule: "@hourly", description: "CTA策略4：多周期共振。"},
	}
	for _, s := range seeds {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM strategies WHERE name = $1 AND kind = $2 AND deleted_at IS NULL)
		`, s.name, s.kind).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO strategies (name, kind, symbol, timeframe, params, schedule_cron, enabled, status, description, created_at, updated_at, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6, FALSE, 'draft', $7, NOW(), NOW(), 'system', 'system')
		`, s.name, s.kind, s.symbol, s.timeframe, s.params, s.schedule, s.description)
		if err != nil {
			return err
		}
	}
	return nil
}
