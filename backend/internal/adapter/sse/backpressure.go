package sse

import (
	"fmt"
	"time"

	streamv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

const (
	// MaxEventSize is the maximum size of a single event (64KiB).
	MaxEventSize = 64 * 1024
	
	// MaxQueueSize is the maximum number of events per client.
	MaxQueueSize = 100
	
	// DropThreshold triggers dropping oldest events.
	DropThreshold = 80
)

// BackpressureConfig configures backpressure behavior.
type BackpressureConfig struct {
	MaxEventSize   int
	MaxQueueSize   int
	DropThreshold  int
	DropOldest     bool
}

func DefaultBackpressureConfig() *BackpressureConfig {
	return &BackpressureConfig{
		MaxEventSize:  MaxEventSize,
		MaxQueueSize:  MaxQueueSize,
		DropThreshold: DropThreshold,
		DropOldest:    true,
	}
}

// EventQueue manages event queue with backpressure.
type EventQueue struct {
	config *BackpressureConfig
	events []*streamv1.SubscribeEventsResponse
	dropped int
}

func NewEventQueue(config *BackpressureConfig) *EventQueue {
	if config == nil {
		config = DefaultBackpressureConfig()
	}
	return &EventQueue{
		config: config,
		events: make([]*streamv1.SubscribeEventsResponse, 0, config.MaxQueueSize),
	}
}

// Push adds an event to the queue, applying backpressure if needed.
func (eq *EventQueue) Push(event *streamv1.SubscribeEventsResponse) (*streamv1.SubscribeEventsResponse, error) {
	// Check event size
	size := len(event.Payload)
	if size > eq.config.MaxEventSize {
		// Event too large, push snapshot URI instead
		return &streamv1.SubscribeEventsResponse{
			Id:        GenerateEventID(),
			Type:      "system.oversized",
			Payload:   fmt.Sprintf(`{"snapshot_uri":"/snapshots/%s","original_size":%d}`, event.Id, size),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}
	
	// Check if queue is full
	if len(eq.events) >= eq.config.MaxQueueSize {
		if eq.config.DropOldest {
			// Drop oldest event
			eq.events = eq.events[1:]
			eq.dropped++
		}
	}
	
	// Add event
	eq.events = append(eq.events, event)
	
	// Send dropped notice if needed
	if eq.dropped > 0 && eq.shouldNotifyDrop() {
		notice := &streamv1.SubscribeEventsResponse{
			Id:        GenerateEventID(),
			Type:      "system.notice.dropped",
			Payload:   fmt.Sprintf(`{"dropped":%d,"reason":"queue_full"}`, eq.dropped),
			Timestamp: time.Now().Format(time.RFC3339),
		}
		return notice, nil
	}
	
	return event, nil
}

// Pop removes and returns the oldest event.
func (eq *EventQueue) Pop() *streamv1.SubscribeEventsResponse {
	if len(eq.events) == 0 {
		return nil
	}
	event := eq.events[0]
	eq.events = eq.events[1:]
	return event
}

// Len returns the current queue length.
func (eq *EventQueue) Len() int {
	return len(eq.events)
}

// Dropped returns the number of dropped events.
func (eq *EventQueue) Dropped() int {
	return eq.dropped
}

func (eq *EventQueue) shouldNotifyDrop() bool {
	// Notify every 10 drops
	return eq.dropped%10 == 0
}

// SlowConsumerCheck checks if a client is a slow consumer.
func SlowConsumerCheck(queueLen, threshold int) bool {
	return queueLen >= threshold
}
