// Package fred 是对 St. Louis Fed FRED API 的薄包装。
//
// 当前实现仅是平铺老 client（apiclient.FredClient）的子包别名，
// 目的：保留同等功能与稳定性的同时统一目录结构 apiclient/<vendor>/。
// 后续如需改造（接入 apiclient.Source 中间件、限流、断路器），
// 在本包内推进，不影响调用方导入路径。
package fred

import "github.com/antclaw/antclaw/internal/infra/apiclient"

// Client 与 apiclient.FredClient 等价。
type Client = apiclient.FredClient

// Observation 与 apiclient.FredObservation 等价。
type Observation = apiclient.FredObservation

// NewClient 等价于 apiclient.NewFredClient。
func NewClient(apiKey string) *Client { return apiclient.NewFredClient(apiKey) }
