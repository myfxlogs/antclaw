package trader

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// fakeTraderRepo implements TraderRepository with in-memory maps.
type fakeTraderRepo struct {
	profiles map[string]*postgres.TraderProfileRow
	follows  map[string]map[string]bool // followerID -> followingID -> true
}

func newFakeTraderRepo() *fakeTraderRepo {
	return &fakeTraderRepo{
		profiles: make(map[string]*postgres.TraderProfileRow),
		follows:  make(map[string]map[string]bool),
	}
}

func (f *fakeTraderRepo) seedUser(userID string) {
	f.profiles[userID] = &postgres.TraderProfileRow{
		UserID:         userID,
		DisplayName:    "User-" + userID,
		Tier:           "normal",
		FollowerCount:  0,
		FollowingCount: 0,
		CreatedAt:      time.Now(),
	}
}

func (f *fakeTraderRepo) GetProfile(_ context.Context, userID string) (*postgres.TraderProfileRow, error) {
	row, ok := f.profiles[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return row, nil
}

func (f *fakeTraderRepo) UpdateProfile(_ context.Context, userID string, displayName string) error {
	if row, ok := f.profiles[userID]; ok {
		row.DisplayName = displayName
	}
	return nil
}

func (f *fakeTraderRepo) CheckUserExists(_ context.Context, userID string) (bool, error) {
	_, ok := f.profiles[userID]
	return ok, nil
}

func (f *fakeTraderRepo) GetUserName(_ context.Context, userID string) (string, error) {
	if row, ok := f.profiles[userID]; ok {
		return row.DisplayName, nil
	}
	return "", errors.New("not found")
}

func (f *fakeTraderRepo) Follow(_ context.Context, followerID, followingID string) error {
	if _, ok := f.follows[followerID]; !ok {
		f.follows[followerID] = make(map[string]bool)
	}
	f.follows[followerID][followingID] = true
	return nil
}

func (f *fakeTraderRepo) Unfollow(_ context.Context, followerID, followingID string) error {
	if m, ok := f.follows[followerID]; ok {
		delete(m, followingID)
	}
	return nil
}

func (f *fakeTraderRepo) CheckFollowExists(_ context.Context, followerID, followingID string) (bool, error) {
	return f.isFollowing(followerID, followingID), nil
}

func (f *fakeTraderRepo) isFollowing(followerID, followingID string) bool {
	m, ok := f.follows[followerID]
	if !ok {
		return false
	}
	return m[followingID]
}

func (f *fakeTraderRepo) GetFollowerCount(_ context.Context, userID string) (int32, error) {
	var cnt int32
	for _, m := range f.follows {
		if m[userID] {
			cnt++
		}
	}
	return cnt, nil
}

func (f *fakeTraderRepo) GetFollowingCount(_ context.Context, userID string) (int32, error) {
	m, ok := f.follows[userID]
	if !ok {
		return 0, nil
	}
	return int32(len(m)), nil
}

func (f *fakeTraderRepo) GetFollowers(_ context.Context, userID string, cursor *postgres.SocialCursor, limit int32) ([]*postgres.UserInfoRow, *postgres.SocialCursor, error) {
	var result []*postgres.UserInfoRow
	for followerID, following := range f.follows {
		if following[userID] {
			result = append(result, &postgres.UserInfoRow{
				UserID:      followerID,
				DisplayName: "User-" + followerID,
				Tier:        "normal",
			})
		}
	}
	hasMore := int32(len(result)) > limit
	if hasMore {
		result = result[:limit]
	}
	var nextCursor *postgres.SocialCursor
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		nextCursor = &postgres.SocialCursor{CreatedAt: time.Time{}, ID: last.UserID}
	}
	return result, nextCursor, nil
}

func (f *fakeTraderRepo) GetFollowing(_ context.Context, userID string, cursor *postgres.SocialCursor, limit int32) ([]*postgres.UserInfoRow, *postgres.SocialCursor, error) {
	var result []*postgres.UserInfoRow
	m, ok := f.follows[userID]
	if !ok {
		return nil, nil, nil
	}
	for followingID := range m {
		result = append(result, &postgres.UserInfoRow{
			UserID:      followingID,
			DisplayName: "User-" + followingID,
			Tier:        "normal",
		})
	}
	hasMore := int32(len(result)) > limit
	if hasMore {
		result = result[:limit]
	}
	var nextCursor *postgres.SocialCursor
	if hasMore && len(result) > 0 {
		last := result[len(result)-1]
		nextCursor = &postgres.SocialCursor{CreatedAt: time.Time{}, ID: last.UserID}
	}
	return result, nextCursor, nil
}

