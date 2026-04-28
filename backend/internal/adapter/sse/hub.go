package sse

import (
	"sync"

	streamv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Hub manages SSE client subscriptions.
type Hub struct {
	clients   map[string]*Client
	channels  map[string]map[string]*Client
	mu        sync.RWMutex
	maxClients int
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		channels:   make(map[string]map[string]*Client),
		maxClients: 10000,
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Backpressure: check max clients
	if len(h.clients) >= h.maxClients {
		// Drop oldest client (simplified)
		for id := range h.clients {
			h.removeClientUnsafe(id)
			break
		}
	}
	
	h.clients[c.id] = c
	if h.channels[c.channel] == nil {
		h.channels[c.channel] = make(map[string]*Client)
	}
	h.channels[c.channel][c.id] = c
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeClientUnsafe(c.id)
}

func (h *Hub) removeClientUnsafe(id string) {
	c, ok := h.clients[id]
	if !ok {
		return
	}
	delete(h.clients, id)
	if c != nil {
		delete(h.channels[c.channel], id)
		close(c.done)
	}
}

// Broadcast sends an event to all subscribers of a channel.
func (h *Hub) Broadcast(channel string, event *streamv1.SubscribeEventsResponse) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients := h.channels[channel]
	if clients == nil {
		return
	}
	
	// Drop oldest if event > 64KiB (simplified check)
	for _, c := range clients {
		select {
		case c.send <- event:
		default:
			// Channel full, drop oldest
			select {
			case <-c.send:
				c.send <- event
			default:
			}
		}
	}
}

// SubscriberCount returns the number of subscribers for a channel.
func (h *Hub) SubscriberCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.channels[channel])
}

// ClientCount returns total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Close shuts down all client connections.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	for id := range h.clients {
		h.removeClientUnsafe(id)
	}
}
