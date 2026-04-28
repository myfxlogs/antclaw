package apiclient

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	red "github.com/antclaw/antclaw/internal/infra/redis"
)

// Source is a unified adapter for vendor HTTP calls with basic middleware.
type Source interface {
	Vendor() string
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	Healthz(ctx context.Context) HealthSnapshot
}

type Options struct {
	Timeout   time.Duration
	HTTPClient *http.Client
	// 可选：接入 Redis 限流与断路器
	RedisClient     *red.Client
	RateLimitPerMin int           // 每分钟请求上限（0 表示不启用）
	BreakerFailures int           // 连续失败阈值（0 表示不启用）
	BreakerReset    time.Duration // 熔断恢复窗口
}

type HealthSnapshot struct {
	Vendor        string
	LastSuccessAt time.Time
	LastError     string
	RequestsTotal int64
	ErrorsTotal   int64
}

type defaultSource struct {
	vendor string
	client *http.Client
	timeout time.Duration

	rl        *red.RateLimiter
	cb        *red.CircuitBreaker
	rateLimit int

	lastSuccess atomic.Value // time.Time
	lastError   atomic.Value // string
	requests    atomic.Int64
	errors      atomic.Int64
}

// NewSource creates a default Source with minimal middleware (timeout + counters).
func NewSource(vendor string, opts Options) Source {
	t := opts.Timeout
	if t == 0 { t = 15 * time.Second }
	c := opts.HTTPClient
	if c == nil {
		// 独立 client 持有 Timeout，覆盖整个请求生命周期（含 body 读取）
		c = &http.Client{Timeout: t}
	}
	d := &defaultSource{vendor: vendor, client: c, timeout: t}
	if opts.RedisClient != nil {
		if opts.RateLimitPerMin > 0 {
			d.rl = red.NewRateLimiter(opts.RedisClient)
			d.rateLimit = opts.RateLimitPerMin
		}
		if opts.BreakerFailures > 0 {
			reset := opts.BreakerReset
			if reset == 0 { reset = 60 * time.Second }
			d.cb = red.NewCircuitBreaker(opts.RedisClient, "apiclient:"+vendor, opts.BreakerFailures, reset)
		}
	}
	return d
}

func (s *defaultSource) Vendor() string { return s.vendor }

func (s *defaultSource) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	s.requests.Add(1)
	key := s.vendor
	if req != nil && req.URL != nil { key += ":" + req.URL.Path }

	// 断路器：若已打开则快速失败
	if s.cb != nil {
		ok, err := s.cb.Allow(ctx)
		if err != nil { return nil, err }
		if !ok { return nil, errors.New("circuit open") }
	}

	// 限流：按分钟窗口
	if s.rl != nil && s.rateLimit > 0 {
		ok, err := s.rl.AllowPerMinute(ctx, key, s.rateLimit)
		if err != nil { return nil, err }
		if !ok { return nil, errors.New("rate limited") }
	}
	// http.Client.Timeout 已覆盖整个请求生命周期（含 body 读取），无需再 WithTimeout。
	req = req.Clone(ctx)
	resp, err := s.client.Do(req)
	if err != nil {
		s.errors.Add(1)
		s.lastError.Store(err.Error())
		if s.cb != nil { _ = s.cb.RecordFailure(ctx) }
		return nil, err
	}
	if resp.StatusCode >= 500 {
		s.errors.Add(1)
		s.lastError.Store(resp.Status)
		if s.cb != nil { _ = s.cb.RecordFailure(ctx) }
	} else {
		s.lastSuccess.Store(time.Now())
		if s.cb != nil { _ = s.cb.RecordSuccess(ctx) }
	}
	return resp, nil
}

func (s *defaultSource) Healthz(ctx context.Context) HealthSnapshot { // ctx for future use
	_ = ctx
	var lastOK time.Time
	if v := s.lastSuccess.Load(); v != nil {
		lastOK, _ = v.(time.Time)
	}
	var lastErr string
	if v := s.lastError.Load(); v != nil {
		lastErr, _ = v.(string)
	}
	return HealthSnapshot{
		Vendor: s.vendor,
		LastSuccessAt: lastOK,
		LastError: lastErr,
		RequestsTotal: s.requests.Load(),
		ErrorsTotal: s.errors.Load(),
	}
}