func (f *fakeTraderRepo) IsFollowing(_ context.Context, currentUserID, targetUserID string) (bool, error) {
	return f.isFollowing(currentUserID, targetUserID), nil
}

func (f *fakeTraderRepo) ListRecommendedTraders(_ context.Context, cursor *postgres.SocialCursor, limit int32) ([]*postgres.UserInfoRow, *postgres.SocialCursor, error) {
	var result []*postgres.UserInfoRow
	for userID, p := range f.profiles {
		result = append(result, &postgres.UserInfoRow{
			UserID:      userID,
			DisplayName: p.DisplayName,
			Tier:        p.Tier,
		})
	}
	// Sort by follower_count DESC for test fidelity (optional — tests don't depend on order)
	return result, nil, nil
}

// ----- Helpers -----

func mustFollow(t *testing.T, svc *Service, followerID, targetID string) {
	t.Helper()
	if _, err := svc.Follow(context.Background(), followerID, &alfqv1.FollowRequest{TargetUserId: targetID}); err != nil {
		t.Fatalf("Follow(%s, %s): %v", followerID, targetID, err)
	}
}

// ----- Tests -----

func TestGetProfile_UserNotFound(t *testing.T) {
	svc := NewService(newFakeTraderRepo())
	_, err := svc.GetProfile(context.Background(), "", &alfqv1.GetTraderProfileRequest{UserId: "noone"})
	if err == nil {
		t.Fatal("expected NotFound")
	}
	if ce := new(connect.Error); errors.As(err, &ce) {
		if ce.Code() != connect.CodeNotFound {
			t.Fatalf("expected NotFound, got %v", ce.Code())
		}
	}
}

func TestGetProfile_Success(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	svc := NewService(repo)
	p, err := svc.GetProfile(context.Background(), "", &alfqv1.GetTraderProfileRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.UserId != "user-1" {
		t.Fatalf("expected user-1, got %s", p.UserId)
	}
}

func TestFollow_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeTraderRepo())
	_, err := svc.Follow(context.Background(), "", &alfqv1.FollowRequest{TargetUserId: "user-2"})
	if err == nil {
		t.Fatal("expected Unauthenticated")
	}
}

func TestFollow_SelfFollow(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	svc := NewService(repo)
	_, err := svc.Follow(context.Background(), "user-1", &alfqv1.FollowRequest{TargetUserId: "user-1"})
	if err == nil {
		t.Fatal("expected InvalidArgument for self-follow")
	}
}

func TestFollow_IdempotentAndFollowerCount(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	repo.seedUser("user-2")
	svc := NewService(repo)

	resp1, err := svc.Follow(context.Background(), "user-1", &alfqv1.FollowRequest{TargetUserId: "user-2"})
	if err != nil {
		t.Fatalf("first follow error: %v", err)
	}
	if !resp1.Success {
		t.Fatal("expected success")
	}
	if resp1.FollowerCount != 1 {
		t.Fatalf("expected follower_count 1, got %d", resp1.FollowerCount)
	}

	resp2, err := svc.Follow(context.Background(), "user-1", &alfqv1.FollowRequest{TargetUserId: "user-2"})
	if err != nil {
		t.Fatalf("second follow error: %v", err)
	}
	if !resp2.Success {
		t.Fatal("expected idempotent success")
	}
	if resp2.FollowerCount != 1 {
		t.Fatalf("expected follower_count still 1, got %d", resp2.FollowerCount)
	}
}

func TestFollow_TargetUserNotFound(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	svc := NewService(repo)
	_, err := svc.Follow(context.Background(), "user-1", &alfqv1.FollowRequest{TargetUserId: "noone"})
	if err == nil {
		t.Fatal("expected NotFound")
	}
}

