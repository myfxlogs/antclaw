// Package search provides the business logic for cross-entity search (P1).
package search

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service"
	feedpkg "github.com/antclaw/antclaw/internal/service/feed"
)

// Service holds the search business logic.
type Service struct {
	repo postgres.SearchRepository
}

func NewService(repo postgres.SearchRepository) *Service {
	return &Service{repo: repo}
}

// defaultScopes returns the default set of search scopes.
func defaultScopes() []string {
	return []string{"users", "posts", "symbols"}
}

// normalizeScopes returns the canonical scope list. Empty → defaults.
func normalizeScopes(raw []string) []string {
	if len(raw) == 0 {
		return defaultScopes()
	}
	// Deduplicate and validate
	seen := make(map[string]bool, len(raw))
	var out []string
	for _, s := range raw {
		switch s {
		case "users", "posts", "symbols":
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	if len(out) == 0 {
		return defaultScopes()
	}
	return out
}

// Search executes a cross-entity search.
func (s *Service) Search(ctx context.Context, req *alfqv1.SearchRequest) (*alfqv1.SearchResponse, error) {
	query := strings.TrimSpace(req.Query)
	if len(query) < 2 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("query must be at least 2 characters"))
	}
	scopes := normalizeScopes(req.Scopes)
	limit := service.ClampPageSize(req.PageSize, 10, 50)

	resp := &alfqv1.SearchResponse{}

	for _, scope := range scopes {
		switch scope {
		case "users":
			users, err := s.repo.SearchUsers(ctx, query, limit)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("search users: %w", err))
			}
			for _, u := range users {
				resp.Users = append(resp.Users, &alfqv1.UserSearchResult{
					UserId:        u.UserID,
					DisplayName:   u.DisplayName,
					Tier:          u.Tier,
					FollowerCount: u.FollowerCount,
				})
			}
		case "posts":
			rows, _, err := s.repo.SearchPosts(ctx, query, limit)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("search posts: %w", err))
			}
			for _, row := range rows {
				resp.Posts = append(resp.Posts, feedpkg.PostRowToProto(row, nil))
			}
		case "symbols":
			symbols, err := s.repo.SearchSymbols(ctx, query, limit)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("search symbols: %w", err))
			}
			for _, s := range symbols {
				resp.Symbols = append(resp.Symbols, &alfqv1.SymbolSearchResult{
					Symbol:      s.Symbol,
					DisplayName: s.DisplayName,
					Market:      s.Market,
				})
			}
		}
	}
	return resp, nil
}


