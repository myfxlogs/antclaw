package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DistributedLock provides Redis-based distributed locking
type DistributedLock struct {
	client   *Client
	workerID string
}

// NewDistributedLock creates a new distributed lock manager
func NewDistributedLock(client *Client) *DistributedLock {
	return &DistributedLock{
		client:   client,
		workerID: uuid.New().String(),
	}
}

// Acquire attempts to acquire a lock
// Returns true if lock was acquired, false if already held by another
func (d *DistributedLock) Acquire(ctx context.Context, jobKey string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:job:%s", jobKey)
	return d.client.SetNX(ctx, key, d.workerID, ttl)
}

// Release releases a lock (only if held by this worker)
func (d *DistributedLock) Release(ctx context.Context, jobKey string) error {
	key := fmt.Sprintf("lock:job:%s", jobKey)
	// Get current value
	val, err := d.client.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil // Already released
		}
		return err
	}
	// Only delete if we hold the lock
	if val == d.workerID {
		return d.client.Delete(ctx, key)
	}
	return nil // Not our lock
}

// Extend extends the TTL of a lock we hold
func (d *DistributedLock) Extend(ctx context.Context, jobKey string, ttl time.Duration) error {
	key := fmt.Sprintf("lock:job:%s", jobKey)
	val, err := d.client.Get(ctx, key)
	if err != nil {
		return err
	}
	if val != d.workerID {
		return fmt.Errorf("lock not held by this worker")
	}
	return d.client.Set(ctx, key, d.workerID, ttl)
}

// IsHeld checks if a lock is currently held (by any worker)
func (d *DistributedLock) IsHeld(ctx context.Context, jobKey string) (bool, error) {
	key := fmt.Sprintf("lock:job:%s", jobKey)
	_, err := d.client.Get(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetWorkerID returns this worker's ID
func (d *DistributedLock) GetWorkerID() string {
	return d.workerID
}
