package user

import (
	"context"
	"errors"
	"testing"

	"github.com/antclaw/antclaw/internal/domain/apperror"
)

func TestGetMe_EmptyUserID(t *testing.T) {
	svc := NewService()
	_, err := svc.GetMe(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
	if !errors.Is(err, apperror.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestGetMe_UserNotFound(t *testing.T) {
	svc := NewService()
	_, err := svc.GetMe(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateSettings_EmptyUserID(t *testing.T) {
	svc := NewService()
	_, err := svc.UpdateSettings(context.Background(), "", "", 0, "")
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
	if !errors.Is(err, apperror.ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestSetAiKey_ReturnsError(t *testing.T) {
	svc := NewService()
	_, err := svc.SetAiKey(context.Background(), "u1", 0, "")
	if err == nil {
		t.Fatal("expected unimplemented error")
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
