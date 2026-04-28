package redis

import (
	"context"
	"fmt"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitClosed    CircuitState = "closed"    // Normal operation
	CircuitOpen      CircuitState = "open"      // Failing fast
	CircuitHalfOpen  CircuitState = "half-open" // Testing if recovered
)

// CircuitBreaker implements the circuit breaker pattern using Redis
type CircuitBreaker struct {
	client       *Client
	failuresKey  string
	stateKey     string
	lastFailKey  string
	maxFailures  int
	resetTimeout time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(client *Client, name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		client:       client,
		failuresKey:  fmt.Sprintf("circuit:%s:failures", name),
		stateKey:     fmt.Sprintf("circuit:%s:state", name),
		lastFailKey:  fmt.Sprintf("circuit:%s:lastfail", name),
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Allow checks if the circuit allows requests
func (cb *CircuitBreaker) Allow(ctx context.Context) (bool, error) {
	state, err := cb.getState(ctx)
	if err != nil {
		return false, err
	}

	switch state {
	case CircuitClosed:
		return true, nil
	case CircuitOpen:
		// Check if timeout has passed
		lastFail, err := cb.getLastFailure(ctx)
		if err != nil {
			return false, err
		}
		if time.Since(lastFail) > cb.resetTimeout {
			// Transition to half-open
			if err := cb.setState(ctx, CircuitHalfOpen); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, nil
	case CircuitHalfOpen:
		return true, nil
	default:
		return false, fmt.Errorf("unknown circuit state: %s", state)
	}
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess(ctx context.Context) error {
	state, err := cb.getState(ctx)
	if err != nil {
		return err
	}

	// Reset failures
	if err := cb.client.Delete(ctx, cb.failuresKey); err != nil {
		return err
	}

	// If half-open, close the circuit
	if state == CircuitHalfOpen {
		return cb.setState(ctx, CircuitClosed)
	}
	return nil
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure(ctx context.Context) error {
	count, err := cb.client.Incr(ctx, cb.failuresKey)
	if err != nil {
		return err
	}

	// Set expiry on first failure
	if count == 1 {
		if err := cb.client.Expire(ctx, cb.failuresKey, cb.resetTimeout*2); err != nil {
			return err
		}
	}

	// Record last failure time
	if err := cb.client.Set(ctx, cb.lastFailKey, time.Now().Unix(), cb.resetTimeout*2); err != nil {
		return err
	}

	// Open circuit if threshold reached
	if count >= int64(cb.maxFailures) {
		return cb.setState(ctx, CircuitOpen)
	}
	return nil
}

// GetState returns current circuit state
func (cb *CircuitBreaker) GetState(ctx context.Context) (CircuitState, error) {
	return cb.getState(ctx)
}

func (cb *CircuitBreaker) getState(ctx context.Context) (CircuitState, error) {
	val, err := cb.client.Get(ctx, cb.stateKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return CircuitClosed, nil // Default state
		}
		return "", err
	}
	return CircuitState(val), nil
}

func (cb *CircuitBreaker) setState(ctx context.Context, state CircuitState) error {
	return cb.client.Set(ctx, cb.stateKey, string(state), cb.resetTimeout*2)
}

func (cb *CircuitBreaker) getLastFailure(ctx context.Context) (time.Time, error) {
	val, err := cb.client.Get(ctx, cb.lastFailKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	var unix int64
	_, err = fmt.Sscanf(val, "%d", &unix)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(unix, 0), nil
}

// Reset manually resets the circuit to closed
func (cb *CircuitBreaker) Reset(ctx context.Context) error {
	if err := cb.client.Delete(ctx, cb.failuresKey, cb.lastFailKey); err != nil {
		return err
	}
	return cb.setState(ctx, CircuitClosed)
}
