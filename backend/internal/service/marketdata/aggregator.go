// Package marketdata 提供多源价格回退聚合。
//
// 设计：上层只关心一组统一的 Bar；底层 vendor 通过 Source 接口对接，
// Aggregator 按声明顺序依次尝试，**返回首个非空结果**并记录命中 vendor，
// 便于审计与告警。
//
// 当前生产场景中各 vendor 已分别由对应包实现 FetchOHLC；本包只做适配与编排。
package marketdata

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Bar 与 quant 模块兼容的 OHLCV 表示。
type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// Source 单一 vendor 的统一接口。
type Source interface {
	Name() string
	Available() bool
	FetchOHLC(ctx context.Context, symbol, timeframe string, count int) ([]Bar, error)
}

// Aggregator 多源回退聚合器。
type Aggregator struct {
	sources []Source
	minBars int
}

// NewAggregator sources 顺序即优先级；minBars=0 表示 1。
func NewAggregator(minBars int, sources ...Source) *Aggregator {
	if minBars <= 0 {
		minBars = 1
	}
	return &Aggregator{sources: sources, minBars: minBars}
}

// FetchOHLC 依次询问每个可用 source；首个返回 ≥ minBars 根 K 线者胜出。
// 返回值：bars / 命中的 vendor / 各 vendor 错误集合。
func (a *Aggregator) FetchOHLC(ctx context.Context, symbol, timeframe string, count int) ([]Bar, string, error) {
	if len(a.sources) == 0 {
		return nil, "", errors.New("marketdata: no sources configured")
	}
	var lastErr error
	var failures []string
	for _, s := range a.sources {
		if !s.Available() {
			failures = append(failures, fmt.Sprintf("%s: unavailable", s.Name()))
			continue
		}
		bars, err := s.FetchOHLC(ctx, symbol, timeframe, count)
		if err != nil {
			lastErr = err
			failures = append(failures, fmt.Sprintf("%s: %v", s.Name(), err))
			continue
		}
		if len(bars) >= a.minBars {
			return bars, s.Name(), nil
		}
		failures = append(failures, fmt.Sprintf("%s: only %d bars (need %d)", s.Name(), len(bars), a.minBars))
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("marketdata: all sources failed: %v", failures)
	}
	return nil, "", lastErr
}
