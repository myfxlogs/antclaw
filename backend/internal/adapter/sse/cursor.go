package sse

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CursorManager manages Last-Event-ID for SSE resume.
type CursorManager struct {
	client *redis.Client
	maxHistory int
}

func NewCursorManager(client *redis.Client) *CursorManager {
	return &CursorManager{
		client:     client,
		maxHistory: 500, // Keep last 500 events
	}
}

// Store saves event ID for a channel.
func (cm *CursorManager) Store(ctx context.Context, channel, eventID string) error {
	key := fmt.Sprintf("sse:events:%s", channel)
	
	// Add to sorted set with timestamp as score
	score := float64(time.Now().UnixMilli())
	if err := cm.client.ZAdd(ctx, key, redis.Z{Score: score, Member: eventID}).Err(); err != nil {
		return err
	}
	
	// Trim to max history
	return cm.client.ZRemRangeByRank(ctx, key, 0, -int64(cm.maxHistory)-1).Err()
}

// GetEventsAfter returns events after a given ID.
func (cm *CursorManager) GetEventsAfter(ctx context.Context, channel, lastID string) ([]string, error) {
	key := fmt.Sprintf("sse:events:%s", channel)
	
	// Get rank of lastID
	rank := cm.client.ZRank(ctx, key, lastID).Val()
	if rank < 0 {
		// ID not found, return empty (cursor_reset scenario)
		return nil, fmt.Errorf("cursor_reset")
	}
	
	// Get events after this rank
	members := cm.client.ZRange(ctx, key, rank+1, -1).Val()
	return members, nil
}

// IsValidID checks if an event ID exists.
func (cm *CursorManager) IsValidID(ctx context.Context, channel, eventID string) bool {
	key := fmt.Sprintf("sse:events:%s", channel)
	return cm.client.ZScore(ctx, key, eventID).Val() > 0
}

// GenerateEventID generates a unique event ID.
func GenerateEventID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixMilli(), time.Now().UnixNano()%1000)
}

// ParseEventID parses an event ID to extract timestamp.
func ParseEventID(eventID string) (int64, error) {
	parts := strings.Split(eventID, "-")
	if len(parts) < 1 {
		return 0, fmt.Errorf("invalid event ID format")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

// TrimEvents removes old events beyond max history.
func (cm *CursorManager) TrimEvents(ctx context.Context, channel string) error {
	key := fmt.Sprintf("sse:events:%s", channel)
	count, err := cm.client.ZCard(ctx, key).Result()
	if err != nil {
		return err
	}
	if count > int64(cm.maxHistory) {
		return cm.client.ZRemRangeByRank(ctx, key, 0, count-int64(cm.maxHistory)-1).Err()
	}
	return nil
}