func TestUnfollow_Idempotent(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	repo.seedUser("user-2")
	svc := NewService(repo)

	// Unfollow when not following should succeed
	resp, err := svc.Unfollow(context.Background(), "user-1", &alfqv1.UnfollowRequest{TargetUserId: "user-2"})
	if err != nil {
		t.Fatalf("unfollow without follow error: %v", err)
	}
	if resp.FollowerCount != 0 {
		t.Fatalf("expected follower_count 0, got %d", resp.FollowerCount)
	}

	// Follow then unfollow
	mustFollow(t, svc, "user-1", "user-2")
	resp, err = svc.Unfollow(context.Background(), "user-1", &alfqv1.UnfollowRequest{TargetUserId: "user-2"})
	if err != nil {
		t.Fatalf("unfollow after follow error: %v", err)
	}
	if resp.FollowerCount != 0 {
		t.Fatalf("expected follower_count 0 after unfollow, got %d", resp.FollowerCount)
	}
}

func TestGetFollowers_Pagination(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("target")
	repo.seedUser("f1")
	repo.seedUser("f2")
	repo.seedUser("f3")
	svc := NewService(repo)

	mustFollow(t, svc, "f1", "target")
	mustFollow(t, svc, "f2", "target")
	mustFollow(t, svc, "f3", "target")

	resp, err := svc.GetFollowers(context.Background(), &alfqv1.GetFollowersRequest{UserId: "target", PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("expected 2 followers (page 1), got %d", len(resp.Users))
	}
	if resp.NextCursor == "" {
		t.Fatal("expected next_cursor for page 2")
	}
}

func TestGetFollowing_Pagination(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	repo.seedUser("u2")
	repo.seedUser("u3")
	repo.seedUser("u4")
	svc := NewService(repo)

	mustFollow(t, svc, "user-1", "u2")
	mustFollow(t, svc, "user-1", "u3")
	mustFollow(t, svc, "user-1", "u4")

	resp, err := svc.GetFollowing(context.Background(), &alfqv1.GetFollowingRequest{UserId: "user-1", PageSize: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("expected 2 following (page 1), got %d", len(resp.Users))
	}
}

func TestIsFollowing_Accurate(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-a")
	repo.seedUser("user-b")
	svc := NewService(repo)

	// Before follow: is_following should be false
	p, err := svc.GetProfile(context.Background(), "user-a", &alfqv1.GetTraderProfileRequest{UserId: "user-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IsFollowing {
		t.Fatal("expected is_following false before follow")
	}

	// After follow: is_following should be true
	mustFollow(t, svc, "user-a", "user-b")
	p, err = svc.GetProfile(context.Background(), "user-a", &alfqv1.GetTraderProfileRequest{UserId: "user-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsFollowing {
		t.Fatal("expected is_following true after follow")
	}

	// Viewing own profile: is_following should be false
	repo.seedUser("user-a")
	p, err = svc.GetProfile(context.Background(), "user-a", &alfqv1.GetTraderProfileRequest{UserId: "user-a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IsFollowing {
		t.Fatal("expected is_following false for self")
	}
}

func TestUpdateProfile_Unauthenticated(t *testing.T) {
	svc := NewService(newFakeTraderRepo())
	_, err := svc.UpdateProfile(context.Background(), "", &alfqv1.UpdateTraderProfileRequest{DisplayName: "new"})
	if err == nil {
		t.Fatal("expected Unauthenticated")
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := newFakeTraderRepo()
	repo.seedUser("user-1")
	svc := NewService(repo)
	p, err := svc.UpdateProfile(context.Background(), "user-1", &alfqv1.UpdateTraderProfileRequest{DisplayName: "NewName"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DisplayName != "NewName" {
		t.Fatalf("expected NewName, got %s", p.DisplayName)
	}
}

func TestGetFollowers_RequiresUserID(t *testing.T) {
	svc := NewService(newFakeTraderRepo())
	_, err := svc.GetFollowers(context.Background(), &alfqv1.GetFollowersRequest{UserId: ""})
	if err == nil {
		t.Fatal("expected InvalidArgument")
	}
}

func TestGetFollowing_RequiresUserID(t *testing.T) {
	svc := NewService(newFakeTraderRepo())
	_, err := svc.GetFollowing(context.Background(), &alfqv1.GetFollowingRequest{UserId: ""})
	if err == nil {
		t.Fatal("expected InvalidArgument")
	}
}
