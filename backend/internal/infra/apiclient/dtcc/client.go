package dtcc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 DTCC PPD 公开 FX 衍生品累计数据 API（无需 Key）。
// 接口示例：https://pddata.dtcc.com/ppd/api/report/cumulative/CFTC/FOREX?asof=YYYY-MM-DD
// 参见 docs/数据采集源密钥清单.md（DTCC 段）。
type Client struct {
	src  apiclient.Source
	base string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{
		src:  src,
		base: "https://pddata.dtcc.com/ppd/api/report/cumulative/CFTC/FOREX",
	}
}

// PairAggregate 单货币对的累计名义量（百万 USD 当量）与笔数。
type PairAggregate struct {
	Pair          string
	TotalNotional float64
	TradeCount    int
	AsOf          time.Time
}

// FetchByDate 拉取某交易日的 FOREX 累计明细，按 currency_pair 聚合。
// pairs 限定要保留的交叉盘；为空时聚合 majors 默认集合。
// 如目标日 404（周末/假期/上游下线），自动向前回退至多 5 个工作日。
func (c *Client) FetchByDate(ctx context.Context, asOf time.Time, pairs []string) ([]*PairAggregate, error) {
	var resp *http.Response
	var err error
	candidate := asOf
	for attempt := 0; attempt < 5; attempt++ {
		url := fmt.Sprintf("%s?asof=%s", c.base, candidate.Format("2006-01-02"))
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		req.Header.Set("Accept", "application/json")
		resp, err = c.src.Do(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			return nil, fmt.Errorf("dtcc http %d", resp.StatusCode)
		}
		candidate = candidate.AddDate(0, 0, -1)
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dtcc no data within 5 business days of %s", asOf.Format("2006-01-02"))
	}
	defer resp.Body.Close()
	asOf = candidate
	var rows []struct {
		Cur1     string `json:"NOTIONAL_CURRENCY_1"`
		Cur2     string `json:"NOTIONAL_CURRENCY_2"`
		Notional string `json:"ROUNDED_NOTIONAL_AMOUNT_1"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("dtcc decode: %w", err)
	}

	allow := normalizePairSet(pairs)
	agg := make(map[string]*PairAggregate, len(allow))
	for _, r := range rows {
		pair := strings.ToUpper(r.Cur1 + r.Cur2)
		if _, ok := allow[pair]; !ok {
			continue
		}
		amt := parseNotional(r.Notional)
		entry, ok := agg[pair]
		if !ok {
			entry = &PairAggregate{Pair: pair, AsOf: asOf}
			agg[pair] = entry
		}
		entry.TotalNotional += amt
		entry.TradeCount++
	}
	out := make([]*PairAggregate, 0, len(agg))
	for _, v := range agg {
		out = append(out, v)
	}
	return out, nil
}

// 默认主要交叉盘
var defaultPairs = []string{
	"EURUSD", "USDJPY", "GBPUSD", "AUDUSD", "USDCAD",
	"USDCHF", "NZDUSD", "USDMXN", "USDBRL", "USDHKD",
}

func normalizePairSet(in []string) map[string]struct{} {
	src := in
	if len(src) == 0 {
		src = defaultPairs
	}
	set := make(map[string]struct{}, len(src))
	for _, p := range src {
		set[strings.ToUpper(strings.TrimSpace(p))] = struct{}{}
	}
	return set
}

// parseNotional 兼容 DTCC 中常见的逗号千分位、>$ 前缀等格式；解析失败按 0 计。
func parseNotional(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, ">$")
	s = strings.TrimPrefix(s, "$")
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
