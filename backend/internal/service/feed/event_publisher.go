// Package feed provides social event publishing for notification events (S12-P0-05).
package feed

import (
	"context"

	"github.com/google/uuid"
	"github.com/antclaw/antclaw/internal/notify"
)

// SocialEvent represents a social interaction that may trigger a notification.
type SocialEvent struct {
	Type       string // post_commented, post_shared, user_followed, post_liked
	ActorID    string // who did the action
	ActorName  string
	TargetID   string // who receives the notification (post author / followed user)
	PostID     string // optional, the post involved
	PostTitle  string // optional, first 80 chars of post content
	CommentID  string // optional
}

// SocialEventPublisher publishes social events for notification consumption.
type SocialEventPublisher interface {
	Publish(ctx context.Context, event SocialEvent) error
}

// NotifyEventPublisher publishes social events via the notify.Service.
type NotifyEventPublisher struct {
	notifySvc *notify.Service
}

// NewNotifyEventPublisher creates a publisher backed by the notification service.
func NewNotifyEventPublisher(notifySvc *notify.Service) *NotifyEventPublisher {
	return &NotifyEventPublisher{notifySvc: notifySvc}
}

func (p *NotifyEventPublisher) Publish(ctx context.Context, event SocialEvent) error {
	if event.TargetID == "" {
		return nil
	}

	var title, body, category string
	var data map[string]string

	switch event.Type {
	case "post_commented":
		title = event.ActorName + " 评论了你的帖子"
		body = firstN(event.PostTitle, 80)
		category = "social"
		data = map[string]string{
			"event_type": "post_commented",
			"post_id":    event.PostID,
			"comment_id": event.CommentID,
			"actor_id":   event.ActorID,
		}
	case "post_shared":
		title = event.ActorName + " 分享了你的帖子"
		body = firstN(event.PostTitle, 80)
		category = "social"
		data = map[string]string{
			"event_type": "post_shared",
			"post_id":    event.PostID,
			"actor_id":   event.ActorID,
		}
	case "user_followed":
		title = event.ActorName + " 关注了你"
		body = ""
		category = "social"
		data = map[string]string{
			"event_type": "user_followed",
			"actor_id":   event.ActorID,
		}
	case "post_liked":
		// Audit-only, no push notification per spec
		return nil
	default:
		return nil
	}

	targetUUID, err := uuid.Parse(event.TargetID)
	if err != nil {
		NewSocialLogger().Error("invalid_target_uuid", event.ActorID, "", err)
		return nil
	}

	n := &notify.Notification{
		UserID:   targetUUID,
		Category: category,
		Title:    title,
		Body:     body,
		Data:     data,
	}
	return p.notifySvc.Send(ctx, n)
}

// NoopEventPublisher discards all events. Useful for testing.
type NoopEventPublisher struct{}

func (NoopEventPublisher) Publish(_ context.Context, _ SocialEvent) error { return nil }

func firstN(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
