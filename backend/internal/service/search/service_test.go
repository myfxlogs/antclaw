package search

import (
	"context"
	"testing"
	"time"

	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// fakeSearchRepo implements SearchRepository with canned test data.
type fakeSearchRepo struct {
	users   []*postgres.SearchUserRow
	posts   []*postgres.FeedPostRow
	symbols []*postgres.SearchSymbolRow
}

func newFakeSearchRepo() *fakeSearchRepo {
	return &fakeSearchRepo{}
}

func (f *fakeSearchRepo) SearchUsers(_ context.Context, query string, limit int32) ([]*postgres.SearchUserRow, error) {
	var out []*postgres.SearchUserRow
	for _, u := range f.users {
		if int32(len(out)) >= limit {
			break
		}
		out = append(out, u)
	}
	return out, nil
}

func (f *fakeSearchRepo) SearchPosts(_ context.Context, query string, limit int32) ([]*postgres.FeedPostRow, [][]string, error) {
	var out []*postgres.FeedPostRow
	for _, p := range f.posts {
		if int32(len(out)) >= limit {
			break
		}
		out = append(out, p)
	}
	likedBy := make([][]string, len(out))
	return out, likedBy, nil
}

func (f *fakeSearchRepo) SearchSymbols(_ context.Context, query string, limit int32) ([]*postgres.SearchSymbolRow, error) {
	var out []*postgres.SearchSymbolRow
	for _, s := range f.symbols {
		if int32(len(out)) >= limit {
			break
		}
		out = append(out, s)
	}
	return out, nil
}

// ----- Tests -----

func TestSearch_QueryTooShort(t *testing.T) {
	svc := NewService(newFakeSearchRepo())
	_, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "a"})
	if err == nil {
		t.Fatal("expected InvalidArgument for query < 2 chars")
	}
}

func TestSearch_QueryTrimmed(t *testing.T) {
	svc := NewService(newFakeSearchRepo())
	_, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: " a "})
	if err == nil {
		t.Fatal("expected InvalidArgument for trimmed query < 2 chars")
	}
}

func TestSearch_QueryValid(t *testing.T) {
	repo := newFakeSearchRepo()
	repo.users = []*postgres.SearchUserRow{
		{UserID: "u1", DisplayName: "Alice", Tier: "normal", FollowerCount: 10},
	}
	svc := NewService(repo)
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "alice", PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Users))
	}
	if resp.Users[0].UserId != "u1" {
		t.Fatalf("expected u1, got %s", resp.Users[0].UserId)
	}
}

func TestSearch_PostsMappedCorrectly(t *testing.T) {
	repo := newFakeSearchRepo()
	now := time.Now()
	repo.posts = []*postgres.FeedPostRow{
		{ID: "p1", AuthorID: "a1", AuthorName: "User1", Content: "hello world", PostType: "text",
			Visibility: "public", LikeCount: 3, CommentCount: 1, CreatedAt: now},
	}
	svc := NewService(repo)
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "hello", PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(resp.Posts))
	}
	if resp.Posts[0].Id != "p1" {
		t.Fatalf("expected p1, got %s", resp.Posts[0].Id)
	}
	if resp.Posts[0].Content != "hello world" {
		t.Fatalf("content mismatch")
	}
}

func TestSearch_SymbolsMappedCorrectly(t *testing.T) {
	repo := newFakeSearchRepo()
	repo.symbols = []*postgres.SearchSymbolRow{
		{Symbol: "EURUSD", DisplayName: "EURUSD", Market: "forex"},
		{Symbol: "BTC:USD", DisplayName: "BTC:USD", Market: "crypto"},
	}
	svc := NewService(repo)
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "usd", PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(resp.Symbols))
	}
	if resp.Symbols[0].Symbol != "EURUSD" {
		t.Fatalf("expected EURUSD, got %s", resp.Symbols[0].Symbol)
	}
}

func TestSearch_DefaultScopes(t *testing.T) {
	repo := newFakeSearchRepo()
	repo.users = []*postgres.SearchUserRow{
		{UserID: "u1", DisplayName: "Test", Tier: "normal"},
	}
	repo.symbols = []*postgres.SearchSymbolRow{
		{Symbol: "AAPL", DisplayName: "AAPL", Market: "stock"},
	}
	svc := NewService(repo)
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "aa", PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty scopes ⇒ defaults: users, posts, symbols
	if len(resp.Users) == 0 {
		t.Fatal("expected users in default scopes")
	}
	if len(resp.Symbols) == 0 {
		t.Fatal("expected symbols in default scopes")
	}
}

func TestSearch_ScopedFiltering(t *testing.T) {
	repo := newFakeSearchRepo()
	repo.users = []*postgres.SearchUserRow{
		{UserID: "u1", DisplayName: "Test", Tier: "normal"},
	}
	svc := NewService(repo)
	// Only search posts — no users should appear
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{
		Query:  "aa",
		Scopes: []string{"posts"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 0 {
		t.Fatalf("expected 0 users with posts-only scope, got %d", len(resp.Users))
	}
}

func TestSearch_LimitClamping(t *testing.T) {
	repo := newFakeSearchRepo()
	for i := 0; i < 60; i++ {
		repo.users = append(repo.users, &postgres.SearchUserRow{
			UserID: "u", DisplayName: "U", Tier: "normal",
		})
	}
	svc := NewService(repo)
	resp, err := svc.Search(context.Background(), &alfqv1.SearchRequest{Query: "aa", PageSize: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// PageSize 100 should be clamped to 50
	if len(resp.Users) > 50 {
		t.Fatalf("expected max 50 users, got %d", len(resp.Users))
	}
}
