package stub

import (
	"context"
	"errors"
	"strings"

	"github.com/antclaw/antclaw/internal/ports"
)

// StubBot is a stub implementation of ports.BotPort for P7.
type StubBot struct {
	platform ports.BotPlatform
}

func NewStubBot(platform ports.BotPlatform) *StubBot {
	return &StubBot{platform: platform}
}

// Send implements ports.BotPort.Send.
func (s *StubBot) Send(ctx context.Context, msg ports.OutboundMessage) error {
	// Stub: just validate, don't actually send
	if msg.Platform != s.platform {
		return errors.New("platform mismatch")
	}
	if msg.ExternalUserID == "" {
		return errors.New("external_user_id required")
	}
	return nil
}

// ParseCommand implements ports.BotPort.ParseCommand.
func (s *StubBot) ParseCommand(text string) (command string, args []string) {
	// Simple parsing: /command arg1 arg2
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") {
		return "", nil
	}
	
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil
	}
	
	command = strings.TrimPrefix(parts[0], "/")
	if len(parts) > 1 {
		args = parts[1:]
	}
	return command, args
}

// Ensure StubBot implements BotPort
var _ ports.BotPort = (*StubBot)(nil)
