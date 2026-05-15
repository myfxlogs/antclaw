package feed

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// fakeFeedRepo implements FeedRepository with in-memory maps for testing.
type fakeFeedRepo struct {
	posts      map[string]*postgres.FeedPostRow
	comments   map[string]*postgres.FeedCommentRow
	likes      map[string]map[string]bool // postID -> userID -> true
	postIdx    int
	commentIdx int
}

func newFakeFeedRepo() *fakeFeedRepo {
	return &fakeFeedRepo{
		posts:    make(map[string]*postgres.FeedPostRow),
		comments: make(map[string]*postgres.FeedCommentRow),
		likes:    make(map[string]map[string]bool),
	}
}

func (f *fakeFeedRepo) newPostID() string {
	f.postIdx++
	return formatID("post", f.postIdx)
}

func (f *fakeFeedRepo) newCommentID() string {
	f.commentIdx++
	return formatID("comment", f.commentIdx)
}

func formatID(prefix string, idx int) string {
	return fmt.Sprintf("%s-%04d", prefix, idx)
}

func (f *fakeFeedRepo) CreatePost(_ context.Context, row *postgres.FeedPostRow) (*postgres.FeedPostRow, error) {
	row.ID = f.newPostID()
	row.CreatedAt = time.Now()
	row.LikeCount = 0
	row.CommentCount = 0
	row.ShareCount = 0
	f.posts[row.ID] = row
	return row, nil
}

func (f *fakeFeedRepo) GetPost(_ context.Context, postID string, currentUserID string) (*postgres.FeedPostRow, []string, error) {
	row, ok := f.posts[postID]
	if !ok {
		return nil, nil, errors.New("not found")
	}
	var likedBy []string
	if currentUserID != "" && f.isLikedBy(postID, currentUserID) {
		likedBy = []string{currentUserID}
	}
	return row, likedBy, nil
}

func (f *fakeFeedRepo) CheckPostExists(_ context.Context, postID string) (bool, error) {
	_, ok := f.posts[postID]
	return ok, nil
}

func (f *fakeFeedRepo) CheckPostVisibility(_ context.Context, postID, currentUserID string) (bool, error) {
	row, ok := f.posts[postID]
	if !ok {
		return false, nil
	}
	if row.Visibility == "public" {
		return true, nil
	}
	if row.AuthorID == currentUserID {
		return true, nil
	}
	return false, nil
}

func (f *fakeFeedRepo) GetFeed(_ context.Context, filter string, cursor *postgres.SocialCursor, limit int32, currentUserID string) ([]*postgres.FeedPostRow, [][]string, *postgres.SocialCursor, error) {
	var result []*postgres.FeedPostRow
	for _, row := range f.posts {
		if row.Visibility != "public" {
			continue
		}
		if filter == "signals_only" && row.PostType != "signal_card" {
			continue
		}
		if filter == "posts_only" && row.PostType != "text" {
			continue
		}
		if filter == "shares" && row.PostType != "share" {
			continue
		}
		if cursor != nil {
			if row.CreatedAt.After(cursor.CreatedAt) {
				continue
			}
			if row.CreatedAt.Equal(cursor.CreatedAt) && row.ID >= cursor.ID {
				continue
			}
		}
		result = append(result, row)
	}
	// sort by created_at DESC, id DESC
	// Simple bubble for test; real code uses DB ORDER BY
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) ||
				(result[i].CreatedAt.Equal(result[j].CreatedAt) && result[i].ID < result[j].ID) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	hasMore := int32(len(result)) > limit
	if hasMore {
		result = result[:limit]
	}
	var nextCursor *postgres.SocialCursor
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		nextCursor = &postgres.SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	likedByList := make([][]string, len(result))
	for i, row := range result {
		if currentUserID != "" && f.isLikedBy(row.ID, currentUserID) {
			likedByList[i] = []string{currentUserID}
		}
	}
	return result, likedByList, nextCursor, nil
}

