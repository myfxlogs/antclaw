package bot

import (
	"context"
	"errors"

	"github.com/antclaw/antclaw/internal/ports"
)

// Router implements ports.BotRouter for command routing.
type Router struct {
	handlers map[string]ports.CommandHandler
	repo     ports.BotRepository
}

func NewRouter(repo ports.BotRepository) *Router {
	return &Router{
		handlers: make(map[string]ports.CommandHandler),
		repo:     repo,
	}
}

// Route routes an inbound message to appropriate handler.
func (r *Router) Route(ctx context.Context, msg ports.InboundMessage) (ports.OutboundMessage, error) {
	// Check if user is bound
	binding, err := r.repo.GetBinding(ctx, msg.Platform, msg.ExternalUserID)
	if err != nil {
		// Not bound: only allow whitelist commands
		return r.handleUnboundUser(ctx, msg)
	}
	
	// Parse command
	command, args := r.parseCommand(msg.Text)
	
	// Find handler
	handler, ok := r.handlers[command]
	if !ok {
		return ports.OutboundMessage{
			Platform:       msg.Platform,
			ExternalUserID: msg.ExternalUserID,
			MessageKey:     "error.unknown_command",
		}, nil
	}
	
	// Call handler
	return handler(ctx, binding.UserID, args)
}

// Register registers a command handler.
func (r *Router) Register(command string, handler ports.CommandHandler) {
	r.handlers[command] = handler
}

// handleUnboundUser handles commands from unbound users.
func (r *Router) handleUnboundUser(ctx context.Context, msg ports.InboundMessage) (ports.OutboundMessage, error) {
	// Only allow /bind and /help commands
	command, _ := r.parseCommand(msg.Text)
	
	switch command {
	case "bind":
		return ports.OutboundMessage{
			Platform:       msg.Platform,
			ExternalUserID: msg.ExternalUserID,
			MessageKey:     "bot.bind.instructions",
		}, nil
	case "help":
		return ports.OutboundMessage{
			Platform:       msg.Platform,
			ExternalUserID: msg.ExternalUserID,
			MessageKey:     "bot.help.message",
		}, nil
	default:
		return ports.OutboundMessage{
			Platform:       msg.Platform,
			ExternalUserID: msg.ExternalUserID,
			MessageKey:     "bot.error.requires_binding",
		}, nil
	}
}

// parseCommand extracts command from message.
func (r *Router) parseCommand(text string) (string, []string) {
	// Simple stub parsing
	// Real implementation should use platform-specific parser
	return "", nil
}

// Ensure Router implements BotRouter
var _ ports.BotRouter = (*Router)(nil)

// ErrNotBound indicates user is not bound.
var ErrNotBound = errors.New("user not bound")
