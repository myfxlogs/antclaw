// Package presence tracks online users connected via SSE.
// Thread-safe; call Register/Unregister from SSE handlers.
package presence

import (
	"sync"
	"time"
)

// OnlineUser represents a connected user.
type OnlineUser struct {
	UserID     string    `json:"user_id"`
	RemoteAddr string    `json:"remote_addr"`
	ConnectedAt time.Time `json:"connected_at"`
}

// Tracker maintains the set of currently connected users.
type Tracker struct {
	mu    sync.RWMutex
	users map[string]OnlineUser // userID → OnlineUser
}

// NewTracker returns an empty Tracker.
func NewTracker() *Tracker {
	return &Tracker{users: make(map[string]OnlineUser)}
}

// Register adds a user to the online set.
// If already registered, updates RemoteAddr and ConnectedAt.
func (t *Tracker) Register(userID, remoteAddr string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.users[userID] = OnlineUser{
		UserID:      userID,
		RemoteAddr:  remoteAddr,
		ConnectedAt: time.Now(),
	}
}

// Unregister removes a user from the online set.
func (t *Tracker) Unregister(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.users, userID)
}

// Count returns the current number of online users.
func (t *Tracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.users)
}

// List returns a snapshot of all online users.
func (t *Tracker) List() []OnlineUser {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]OnlineUser, 0, len(t.users))
	for _, u := range t.users {
		out = append(out, u)
	}
	return out
}
