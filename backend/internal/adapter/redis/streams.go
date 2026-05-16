package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	streamv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/structpb"
)

// StreamsConsumer consumes events from Redis Streams.
type StreamsConsumer struct {
	client   *redis.Client
	group    string
	consumer string
	handlers map[string]EventHandler
}

// EventHandler processes stream events.
type EventHandler func(ctx context.Context, event *streamv1.SubscribeEventsResponse) error

func NewStreamsConsumer(client *redis.Client) *StreamsConsumer {
	return &StreamsConsumer{
		client:   client,
		group:    "antclaw-events",
		consumer: fmt.Sprintf("consumer-%d", time.Now().Unix()),
		handlers: make(map[string]EventHandler),
	}
}

// RegisterHandler registers an event handler for a channel.
func (c *StreamsConsumer) RegisterHandler(channel string, handler EventHandler) {
	c.handlers[channel] = handler
}

// CreateConsumerGroup creates a consumer group for a stream.
func (c *StreamsConsumer) CreateConsumerGroup(ctx context.Context, stream string) error {
	// Create consumer group, ignore error if already exists
	_ = c.client.XGroupCreateMkStream(ctx, stream, c.group, "0").Err()
	return nil
}

// Consume starts consuming events from streams.
func (c *StreamsConsumer) Consume(ctx context.Context, streams ...string) error {
	// Create consumer groups
	for _, stream := range streams {
		if err := c.CreateConsumerGroup(ctx, stream); err != nil {
			return fmt.Errorf("create consumer group for %s: %w", stream, err)
		}
	}
	
	// Build streams argument for XReadGroup
	args := &redis.XReadGroupArgs{
		Group:    c.group,
		Consumer: c.consumer,
		Streams:  make([]string, 0, len(streams)*2),
		Count:    100,
		Block:    5 * time.Second,
	}
	
	for _, stream := range streams {
		args.Streams = append(args.Streams, stream, ">")
	}
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Read messages
		streams, err := c.client.XReadGroup(ctx, args).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return fmt.Errorf("read group: %w", err)
		}
		
		// Process messages
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				if err := c.processMessage(ctx, stream.Stream, msg); err != nil {
					// Log error but continue
					continue
				}
				
				// Acknowledge message
				c.client.XAck(ctx, stream.Stream, c.group, msg.ID)
			}
		}
	}
}

func (c *StreamsConsumer) processMessage(ctx context.Context, stream string, msg redis.XMessage) error {
	// Extract channel from stream name
	channel := stream
	if len(stream) > 7 && stream[:7] == "events:" {
		channel = stream[7:]
	}
	
	// Get handler
	handler, ok := c.handlers[channel]
	if !ok {
		return fmt.Errorf("no handler for channel: %s", channel)
	}
	
	// Parse event
	event := &streamv1.SubscribeEventsResponse{
		Id:   msg.ID,
		Type: msg.Values["type"].(string),
	}
	if payload, ok := msg.Values["payload"].(string); ok {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &m); err == nil {
			s, _ := structpb.NewStruct(m)
			event.Payload = s
		}
	}
	
	// Call handler
	return handler(ctx, event)
}

// Publish publishes an event to a stream.
func Publish(ctx context.Context, client *redis.Client, channel string, event *streamv1.SubscribeEventsResponse) error {
	stream := fmt.Sprintf("events:%s", channel)
	
	payloadJSON, _ := json.Marshal(event.Payload.AsMap())
	values := map[string]interface{}{
		"type":    event.Type,
		"payload": string(payloadJSON),
	}
	
	return client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}