func (f *fakeFeedRepo) ListUserPosts(_ context.Context, userID, filter string, cursor *postgres.SocialCursor, limit int32, currentUserID string) ([]*postgres.FeedPostRow, [][]string, *postgres.SocialCursor, error) {
	var result []*postgres.FeedPostRow
	for _, row := range f.posts {
		if row.AuthorID != userID {
			continue
		}
		if filter == "signals_only" && row.PostType != "signal_card" {
			continue
		}
		if filter == "posts_only" && row.PostType != "text" {
			continue
		}
		if filter == "shares" && row.PostType != "share" {
			continue
		}
		if cursor != nil {
			if row.CreatedAt.After(cursor.CreatedAt) {
				continue
			}
			if row.CreatedAt.Equal(cursor.CreatedAt) && row.ID >= cursor.ID {
				continue
			}
		}
		result = append(result, row)
	}
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].CreatedAt.Before(result[j].CreatedAt) ||
				(result[i].CreatedAt.Equal(result[j].CreatedAt) && result[i].ID < result[j].ID) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	hasMore := int32(len(result)) > limit
	if hasMore {
		result = result[:limit]
	}
	var nextCursor *postgres.SocialCursor
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		nextCursor = &postgres.SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	likedByList := make([][]string, len(result))
	for i, row := range result {
		if currentUserID != "" && f.isLikedBy(row.ID, currentUserID) {
			likedByList[i] = []string{currentUserID}
		}
	}
	return result, likedByList, nextCursor, nil
}

func (f *fakeFeedRepo) LikePost(_ context.Context, postID, userID string) error {
	if _, ok := f.likes[postID]; !ok {
		f.likes[postID] = make(map[string]bool)
	}
	f.likes[postID][userID] = true
	return nil
}

func (f *fakeFeedRepo) UnlikePost(_ context.Context, postID, userID string) error {
	if m, ok := f.likes[postID]; ok {
		delete(m, userID)
	}
	return nil
}

func (f *fakeFeedRepo) isLikedBy(postID, userID string) bool {
	m, ok := f.likes[postID]
	if !ok {
		return false
	}
	return m[userID]
}

func (f *fakeFeedRepo) GetLikedByUser(_ context.Context, postID, userID string) (bool, error) {
	return f.isLikedBy(postID, userID), nil
}

func (f *fakeFeedRepo) CreateComment(_ context.Context, row *postgres.FeedCommentRow) (*postgres.FeedCommentRow, error) {
	row.ID = f.newCommentID()
	row.CreatedAt = time.Now()
	f.comments[row.ID] = row
	return row, nil
}

func (f *fakeFeedRepo) ListComments(_ context.Context, postID string, cursor *postgres.SocialCursor, limit int32) ([]*postgres.FeedCommentRow, *postgres.SocialCursor, error) {
	var result []*postgres.FeedCommentRow
	for _, c := range f.comments {
		if c.PostID != postID {
			continue
		}
		if cursor != nil {
			if c.CreatedAt.Before(cursor.CreatedAt) {
				continue
			}
			if c.CreatedAt.Equal(cursor.CreatedAt) && c.ID <= cursor.ID {
				continue
			}
		}
		result = append(result, c)
	}
	hasMore := int32(len(result)) > limit
	if hasMore {
		result = result[:limit]
	}
	var nextCursor *postgres.SocialCursor
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		nextCursor = &postgres.SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return result, nextCursor, nil
}

func (f *fakeFeedRepo) CheckCommentPostID(_ context.Context, commentID, postID string) (bool, error) {
	c, ok := f.comments[commentID]
	if !ok {
		return false, nil
	}
	return c.PostID == postID, nil
}

func (f *fakeFeedRepo) GetUserName(_ context.Context, userID string) (string, error) {
	return "user-" + userID, nil
}

// ----- Helpers -----

func mustCreatePost(t *testing.T, svc *Service, userID string, req *alfqv1.CreatePostRequest) *alfqv1.Post {
	t.Helper()
	p, err := svc.CreatePost(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("CreatePost(%s): %v", userID, err)
	}
	return p
}

func mustComment(t *testing.T, svc *Service, userID string, req *alfqv1.CommentRequest) *alfqv1.Comment {
	t.Helper()
	c, err := svc.CommentOnPost(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("CommentOnPost(%s): %v", userID, err)
	}
	return c
}

func mustLike(t *testing.T, svc *Service, userID, postID string) {
	t.Helper()
	if _, err := svc.LikePost(context.Background(), userID, &alfqv1.LikePostRequest{PostId: postID}); err != nil {
		t.Fatalf("LikePost(%s, %s): %v", userID, postID, err)
	}
}

// ----- Tests -----

