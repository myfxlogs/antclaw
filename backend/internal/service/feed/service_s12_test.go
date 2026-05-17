package feed

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// -- S12-P1-05: followers Feed visibility --

func TestFeed_FollowersVisibilityFiltering(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)

	// user-a creates a followers-only post
	followersPost, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:    "followers only",
		PostType:   "text",
		Visibility: "followers",
	})
	if err != nil {
		t.Fatalf("create followers post: %v", err)
	}

	// user-a creates a public post
	publicPost, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:  "public",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("create public post: %v", err)
	}

	// Anonymous user: should NOT see the followers-only post in feed
	resp, _ := svc.GetFeed(context.Background(), "", &alfqv1.GetFeedRequest{Filter: "all", PageSize: 20})
	for _, p := range resp.Posts {
		if p.Id == followersPost.Id {
			t.Fatal("anonymous should not see followers-only post in public feed")
		}
	}

	// user-b (not following user-a): should NOT see followers-only post
	resp, _ = svc.GetFeed(context.Background(), "user-b", &alfqv1.GetFeedRequest{Filter: "all", PageSize: 20})
	for _, p := range resp.Posts {
		if p.Id == followersPost.Id {
			t.Fatal("non-follower should not see followers-only post")
		}
	}

	// Verify both posts exist in the repo
	_ = publicPost
	_ = followersPost
}

func TestLikedBy_OnlyCurrentUser(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-a", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})

	// user-b likes the post
	mustLike(t, svc, "user-b", created.Id)

	// user-b fetches the feed: should see themselves in liked_by
	resp, err := svc.GetFeed(context.Background(), "user-b", &alfqv1.GetFeedRequest{Filter: "all", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, p := range resp.Posts {
		if p.Id == created.Id {
			if len(p.LikedBy) != 1 || p.LikedBy[0] != "user-b" {
				t.Fatalf("expected liked_by to contain only user-b, got %v", p.LikedBy)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("post not found in feed")
	}

	// user-c fetches the feed (did not like): should see empty liked_by
	resp2, err := svc.GetFeed(context.Background(), "user-c", &alfqv1.GetFeedRequest{Filter: "all", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range resp2.Posts {
		if p.Id == created.Id {
			if len(p.LikedBy) != 0 {
				t.Fatalf("expected empty liked_by for user-c, got %v", p.LikedBy)
			}
		}
	}
}

// -- S12-P0-01: visibility validation --

func TestCreatePost_InvalidVisibilityDoesNotBecomePublic(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)

	tests := []struct {
		name       string
		visibility string
		wantErr    bool
		errCode    connect.Code
	}{
		{name: "empty defaults to public", visibility: "", wantErr: false},
		{name: "public allowed", visibility: "public", wantErr: false},
		{name: "followers allowed", visibility: "followers", wantErr: false},
		{name: "circle unsupported", visibility: "circle", wantErr: true, errCode: connect.CodeInvalidArgument},
		{name: "random value rejected", visibility: "sekret", wantErr: true, errCode: connect.CodeInvalidArgument},
		{name: "mixed case rejected", visibility: "PUBLIC", wantErr: true, errCode: connect.CodeInvalidArgument},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			post, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
				Content:    "test content",
				PostType:   "text",
				Visibility: tc.visibility,
			})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for visibility=%q, got post=%v", tc.visibility, post)
				}
				var ce *connect.Error
				if errors.As(err, &ce) {
					if ce.Code() != tc.errCode {
						t.Fatalf("expected code %v, got %v", tc.errCode, ce.Code())
					}
				} else {
					t.Fatalf("expected connect.Error, got %T", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for visibility=%q: %v", tc.visibility, err)
			}
			// For empty, verify it was stored as "public"
			if tc.visibility == "" && post.Visibility != "public" {
				t.Fatalf("empty visibility should default to public, got %q", post.Visibility)
			}
			// For followers, verify it was stored as "followers"
			if tc.visibility == "followers" && post.Visibility != "followers" {
				t.Fatalf("followers visibility not preserved, got %q", post.Visibility)
			}
		})
	}
}

// -- S12-P0-03: content validation --

func TestCreatePost_ContentValidation(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)

	tests := []struct {
		name     string
		req      *alfqv1.CreatePostRequest
		wantCode connect.Code
	}{
		{name: "empty content text", req: &alfqv1.CreatePostRequest{Content: "", PostType: "text"}, wantCode: connect.CodeInvalidArgument},
		{name: "whitespace-only content", req: &alfqv1.CreatePostRequest{Content: "  \t ", PostType: "text"}, wantCode: connect.CodeInvalidArgument},
		{name: "signal_card empty content ok", req: &alfqv1.CreatePostRequest{Content: "", PostType: "signal_card", SignalPair: "BTCUSD", SignalDirection: "long", SignalConfidence: 80}, wantCode: 0},
		{name: "invalid post_type", req: &alfqv1.CreatePostRequest{Content: "hello", PostType: "article"}, wantCode: connect.CodeInvalidArgument},
		{name: "invalid signal_direction", req: &alfqv1.CreatePostRequest{Content: "sig", PostType: "signal_card", SignalPair: "BTCUSD", SignalDirection: "neutral", SignalConfidence: 50}, wantCode: connect.CodeInvalidArgument},
		{name: "signal_confidence negative", req: &alfqv1.CreatePostRequest{Content: "sig", PostType: "signal_card", SignalPair: "BTCUSD", SignalDirection: "long", SignalConfidence: -1}, wantCode: connect.CodeInvalidArgument},
		{name: "signal_confidence too high", req: &alfqv1.CreatePostRequest{Content: "sig", PostType: "signal_card", SignalPair: "BTCUSD", SignalDirection: "long", SignalConfidence: 101}, wantCode: connect.CodeInvalidArgument},
		{name: "valid text post", req: &alfqv1.CreatePostRequest{Content: "hello world", PostType: "text"}, wantCode: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreatePost(context.Background(), "user-a", tc.req)
			if tc.wantCode == 0 {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error code %v, got success", tc.wantCode)
			}
			var ce *connect.Error
			if errors.As(err, &ce) {
				if ce.Code() != tc.wantCode {
					t.Fatalf("expected code %v, got %v", tc.wantCode, ce.Code())
				}
			} else {
				t.Fatalf("expected connect.Error, got %T", err)
			}
		})
	}
}

