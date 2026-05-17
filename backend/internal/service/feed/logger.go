// Package feed provides structured logging for social operations (S12-P2-02).
package feed

import (
	"log"
)

// SocialLogger provides structured context for social operations.
type SocialLogger struct{}

// NewSocialLogger creates a logger for social operations.
func NewSocialLogger() *SocialLogger { return &SocialLogger{} }

// Info logs an informational event with structured fields.
func (l *SocialLogger) Info(action, userID, postID string, extra map[string]interface{}) {
	entry := map[string]interface{}{
		"action":  action,
		"user_id": userID,
	}
	if postID != "" {
		entry["post_id"] = postID
	}
	for k, v := range extra {
		entry[k] = v
	}
	log.Printf("[social] %v", entry)
}

// Error logs an error event with structured fields.
func (l *SocialLogger) Error(action, userID, postID string, err error) {
	entry := map[string]interface{}{
		"action":  action,
		"user_id": userID,
		"error":   err.Error(),
	}
	if postID != "" {
		entry["post_id"] = postID
	}
	log.Printf("[social:error] %v", entry)
}
