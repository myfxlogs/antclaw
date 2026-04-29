// Package mql5 是 MQL5 经济日历抓取器的薄包装。
//
// 当前实现是 apiclient.MQL5Fetcher 的别名，目的与 fred 子包一致：
// 统一 apiclient/<vendor>/ 目录约定，便于后续小步迁移到 apiclient.Source 中间件。
package mql5

import "github.com/antclaw/antclaw/internal/infra/apiclient"

// Fetcher 与 apiclient.MQL5Fetcher 等价。
type Fetcher = apiclient.MQL5Fetcher

// NewFetcher 等价于 apiclient.NewMQL5Fetcher。
func NewFetcher() *Fetcher { return apiclient.NewMQL5Fetcher() }
