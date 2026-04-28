// Package notify provides notification services: in-app, mobile push (FCM/APNs/HMS), and email.
// See: AntClaw-重构解决方案.md §5.2
package notify

import (
"context"
"encoding/json"
"fmt"
"time"

"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
"github.com/google/uuid"
"github.com/redis/go-redis/v9"
)

// Service handles notification delivery.
type Service struct {
queries *db.Queries
redis   *redis.Client
}

// NewService creates a notification service.
func NewService(queries *db.Queries, redis *redis.Client) *Service {
return &Service{
queries: queries,
redis:   redis,
}
}

// Notification represents a notification to be sent.
type Notification struct {
UserID    uuid.UUID
Type      string
Title     string
Body      string
Data      map[string]string
Priority  string
}

// Send delivers a notification through appropriate channels.
func (s *Service) Send(ctx context.Context, n *Notification) error {
dataJSON, _ := json.Marshal(n.Data)
priority := n.Priority

_, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
UserID:   n.UserID,
Type:     n.Type,
Title:    n.Title,
Body:     n.Body,
Data:     dataJSON,
Priority: &priority,
})
if err != nil {
return fmt.Errorf("store notification: %w", err)
}

if s.redis != nil {
event := map[string]interface{}{
"type":      n.Type,
"title":     n.Title,
"body":      n.Body,
"data":      n.Data,
"timestamp": time.Now().Format(time.RFC3339),
}
eventJSON, _ := json.Marshal(event)
s.redis.Publish(ctx, "notify:new", eventJSON)
}

return nil
}

// SendInApp sends an in-app notification.
func (s *Service) SendInApp(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
return s.Send(ctx, &Notification{
UserID:   userID,
Type:     "in_app",
Title:    title,
Body:     body,
Data:     data,
Priority: "normal",
})
}

// MarkAsRead marks a notification as read.
func (s *Service) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
return s.queries.MarkNotificationRead(ctx, notificationID)
}

// GetUnread returns unread notifications for a user.
func (s *Service) GetUnread(ctx context.Context, userID uuid.UUID) ([]db.Notification, error) {
return s.queries.GetUnreadNotifications(ctx, userID)
}

// GetHistory returns notification history for a user.
func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, limit int32) ([]db.Notification, error) {
return s.queries.GetNotificationHistory(ctx, db.GetNotificationHistoryParams{
UserID: userID,
Limit:  limit,
})
}