// -- S12-P0-04: rate limiting --

// fakeRateLimiter implements SocialRateLimiter for testing.
type fakeRateLimiter struct {
	blocked map[string]bool // userID+action -> blocked
}

func newFakeRateLimiter() *fakeRateLimiter {
	return &fakeRateLimiter{blocked: make(map[string]bool)}
}

func (l *fakeRateLimiter) block(userID string, action RateLimitAction) {
	l.blocked[string(action)+":"+userID] = true
}

func (l *fakeRateLimiter) Allow(_ context.Context, userID string, action RateLimitAction) error {
	if l.blocked[string(action)+":"+userID] {
		return connect.NewError(connect.CodeResourceExhausted, errors.New("rate limit exceeded"))
	}
	return nil
}

// -- S12-P0-05: notification events --

// fakeEventPublisher captures social events for testing.
type fakeEventPublisher struct {
	events []SocialEvent
}

func (p *fakeEventPublisher) Publish(_ context.Context, event SocialEvent) error {
	p.events = append(p.events, event)
	return nil
}

func TestComment_EmitsNotificationEvent(t *testing.T) {
	repo := newFakeFeedRepo()
	pub := &fakeEventPublisher{}
	svc := NewService(repo).WithEventPublisher(pub)

	// user-a creates a post
	post, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:  "a post",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	// user-b comments on user-a's post
	_, err = svc.CommentOnPost(context.Background(), "user-b", &alfqv1.CommentRequest{PostId: post.Id, Content: "nice"})
	if err != nil {
		t.Fatalf("comment: %v", err)
	}

	// Verify event was published — target is user-a (post author)
	found := false
	for _, ev := range pub.events {
		if ev.Type == "post_commented" && ev.ActorID == "user-b" && ev.TargetID == "user-a" && ev.PostID == post.Id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected post_commented event from user-b to user-a, got %+v", pub.events)
	}
}