func TestCreatePost_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.CreatePost(context.Background(), "", &alfqv1.CreatePostRequest{
		Content: "hello",
	})
	if err == nil {
		t.Fatal("expected Unauthenticated error")
	}
	if ce := new(connect.Error); errors.As(err, &ce) {
		if ce.Code() != connect.CodeUnauthenticated {
			t.Fatalf("expected Unauthenticated, got %v", ce.Code())
		}
	} else {
		t.Fatalf("expected connect.Error, got %T", err)
	}
}

func TestCreatePost_Success(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	p, err := svc.CreatePost(context.Background(), "user-1", &alfqv1.CreatePostRequest{
		Content:  "hello world",
		PostType: "text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.AuthorId != "user-1" {
		t.Fatalf("expected author user-1, got %s", p.AuthorId)
	}
	if p.Content != "hello world" {
		t.Fatalf("expected content 'hello world', got %s", p.Content)
	}
}

func TestCreatePost_CircleUnsupported(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.CreatePost(context.Background(), "user-1", &alfqv1.CreatePostRequest{
		Content:    "hello",
		Visibility: "circle",
	})
	if err == nil {
		t.Fatal("expected InvalidArgument for circle visibility")
	}
}

func TestGetFeed_FilterAll(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	// Seed posts
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "p1", PostType: "text"})
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "p2", PostType: "signal_card"})
	mustCreatePost(t, svc, "user-2", &alfqv1.CreatePostRequest{Content: "p3", PostType: "share", Visibility: "followers"})

	resp, err := svc.GetFeed(context.Background(), "", &alfqv1.GetFeedRequest{Filter: "all", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only public posts returned (p1, p2), p3 is followers-only
	if len(resp.Posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(resp.Posts))
	}
}

func TestGetFeed_FilterSignalsOnly(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "text post", PostType: "text"})
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "signal", PostType: "signal_card"})

	resp, err := svc.GetFeed(context.Background(), "", &alfqv1.GetFeedRequest{Filter: "signals_only", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("expected 1 signal_card, got %d", len(resp.Posts))
	}
	if resp.Posts[0].PostType != "signal_card" {
		t.Fatalf("expected signal_card, got %s", resp.Posts[0].PostType)
	}
}

func TestGetFeed_FilterPostsOnly(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "text post", PostType: "text"})
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "signal", PostType: "signal_card"})

	resp, err := svc.GetFeed(context.Background(), "", &alfqv1.GetFeedRequest{Filter: "posts_only", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Fatalf("expected 1 text post, got %d", len(resp.Posts))
	}
	if resp.Posts[0].PostType != "text" {
		t.Fatalf("expected text, got %s", resp.Posts[0].PostType)
	}
}

func TestGetPost_FieldsConsistentWithGetFeed(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})

	got, err := svc.GetPost(context.Background(), "", &alfqv1.GetPostRequest{PostId: created.Id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Id != created.Id {
		t.Fatalf("expected id %s, got %s", created.Id, got.Id)
	}
	if got.Content != created.Content {
		t.Fatalf("content mismatch")
	}
}

func TestLikePost_Idempotent(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})

	p1, err := svc.LikePost(context.Background(), "user-2", &alfqv1.LikePostRequest{PostId: created.Id})
	if err != nil {
		t.Fatalf("first like error: %v", err)
	}
	p2, err := svc.LikePost(context.Background(), "user-2", &alfqv1.LikePostRequest{PostId: created.Id})
	if err != nil {
		t.Fatalf("second like error: %v", err)
	}
	if p1.Id != p2.Id {
		t.Fatalf("id mismatch on idempotent like")
	}
}

func TestUnlikePost_Idempotent(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})

	// Unlike without like should succeed
	_, err := svc.UnlikePost(context.Background(), "user-2", &alfqv1.UnlikePostRequest{PostId: created.Id})
	if err != nil {
		t.Fatalf("unlike without like error: %v", err)
	}
	// Like then unlike
	mustLike(t, svc, "user-2", created.Id)
	_, err = svc.UnlikePost(context.Background(), "user-2", &alfqv1.UnlikePostRequest{PostId: created.Id})
	if err != nil {
		t.Fatalf("unlike after like error: %v", err)
	}
}

func TestLikePost_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.LikePost(context.Background(), "", &alfqv1.LikePostRequest{PostId: "post-1"})
	if err == nil {
		t.Fatal("expected Unauthenticated")
	}
}

