// Package presence tracks online users connected via SSE.
// Thread-safe; multi-connection per user — a user is online as long as at least one connection is alive.
package presence

import (
	"sync"
	"time"
)

// Connection represents a single SSE connection for a user.
type Connection struct {
	ConnID      string    `json:"conn_id"`
	UserID      string    `json:"user_id"`
	RemoteAddr  string    `json:"remote_addr"`
	UserAgent   string    `json:"user_agent"`
	DeviceID    string    `json:"device_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

// OnlineUser represents an aggregated online user (from one or more connections).
type OnlineUser struct {
	UserID      string    `json:"user_id"`
	RemoteAddr  string    `json:"remote_addr"` // 最后建连的地址
	ConnectedAt time.Time `json:"connected_at"`
	ConnCount   int       `json:"conn_count"`
}

// Tracker maintains the set of currently connected users with multi-connection support.
type Tracker struct {
	mu    sync.RWMutex
	conns map[string]map[string]Connection // userID → connID → Connection
}

// NewTracker returns an empty Tracker.
func NewTracker() *Tracker {
	return &Tracker{conns: make(map[string]map[string]Connection)}
}

// Register adds a connection to the online set.
// If the same connID already exists, it is updated.
func (t *Tracker) Register(c Connection) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cm, ok := t.conns[c.UserID]
	if !ok {
		cm = make(map[string]Connection)
		t.conns[c.UserID] = cm
	}
	c.LastSeenAt = time.Now()
	cm[c.ConnID] = c
}

// DisconnectUser removes all connections for a user (single-device enforcement).
func (t *Tracker) DisconnectUser(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.conns, userID)
}

// Unregister removes a specific connection.
// When the user's last connection is removed, the user goes offline.
func (t *Tracker) Unregister(userID, connID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cm, ok := t.conns[userID]
	if !ok {
		return
	}
	delete(cm, connID)
	if len(cm) == 0 {
		delete(t.conns, userID)
	}
}

// Count returns the number of online users (not connections).
func (t *Tracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.conns)
}

// List returns a snapshot of all online users, aggregating multiple connections per user.
func (t *Tracker) List() []OnlineUser {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]OnlineUser, 0, len(t.conns))
	for userID, cm := range t.conns {
		u := OnlineUser{UserID: userID, ConnCount: len(cm)}
		for _, c := range cm {
			if u.ConnectedAt.IsZero() || c.ConnectedAt.Before(u.ConnectedAt) {
				u.ConnectedAt = c.ConnectedAt
				u.RemoteAddr = c.RemoteAddr
			}
		}
		out = append(out, u)
	}
	return out
}
