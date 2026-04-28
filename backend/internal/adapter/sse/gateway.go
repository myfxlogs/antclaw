package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/antclaw/antclaw/internal/auth"
	streamv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Gateway implements SSE streaming server.
type Gateway struct {
	hub       *Hub
	publisher *Publisher
}

func NewGateway(hub *Hub, publisher *Publisher) *Gateway {
	return &Gateway{hub: hub, publisher: publisher}
}

// ServeHTTP handles SSE connections.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Validate auth (optional for public channels)
	userID, _ := auth.UserIDFromContext(r.Context())
	
	// Get requested channel
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "channel required", http.StatusBadRequest)
		return
	}
	
	// Last-Event-ID for resume
	lastEventID := r.Header.Get("Last-Event-ID")
	if lastEventID == "" {
		lastEventID = r.URL.Query().Get("last_event_id")
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Create client
	client := &Client{
		id:        generateClientID(),
		channel:   channel,
		userID:    userID,
		lastID:    lastEventID,
		writer:    w,
		done:      make(chan struct{}),
	}
	
	// Register with hub
	g.hub.Register(client)
	defer g.hub.Unregister(client)
	
	// Send initial connection event
	g.sendEvent(w, &streamv1.SubscribeEventsResponse{
		Id:        fmt.Sprintf("evt-%d", time.Now().Unix()),
		Type:      "system.connected",
		Payload:   fmt.Sprintf(`{"client_id":"%s","channel":"%s"}`, client.id, channel),
		Timestamp: time.Now().Format(time.RFC3339),
	})
	
	// Keep connection alive
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	
	// Heartbeat ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-client.done:
			return
		case <-ticker.C:
			g.sendEvent(w, &streamv1.SubscribeEventsResponse{
				Id:        fmt.Sprintf("hb-%d", time.Now().Unix()),
				Type:      "system.heartbeat",
				Payload:   "{}",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			flusher.Flush()
		case event := <-client.send:
			g.sendEvent(w, event)
			flusher.Flush()
		}
	}
}

func (g *Gateway) sendEvent(w http.ResponseWriter, event *streamv1.SubscribeEventsResponse) {
	fmt.Fprintf(w, "id: %s\n", event.Id)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	
	payload, _ := json.Marshal(event.Payload)
	fmt.Fprintf(w, "data: %s\n", string(payload))
	fmt.Fprintf(w, "\n")
}

func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// Client represents an SSE client connection.
type Client struct {
	id      string
	channel string
	userID  string
	lastID  string
	writer  http.ResponseWriter
	send    chan *streamv1.SubscribeEventsResponse
	done    chan struct{}
}

// Publisher publishes events to Redis Streams.
type Publisher struct {
	hub *Hub
}

func NewPublisher(hub *Hub) *Publisher {
	return &Publisher{hub: hub}
}

// Publish sends event to channel subscribers.
func (p *Publisher) Publish(channel string, event *streamv1.SubscribeEventsResponse) {
	p.hub.Broadcast(channel, event)
}

// SubscribeCount returns number of subscribers for a channel.
func (p *Publisher) SubscribeCount(channel string) int {
	return p.hub.SubscriberCount(channel)
}
