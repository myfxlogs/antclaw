package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	redisv9 "github.com/redis/go-redis/v9"

	streamv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// streamKeyMap maps channel names to Redis Stream keys.
var streamKeyMap = map[string]string{
	"jobs":           "stream:jobs_events",
	"audit":          "stream:audit_events",
	"macro_alerts":   "stream:macro_alerts",
	"options_alerts": "stream:options_alerts",
	"signals_alerts": "stream:signals_alerts",
}

// StreamHandler implements the StreamService Connect streaming RPC.
// Replaces legacy SSE handlers with protobuf binary streaming.
type StreamHandler struct {
	rdb *redisv9.Client
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(rdb *redisv9.Client) *StreamHandler {
	return &StreamHandler{rdb: rdb}
}

// SubscribeEvents handles server-streaming subscription requests.
func (h *StreamHandler) SubscribeEvents(
	ctx context.Context,
	req *connect.Request[streamv1.SubscribeEventsRequest],
	stream *connect.ServerStream[streamv1.SubscribeEventsResponse],
) error {
	channel := req.Msg.Channel
	streamKey, ok := streamKeyMap[channel]
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown channel: %s", channel))
	}

	lastID := req.Msg.LastEventId
	if lastID == "" {
		lastID = "$"
	}

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeat.C:
			if err := stream.Send(&streamv1.SubscribeEventsResponse{
				Type:      "system.heartbeat",
				Timestamp: time.Now().Format(time.RFC3339),
			}); err != nil {
				return err
			}
		default:
		}

		streams, err := h.rdb.XRead(ctx, &redisv9.XReadArgs{
			Streams: []string{streamKey, lastID},
			Block:   5 * time.Second,
			Count:   10,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			if err == redisv9.Nil {
				continue
			}
			log.Printf("stream RPC %s XRead error: %v", channel, err)
			time.Sleep(time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				lastID = msg.ID
				payload := msgToStruct(msg.Values)
				if err := stream.Send(&streamv1.SubscribeEventsResponse{
					Id:        msg.ID,
					Type:      "event",
					Payload:   payload,
					Timestamp: time.Now().Format(time.RFC3339),
				}); err != nil {
					return err
				}
			}
		}
	}
}

// msgToStruct converts Redis stream message values to a protobuf Struct.
func msgToStruct(values map[string]interface{}) *structpb.Struct {
	// Try "data" field as JSON string first
	if data, ok := values["data"].(string); ok {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(data), &m); err == nil {
			s, _ := structpb.NewStruct(m)
			return s
		}
	}
	// Fallback: convert all values
	m := make(map[string]interface{}, len(values))
	for k, v := range values {
		m[k] = v
	}
	s, _ := structpb.NewStruct(m)
	return s
}
