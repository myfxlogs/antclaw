package auth

import (
	"strings"
	"testing"
)

func TestGenerateCodeIDFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		id, err := GenerateCodeID(CodeIDDefaultDigits)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if len(id) != CodeIDDefaultDigits {
			t.Fatalf("len=%d, id=%q", len(id), id)
		}
		if id[0] == '0' {
			t.Fatalf("starts with 0: %q", id)
		}
		if strings.ContainsAny(id, "47") {
			t.Fatalf("contains 4/7: %q", id)
		}
		if err := ValidateCodeID(id); err != nil {
			t.Fatalf("validate %q: %v", id, err)
		}
	}
}

func TestValidateCodeIDRejects(t *testing.T) {
	bad := []string{
		"",          // empty
		"1234",      // too short
		"01234",     // starts with 0
		"14235",     // contains 4
		"17235",     // contains 7
		"abc12",     // non-digit
		"12345678901", // too long (>10)
	}
	for _, s := range bad {
		if err := ValidateCodeID(s); err == nil {
			t.Fatalf("expected reject %q", s)
		}
	}
	good := []string{"12356", "98521", "1230568901"}
	for _, s := range good {
		if err := ValidateCodeID(s); err != nil {
			t.Fatalf("expected accept %q: %v", s, err)
		}
	}
}

func TestIsAllDigits(t *testing.T) {
	cases := map[string]bool{
		"":         false,
		"12345":    true,
		"a@b.com":  false,
		"123 45":   false,
		"00000":    true,
	}
	for in, want := range cases {
		if got := IsAllDigits(in); got != want {
			t.Fatalf("IsAllDigits(%q)=%v, want %v", in, got, want)
		}
	}
}
