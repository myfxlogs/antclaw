package ai

import (
	"context"
	"strings"
)

type Summarizer interface {
	Summarize(ctx context.Context, prompt string, maxTokens int) (string, error)
}

type DefaultSummarizer struct{}

func NewDefaultSummarizer() *DefaultSummarizer { return &DefaultSummarizer{} }

func (s *DefaultSummarizer) Summarize(ctx context.Context, prompt string, maxTokens int) (string, error) {
	out := strings.TrimSpace(prompt)
	if len(out) > 280 {
		out = out[:280]
	}
	return out, nil
}
