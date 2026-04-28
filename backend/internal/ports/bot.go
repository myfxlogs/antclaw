package ports

import (
	"context"
	"time"
)

// BotPlatform identifies the messaging platform.
type BotPlatform string

const (
	PlatformTelegram BotPlatform = "telegram"
	PlatformWeChat   BotPlatform = "wechat"
	PlatformFeishu   BotPlatform = "feishu"
)

// InboundMessage represents a message from external platform.
type InboundMessage struct {
	Platform       BotPlatform
	ExternalUserID string
	MessageID      string
	Text           string
	Timestamp      time.Time
}

// OutboundMessage represents a message to be sent to external platform.
type OutboundMessage struct {
	Platform       BotPlatform
	ExternalUserID string
	Text           string
	MessageKey     string
	Args           map[string]string
}

// BotBinding represents a binding between external user and internal user.
type BotBinding struct {
	ID             string
	Platform       BotPlatform
	ExternalUserID string
	UserID         string
	CreatedAt      time.Time
}

// BotPort defines the interface for bot adapters.
type BotPort interface {
	// Send sends an outbound message to the platform.
	Send(ctx context.Context, msg OutboundMessage) error
	
	// ParseCommand extracts command from message text.
	ParseCommand(text string) (command string, args []string)
}

// BotRepository defines storage interface for bot bindings.
type BotRepository interface {
	// GetBinding retrieves binding by platform and external user ID.
	GetBinding(ctx context.Context, platform BotPlatform, externalUserID string) (*BotBinding, error)
	
	// CreateBinding creates a new binding.
	CreateBinding(ctx context.Context, binding *BotBinding) error
	
	// DeleteBinding removes a binding.
	DeleteBinding(ctx context.Context, id string) error
}

// BotRouter defines command routing interface.
type BotRouter interface {
	// Route routes an inbound message to appropriate handler.
	Route(ctx context.Context, msg InboundMessage) (OutboundMessage, error)
	
	// Register registers a command handler.
	Register(command string, handler CommandHandler)
}

// CommandHandler handles bot commands.
type CommandHandler func(ctx context.Context, userID string, args []string) (OutboundMessage, error)

// BindTokenGenerator generates temporary binding tokens.
type BindTokenGenerator interface {
	// Generate creates a temporary token for user binding.
	Generate(userID string) (token string, expiresAt time.Time, err error)
	
	// Validate validates a binding token and returns associated user ID.
	Validate(token string) (userID string, err error)
}

// RateLimiter limits bot command rate.
type RateLimiter interface {
	// Allow checks if a request is allowed.
	Allow(ctx context.Context, key string) (bool, error)
	
	// Key generates rate limit key.
	Key(platform BotPlatform, externalUserID string) string
}