func TestCommentOnPost_Success(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})

	comment, err := svc.CommentOnPost(context.Background(), "user-2", &alfqv1.CommentRequest{
		PostId:  created.Id,
		Content: "nice post",
	})
	if err != nil {
		t.Fatalf("comment error: %v", err)
	}
	if comment.PostId != created.Id {
		t.Fatalf("expected post_id %s, got %s", created.Id, comment.PostId)
	}
	if comment.Content != "nice post" {
		t.Fatalf("content mismatch")
	}
}

func TestCommentOnPost_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.CommentOnPost(context.Background(), "", &alfqv1.CommentRequest{
		PostId:  "post-1",
		Content: "hi",
	})
	if err == nil {
		t.Fatal("expected Unauthenticated")
	}
}

func TestCommentOnPost_ParentCommentInvalidPost(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	post1 := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "post1", PostType: "text"})
	post2 := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "post2", PostType: "text"})
	// Comment on post1
	c1 := mustComment(t, svc, "user-2", &alfqv1.CommentRequest{PostId: post1.Id, Content: "c1"})
	// Try to use c1 as parent for post2
	_, err := svc.CommentOnPost(context.Background(), "user-2", &alfqv1.CommentRequest{
		PostId:          post2.Id,
		Content:         "reply",
		ParentCommentId: c1.Id,
	})
	if err == nil {
		t.Fatal("expected InvalidArgument for cross-post parent comment")
	}
}

func TestListComments_AfterComment(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	created := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "test", PostType: "text"})
	mustComment(t, svc, "user-2", &alfqv1.CommentRequest{PostId: created.Id, Content: "c1"})
	mustComment(t, svc, "user-3", &alfqv1.CommentRequest{PostId: created.Id, Content: "c2"})

	resp, err := svc.ListComments(context.Background(), "", &alfqv1.ListCommentsRequest{PostId: created.Id, PageSize: 50})
	if err != nil {
		t.Fatalf("list comments error: %v", err)
	}
	if len(resp.Comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(resp.Comments))
	}
}

func TestSharePost_WithOriginalPostID(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	original := mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "original", PostType: "text"})

	shared, err := svc.SharePost(context.Background(), "user-2", &alfqv1.SharePostRequest{
		PostId:  original.Id,
		Comment: "check this out",
	})
	if err != nil {
		t.Fatalf("share error: %v", err)
	}
	if shared.PostType != "share" {
		t.Fatalf("expected share type, got %s", shared.PostType)
	}
	if shared.OriginalPostId != original.Id {
		t.Fatalf("expected original_post_id %s, got %s", original.Id, shared.OriginalPostId)
	}
	if shared.Content != "check this out" {
		t.Fatalf("content mismatch")
	}
}

func TestSharePost_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.SharePost(context.Background(), "", &alfqv1.SharePostRequest{PostId: "post-1"})
	if err == nil {
		t.Fatal("expected Unauthenticated")
	}
}

func TestSharePost_OriginalNotFound(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.SharePost(context.Background(), "user-1", &alfqv1.SharePostRequest{PostId: "nonexistent"})
	if err == nil {
		t.Fatal("expected NotFound")
	}
}

func TestGetPost_NotFound(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.GetPost(context.Background(), "", &alfqv1.GetPostRequest{PostId: "nonexistent"})
	if err == nil {
		t.Fatal("expected NotFound")
	}
}

func TestListUserPosts_RequiresUserID(t *testing.T) {
	svc := NewService(newFakeFeedRepo())
	_, err := svc.ListUserPosts(context.Background(), "", &alfqv1.ListUserPostsRequest{UserId: ""})
	if err == nil {
		t.Fatal("expected InvalidArgument")
	}
}

func TestListUserPosts_Success(t *testing.T) {
	repo := newFakeFeedRepo()
	svc := NewService(repo)
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "p1", PostType: "text"})
	mustCreatePost(t, svc, "user-1", &alfqv1.CreatePostRequest{Content: "p2", PostType: "text"})
	mustCreatePost(t, svc, "user-2", &alfqv1.CreatePostRequest{Content: "p3", PostType: "text"})

	resp, err := svc.ListUserPosts(context.Background(), "", &alfqv1.ListUserPostsRequest{UserId: "user-1", PageSize: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Posts) != 2 {
		t.Fatalf("expected 2 posts for user-1, got %d", len(resp.Posts))
	}
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
