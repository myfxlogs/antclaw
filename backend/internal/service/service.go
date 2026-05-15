// Package service implements business services for each aggregate.
// Services are split by aggregate as per domain design.
// See: AntClaw-领域模型.md
package service

// ClampPageSize returns ps clamped to [defaultVal, maxVal]. Zero or negative uses defaultVal.
func ClampPageSize(ps, defaultVal, maxVal int32) int32 {
	if ps <= 0 {
		return defaultVal
	}
	if ps > maxVal {
		return maxVal
	}
	return ps
}
