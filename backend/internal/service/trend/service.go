// Package trend provides the business logic for trending topics and hot symbols (P1).
package trend

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service"
)

// Service holds the trend business logic.
type Service struct {
	repo postgres.TrendRepository
}

func NewService(repo postgres.TrendRepository) *Service {
	return &Service{repo: repo}
}

// validWindows is the set of allowed aggregation windows.
var validWindows = map[string]bool{
	"1h":  true,
	"24h": true,
	"7d":  true,
}

// ListTrendingTopics returns trending hashtags for the given window.
//
// Algorithm (P1 default):
//   - Extract #hashtags from public post content within the time window
//   - Rank by engagement_count = like_count + comment_count + share_count
func (s *Service) ListTrendingTopics(ctx context.Context, req *alfqv1.ListTrendingTopicsRequest) (*alfqv1.ListTrendingTopicsResponse, error) {
	window := req.Window
	if window == "" {
		window = "24h"
	}
	if !validWindows[window] {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("window must be one of: 1h, 24h, 7d"))
	}
	limit := service.ClampPageSize(req.Limit, 10, 50)
	rows, err := s.repo.ListTrendingTopics(ctx, window, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list trending topics: %w", err))
	}
	topics := make([]*alfqv1.TrendingTopic, len(rows))
	for i, r := range rows {
		topics[i] = &alfqv1.TrendingTopic{
			Topic:           r.Topic,
			PostCount:       r.PostCount,
			EngagementCount: r.EngagementCount,
		}
	}
	return &alfqv1.ListTrendingTopicsResponse{Topics: topics}, nil
}

// ListHotSymbols returns the most discussed trading symbols for the given window.
//
// Algorithm (P1 default):
//   - Aggregate signal_card posts by signal_pair within the time window
//   - Rank by engagement_count = like_count + comment_count + share_count
func (s *Service) ListHotSymbols(ctx context.Context, req *alfqv1.ListHotSymbolsRequest) (*alfqv1.ListHotSymbolsResponse, error) {
	window := req.Window
	if window == "" {
		window = "24h"
	}
	if !validWindows[window] {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("window must be one of: 1h, 24h, 7d"))
	}
	limit := service.ClampPageSize(req.Limit, 10, 50)
	rows, err := s.repo.ListHotSymbols(ctx, window, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("list hot symbols: %w", err))
	}
	symbols := make([]*alfqv1.HotSymbol, len(rows))
	for i, r := range rows {
		symbols[i] = &alfqv1.HotSymbol{
			Symbol:          r.Symbol,
			PostCount:       r.PostCount,
			SignalCount:     r.SignalCount,
			EngagementCount: r.EngagementCount,
		}
	}
	return &alfqv1.ListHotSymbolsResponse{Symbols: symbols}, nil
}


