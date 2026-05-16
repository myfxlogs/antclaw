package calendar

import (
	"context"
	"errors"
	"testing"

	"github.com/antclaw/antclaw/internal/domain/apperror"
)

func TestGetEvent_EmptyEventID(t *testing.T) {
	svc := NewService()
	_, err := svc.GetEvent(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty event_id")
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetEvent(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent event")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetImpact_NotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetImpact(context.Background(), "evt-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetImpactHistory_DataInsufficient(t *testing.T) {
	svc := NewService()
	_, err := svc.GetImpactHistory(context.Background(), "NFPs", "EURUSD", 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, apperror.ErrDataInsufficient) {
		t.Fatalf("expected ErrDataInsufficient, got %v", err)
	}
}
