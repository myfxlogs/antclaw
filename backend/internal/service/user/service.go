package user

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements User profile management business logic.
type Service struct {
	users         map[string]*userv1.User
	memberships   map[string]*userv1.Membership
	history       map[string][]*userv1.HistoryItem
	pins          map[string][]*userv1.Pin
	feedbackCount int
}

// NewService creates a new UserService.
func NewService() *Service {
	s := &Service{
		users:         make(map[string]*userv1.User),
		memberships:   make(map[string]*userv1.Membership),
		history:       make(map[string][]*userv1.HistoryItem),
		pins:          make(map[string][]*userv1.Pin),
		feedbackCount: 0,
	}

	// Pre-populate sample user
	now := time.Now().Unix()
	s.users["user-1"] = &userv1.User{
		UserId:        "user-1",
		Email:         "demo@antclaw.io",
		Username:      "demo_user",
		DisplayName:   "Demo User",
		Locale:        userv1.Locale_LOCALE_EN_US,
		Timezone:      "UTC",
		Roles:         []string{"user"},
		EmailVerified: true,
		CreatedAt:     now - 86400*30,
		UpdatedAt:     now,
	}
	s.memberships["user-1"] = &userv1.Membership{
		Tier:           userv1.MembershipTier_MEMBERSHIP_TIER_PREMIUM,
		ExpiresAt:      now + 86400*365,
		QuotaDaily:     1000,
		QuotaUsedToday: 45,
	}

	// Pre-populate sample history
	s.history["user-1"] = []*userv1.HistoryItem{
		{ItemId: "hist-1", Type: "query", Title: "EURUSD analysis", Payload: "{\"pair\":\"EURUSD\"}", CreatedAt: now - 3600},
		{ItemId: "hist-2", Type: "chart", Title: "VIX breakdown", Payload: "{\"indicator\":\"VIX\"}", CreatedAt: now - 7200},
	}

	return s
}

// GetMe returns the current user profile.
func (s *Service) GetMe(ctx context.Context, userID string) (*userv1.GetMeResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	user, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	return &userv1.GetMeResponse{User: user}, nil
}

// UpdateSettings updates user settings.
func (s *Service) UpdateSettings(ctx context.Context, userID, displayName string, locale userv1.Locale, timezone string) (*userv1.UpdateSettingsResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	user, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if displayName != "" {
		user.DisplayName = displayName
	}
	if locale != userv1.Locale_LOCALE_UNSPECIFIED {
		user.Locale = locale
	}
	if timezone != "" {
		user.Timezone = timezone
	}
	user.UpdatedAt = time.Now().Unix()
	return &userv1.UpdateSettingsResponse{User: user}, nil
}

// GetMembership returns membership info.
func (s *Service) GetMembership(ctx context.Context, userID string) (*userv1.GetMembershipResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	membership, ok := s.memberships[userID]
	if !ok {
		membership = &userv1.Membership{
			Tier:           userv1.MembershipTier_MEMBERSHIP_TIER_FREE,
			ExpiresAt:      0,
			QuotaDaily:     100,
			QuotaUsedToday: 0,
		}
	}
	return &userv1.GetMembershipResponse{Membership: membership}, nil
}

// StartOnboarding starts the onboarding flow.
func (s *Service) StartOnboarding(ctx context.Context) (*userv1.StartOnboardingResponse, error) {
	return &userv1.StartOnboardingResponse{
		OnboardingId: fmt.Sprintf("onb-%d", time.Now().Unix()),
		Steps: []string{
			"welcome",
			"select_pairs",
			"set_alerts",
			"connect_accounts",
			"complete",
		},
	}, nil
}

// GetHistory returns interaction history.
func (s *Service) GetHistory(ctx context.Context, userID, cursor string, pageSize int32) (*userv1.GetHistoryResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	items := s.history[userID]
	if items == nil {
		items = []*userv1.HistoryItem{}
	}
	return &userv1.GetHistoryResponse{
		Items:      items,
		NextCursor: "",
	}, nil
}

// ClearHistory clears interaction history.
func (s *Service) ClearHistory(ctx context.Context, userID string, all bool, types []string) (*userv1.ClearHistoryResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	count := int32(len(s.history[userID]))
	if all {
		s.history[userID] = []*userv1.HistoryItem{}
	}
	return &userv1.ClearHistoryResponse{ClearedCount: count}, nil
}

// ListPins lists pinned items.
func (s *Service) ListPins(ctx context.Context, userID string) (*userv1.ListPinsResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	pins := s.pins[userID]
	if pins == nil {
		pins = []*userv1.Pin{}
	}
	return &userv1.ListPinsResponse{Pins: pins}, nil
}

// Pin creates a new pin.
func (s *Service) Pin(ctx context.Context, userID, itemID, itemType, title string) (*userv1.PinResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	pin := &userv1.Pin{
		PinId:     fmt.Sprintf("pin-%d", time.Now().Unix()),
		ItemId:    itemID,
		ItemType:  itemType,
		Title:     title,
		CreatedAt: time.Now().Unix(),
	}
	s.pins[userID] = append(s.pins[userID], pin)
	return &userv1.PinResponse{Pin: pin}, nil
}

// Unpin removes a pin.
func (s *Service) Unpin(ctx context.Context, userID, pinID string) (*userv1.UnpinResponse, error) {
	if userID == "" {
		userID = "user-1"
	}
	var filtered []*userv1.Pin
	for _, pin := range s.pins[userID] {
		if pin.PinId != pinID {
			filtered = append(filtered, pin)
		}
	}
	s.pins[userID] = filtered
	return &userv1.UnpinResponse{}, nil
}

// SubmitFeedback submits user feedback.
func (s *Service) SubmitFeedback(ctx context.Context, category, content, contact string) (*userv1.SubmitFeedbackResponse, error) {
	s.feedbackCount++
	return &userv1.SubmitFeedbackResponse{
		FeedbackId: fmt.Sprintf("fb-%d", s.feedbackCount),
	}, nil
}

// SetAiKey sets user AI key (BYOK).
func (s *Service) SetAiKey(ctx context.Context, userID string, provider userv1.AiProvider, apiKey string) (*userv1.SetAiKeyResponse, error) {
	// In a real implementation, this would securely store the API key
	return &userv1.SetAiKeyResponse{Success: true}, nil
}