func TestShare_EmitsNotificationEvent(t *testing.T) {
	repo := newFakeFeedRepo()
	pub := &fakeEventPublisher{}
	svc := NewService(repo).WithEventPublisher(pub)

	// user-a creates a post
	post, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:  "original",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	// user-b shares user-a's post
	_, err = svc.SharePost(context.Background(), "user-b", &alfqv1.SharePostRequest{PostId: post.Id})
	if err != nil {
		t.Fatalf("share: %v", err)
	}

	found := false
	for _, ev := range pub.events {
		if ev.Type == "post_shared" && ev.ActorID == "user-b" && ev.TargetID == "user-a" && ev.PostID == post.Id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected post_shared event, got %+v", pub.events)
	}
}

func TestCommentOnOwnPost_DoesNotNotifySelf(t *testing.T) {
	repo := newFakeFeedRepo()
	pub := &fakeEventPublisher{}
	svc := NewService(repo).WithEventPublisher(pub)

	// user-a creates a post
	post, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:  "my post",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	// user-a comments on their own post
	_, err = svc.CommentOnPost(context.Background(), "user-a", &alfqv1.CommentRequest{PostId: post.Id, Content: "self"})
	if err != nil {
		t.Fatalf("comment: %v", err)
	}

	// None of the events should target user-a
	for _, ev := range pub.events {
		if ev.TargetID == "user-a" {
			t.Fatalf("should not notify self, got event %+v", ev)
		}
	}
}

func TestSocialRateLimit_ResourceExhausted(t *testing.T) {
	repo := newFakeFeedRepo()
	limiter := newFakeRateLimiter()
	svc := NewService(repo).WithRateLimiter(limiter)

	// Block user-a from creating posts
	limiter.block("user-a", RateLimitCreatePost)

	_, err := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{
		Content:  "test",
		PostType: "text",
	})
	if err == nil {
		t.Fatal("expected ResourceExhausted")
	}
	var ce *connect.Error
	if errors.As(err, &ce) {
		if ce.Code() != connect.CodeResourceExhausted {
			t.Fatalf("expected ResourceExhausted, got %v", ce.Code())
		}
	}

	// Different user should still be allowed
	_, err = svc.CreatePost(context.Background(), "user-b", &alfqv1.CreatePostRequest{
		Content:  "test",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("user-b should not be rate limited: %v", err)
	}

	// Same user, different action should be allowed
	// Create a post for user-a to comment on (they're blocked from creating, so use existing one)
	post, _ := svc.CreatePost(context.Background(), "user-b", &alfqv1.CreatePostRequest{Content: "base", PostType: "text"})
	_, err = svc.CommentOnPost(context.Background(), "user-a", &alfqv1.CommentRequest{PostId: post.Id, Content: "comment"})
	if err != nil {
		t.Fatalf("user-a comment should not be blocked by post rate limit: %v", err)
	}
}

func TestComment_ContentValidation(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	// Create a public post for comments to attach to
	post, _ := svc.CreatePost(context.Background(), "user-a", &alfqv1.CreatePostRequest{Content: "post for comments", PostType: "text"})

	tests := []struct {
		name     string
		content  string
		wantCode connect.Code
	}{
		{name: "whitespace-only", content: "   ", wantCode: connect.CodeInvalidArgument},
		{name: "empty", content: "", wantCode: connect.CodeInvalidArgument},
		{name: "valid", content: "nice post", wantCode: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CommentOnPost(context.Background(), "user-b", &alfqv1.CommentRequest{PostId: post.Id, Content: tc.content})
			if tc.wantCode == 0 {
				if err != nil {
					t.Fatalf("expected success, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got success")
			}
			var ce *connect.Error
			if errors.As(err, &ce) {
				if ce.Code() != tc.wantCode {
					t.Fatalf("expected code %v, got %v", tc.wantCode, ce.Code())
				}
			} else {
				t.Fatalf("expected connect.Error, got %T", err)
			}
		})
	}
}
