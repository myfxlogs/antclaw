package rpc

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/antclaw/antclaw/internal/auth"
)

func TestRequireAdminNoContext(t *testing.T) {
	ctx := context.Background()
	err := requireAdmin(ctx)
	if err == nil {
		t.Fatal("expected error when no role in context")
	}
	ce, ok := err.(*connect.Error)
	if !ok || ce.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestRequireAdminNormalUser(t *testing.T) {
	ctx := auth.WithRole(context.Background(), "user")
	err := requireAdmin(ctx)
	if err == nil {
		t.Fatal("expected error for normal user role")
	}
}

func TestRequireAdminAdminUser(t *testing.T) {
	ctx := auth.WithRole(context.Background(), "admin")
	if err := requireAdmin(ctx); err != nil {
		t.Fatalf("unexpected error for admin: %v", err)
	}
}

func TestRequireAdminSuperAdminUser(t *testing.T) {
	ctx := auth.WithRole(context.Background(), "super_admin")
	if err := requireAdmin(ctx); err != nil {
		t.Fatalf("unexpected error for super_admin: %v", err)
	}
}

func TestClampPageDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    int32
		expected int32
	}{
		{"zero", 0, 50},
		{"negative", -5, 50},
		{"too_large", 200, 50},
		{"normal", 20, 20},
		{"edge", 100, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampPage(tt.input)
			if got != tt.expected {
				t.Errorf("clampPage(%d) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestReportDeviceInfoRequiresAuth(t *testing.T) {
	// Device handler requires pgxpool, but we can verify auth middleware is configured
	// via main.go integration.  This test documents the expected behaviour.
	// 完整认证+绑定逻辑需要在集成测试或 smoke 脚本中验证。
	t.Skip("requires postgres — tested via smoke-rpc.sh / e2e")
}
