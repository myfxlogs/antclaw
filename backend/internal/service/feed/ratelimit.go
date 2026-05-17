// Package feed provides social rate limiting for write operations (S12-P0-04).
package feed

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/infra/redis"
)

// Default rate limits (per user per action per window).
// Documented defaults per S12-P0-04 spec.
const (
	DefaultPostRateLimit    = 5  // per minute
	DefaultCommentRateLimit = 20 // per minute
	DefaultLikeRateLimit    = 60 // per minute
	DefaultFollowRateLimit  = 20 // per minute
	DefaultShareRateLimit   = 10 // per minute
	rateLimitWindow         = time.Minute
)

// RateLimitAction names the social write action for rate-limit key construction.
type RateLimitAction string

const (
	RateLimitCreatePost    RateLimitAction = "create_post"
	RateLimitCommentOnPost RateLimitAction = "comment"
	RateLimitLikePost      RateLimitAction = "like"
	RateLimitSharePost     RateLimitAction = "share"
	RateLimitFollow        RateLimitAction = "follow"
)

// SocialRateLimiter restricts write frequency per user + action.
type SocialRateLimiter interface {
	// Allow reports whether the action is within rate limit for the given user.
	// Returns a connect.Error (ResourceExhausted) if the limit is exceeded.
	Allow(ctx context.Context, userID string, action RateLimitAction) error
}

// RedisRateLimiter implements SocialRateLimiter using Redis counters.
type RedisRateLimiter struct {
	rdb *redis.Client
}

// NewRedisRateLimiter creates a Redis-backed rate limiter.
func NewRedisRateLimiter(rdb *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{rdb: rdb}
}

func (l *RedisRateLimiter) Allow(ctx context.Context, userID string, action RateLimitAction) error {
	limit := maxPerWindow(action)
	key := fmt.Sprintf("ratelimit:social:%s:%s", userID, string(action))

	count, err := l.rdb.Incr(ctx, key)
	if err != nil {
		// fail-closed: treat Redis error as denial to protect the system
		return connect.NewError(connect.CodeResourceExhausted, errors.New("rate limit check unavailable"))
	}
	if count == 1 {
		_ = l.rdb.Expire(ctx, key, rateLimitWindow)
	}
	if count > int64(limit) {
		return connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("rate limit exceeded: %d/%d per minute", limit, limit))
	}
	return nil
}

func maxPerWindow(action RateLimitAction) int {
	switch action {
	case RateLimitCreatePost:
		return DefaultPostRateLimit
	case RateLimitCommentOnPost:
		return DefaultCommentRateLimit
	case RateLimitLikePost:
		return DefaultLikeRateLimit
	case RateLimitFollow:
		return DefaultFollowRateLimit
	case RateLimitSharePost:
		return DefaultShareRateLimit
	default:
		return DefaultPostRateLimit
	}
}

// NoopRateLimiter allows all actions. Useful for testing when Redis is unavailable.
type NoopRateLimiter struct{}

func (NoopRateLimiter) Allow(_ context.Context, _ string, _ RateLimitAction) error { return nil }
