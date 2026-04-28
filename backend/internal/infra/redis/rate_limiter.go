package redis

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter provides sliding window rate limiting
type RateLimiter struct {
	client *Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks if a request is allowed under rate limit
// limit: max requests per window
// window: time window duration
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now/int64(window.Seconds()))

	count, err := r.client.Incr(ctx, windowKey)
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		if err := r.client.Expire(ctx, windowKey, window); err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}

// AllowPerMinute checks rate limit per minute
func (r *RateLimiter) AllowPerMinute(ctx context.Context, key string, limit int) (bool, error) {
	return r.Allow(ctx, key, limit, time.Minute)
}

// AllowPerHour checks rate limit per hour
func (r *RateLimiter) AllowPerHour(ctx context.Context, key string, limit int) (bool, error) {
	return r.Allow(ctx, key, limit, time.Hour)
}

// AllowPerDay checks rate limit per day
func (r *RateLimiter) AllowPerDay(ctx context.Context, key string, limit int) (bool, error) {
	return r.Allow(ctx, key, limit, 24*time.Hour)
}

// GetCurrent returns current count without incrementing
func (r *RateLimiter) GetCurrent(ctx context.Context, key string, window time.Duration) (int64, error) {
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, now/int64(window.Seconds()))

	val, err := r.client.Get(ctx, windowKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, err
	}

	var count int64
	_, err = fmt.Sscanf(val, "%d", &count)
	return count, err
}
