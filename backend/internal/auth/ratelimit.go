package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoginFailureBackoff implements progressive backoff for login failures
// Returns wait duration; 0 means no backoff required
func LoginFailureBackoff(ctx context.Context, rdb *redis.Client, email string) (time.Duration, error) {
	if rdb == nil {
		return 0, nil
	}
	key := fmt.Sprintf("login:fail:%s", email)
	
	count, err := rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return 0, err
	}

	backoff := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		30 * time.Second,
	}

	if count >= len(backoff) {
		return backoff[len(backoff)-1], nil
	}
	if count > 0 {
		return backoff[count-1], nil
	}
	return 0, nil
}

// RecordLoginFailure increments the failure counter
func RecordLoginFailure(ctx context.Context, rdb *redis.Client, email string) error {
	if rdb == nil {
		return nil
	}
	key := fmt.Sprintf("login:fail:%s", email)
	
	pipe := rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 1*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// ClearLoginFailures resets the failure counter on successful login
func ClearLoginFailures(ctx context.Context, rdb *redis.Client, email string) error {
	if rdb == nil {
		return nil
	}
	key := fmt.Sprintf("login:fail:%s", email)
	return rdb.Del(ctx, key).Err()
}

// CheckAndRecordRefreshReuse checks for refresh token reuse attack
// Returns true if reuse is detected
func CheckAndRecordRefreshReuse(ctx context.Context, rdb *redis.Client, jti string, rotatedTo string) (bool, error) {
	if rdb == nil {
		return false, nil
	}
	// Check if this token was already used after rotation
	key := fmt.Sprintf("refresh:used:%s", jti)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if exists > 0 {
		// Reuse detected
		return true, nil
	}

	// Mark as used if it was rotated
	if rotatedTo != "" {
		err = rdb.Set(ctx, key, rotatedTo, 30*24*time.Hour).Err()
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// RevokeRefreshToken adds a refresh token to the blacklist
func RevokeRefreshToken(ctx context.Context, rdb *redis.Client, jti string, ttl time.Duration) error {
	if rdb == nil {
		return nil
	}
	key := fmt.Sprintf("refresh:revoked:%s", jti)
	return rdb.Set(ctx, key, "1", ttl).Err()
}

// IsRefreshTokenRevoked checks if a refresh token is revoked
func IsRefreshTokenRevoked(ctx context.Context, rdb *redis.Client, jti string) (bool, error) {
	if rdb == nil {
		return false, nil
	}
	key := fmt.Sprintf("refresh:revoked:%s", jti)
	exists, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
