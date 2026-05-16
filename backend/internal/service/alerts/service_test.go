package alerts

import (
	"context"
	"testing"
)

func TestGate_Decide_NilPool(t *testing.T) {
	g := NewGate(nil)
	d := g.Decide(context.Background(), "test_user", "price", "high", []string{"EURUSD"})
	if !d.Send {
		t.Fatal("expected ok=true with nil pool")
	}
}

func TestGate_Decide_Multiple(t *testing.T) {
	g := NewGate(nil)
	categories := []string{"price", "signal", "macro"}

	for _, cat := range categories {
		d := g.Decide(context.Background(), "u1", cat, "medium", []string{"EURUSD"})
		if !d.Send {
			t.Fatalf("expected ok=true for %s", cat)
		}
	}
}
