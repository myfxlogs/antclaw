package macro

import (
	"context"
	"errors"
	"testing"

	"github.com/antclaw/antclaw/internal/domain/apperror"
)

func TestGetFred_ProviderNotConfigured(t *testing.T) {
	svc := NewService() // no FRED client
	_, err := svc.GetFred(context.Background(), "GDP")
	if err == nil {
		t.Fatal("expected error when FRED not configured")
	}
	if !errors.Is(err, apperror.ErrProviderNotConfigured) {
		t.Fatalf("expected ErrProviderNotConfigured, got %v", err)
	}
}

func TestNewService_NoPanic(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}
