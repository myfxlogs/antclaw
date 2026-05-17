// Package feed provides content validation for social write operations (S12-P0-03).
package feed

import (
	"errors"
	"strings"

	"connectrpc.com/connect"
)

const (
	// MaxPostContentChars is the maximum number of runes for post content.
	MaxPostContentChars = 2000
	// MaxCommentContentChars is the maximum number of runes for comment content.
	MaxCommentContentChars = 1000
	// MaxShareCommentChars is the maximum number of runes for share comment (optional).
	MaxShareCommentChars = 1000
)

// validPostTypes is the set of allowed post_type values.
var validPostTypes = map[string]bool{
	"text":        true,
	"signal_card": true,
	"chart_share": true,
	"share":       true,
}

// validSignalDirections is the set of allowed signal_direction values.
var validSignalDirections = map[string]bool{
	"long":  true,
	"short": true,
	"":      true, // signal_card may omit direction
}

// ValidateCreatePostRequest validates a CreatePostRequest.
// Returns a connect.Error on failure, nil on success.
func ValidateCreatePostRequest(content, postType, signalDirection string, signalConfidence int32, visibility string) error {
	// post_type
	if !validPostTypes[postType] {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("post_type must be text, signal_card, chart_share, or share"))
	}

	// content length: 1..2000 (signal_card is exempt from content length floor)
	if postType != "signal_card" {
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("content must not be empty"))
		}
		if len([]rune(trimmed)) > MaxPostContentChars {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("content exceeds maximum length of 2000 characters"))
		}
	}

	// signal direction
	if signalDirection != "" || postType == "signal_card" {
		if !validSignalDirections[signalDirection] {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("signal_direction must be long, short, or empty"))
		}
	}

	// signal confidence: 0..100
	if signalConfidence < 0 || signalConfidence > 100 {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("signal_confidence must be between 0 and 100"))
	}

	return nil
}

// ValidateCommentRequest validates a CommentRequest.
func ValidateCommentRequest(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("comment content must not be empty"))
	}
	if len([]rune(trimmed)) > MaxCommentContentChars {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("comment exceeds maximum length of 1000 characters"))
	}
	return nil
}

// ValidateSharePostRequest validates a SharePostRequest.
func ValidateSharePostRequest(comment string) error {
	if len([]rune(comment)) > MaxShareCommentChars {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("share comment exceeds maximum length of 1000 characters"))
	}
	return nil
}
