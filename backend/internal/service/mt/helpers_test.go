package mt

import (
	"testing"

	mt4grpc "github.com/antclaw/antclaw/gen/go/third_party/mt4grpc"
	mt5grpc "github.com/antclaw/antclaw/gen/go/third_party/mt5grpc"

	"github.com/antclaw/antclaw/internal/domain"
)

func TestParseLoginInt32(t *testing.T) {
	tests := []struct {
		input    string
		expected int32
		wantErr  bool
	}{
		{"12345", 12345, false},
		{"0", 0, false},
		{"-1", -1, false},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		got, err := parseLoginInt32(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseLoginInt32(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("parseLoginInt32(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestParseLoginUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		wantErr  bool
	}{
		{"12345678901", 12345678901, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"-1", 0, true},
	}
	for _, tt := range tests {
		got, err := parseLoginUint64(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseLoginUint64(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("parseLoginUint64(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		input      string
		wantHost   string
		wantPort   int32
	}{
		{"mt4-demo.roboforex.com:443", "mt4-demo.roboforex.com", 443},
		{"95.217.147.61:443", "95.217.147.61", 443},
		{"localhost:8080", "localhost", 8080},
		{"host.without.port", "host.without.port", 443}, // default port
		{"host:badport", "host", 443},                    // invalid port → default
		{"", "", 443},
	}
	for _, tt := range tests {
		host, port := parseHostPort(tt.input)
		if host != tt.wantHost {
			t.Errorf("parseHostPort(%q) host = %q, want %q", tt.input, host, tt.wantHost)
		}
		if port != tt.wantPort {
			t.Errorf("parseHostPort(%q) port = %d, want %d", tt.input, port, tt.wantPort)
		}
	}
}

func TestCalcProfitPercent(t *testing.T) {
	tests := []struct {
		profit, balance float64
		expected        float64
	}{
		{100, 1000, 10.0},
		{0, 1000, 0},
		{-50, 1000, -5.0},
		{100, 0, 0}, // zero balance → zero percent
	}
	for _, tt := range tests {
		got := calcProfitPercent(tt.profit, tt.balance)
		if got != tt.expected {
			t.Errorf("calcProfitPercent(%f, %f) = %f, want %f", tt.profit, tt.balance, got, tt.expected)
		}
	}
}

func TestApplyMT4Summary(t *testing.T) {
	a := &domain.MTAccount{}
	s := &mt4grpc.AccountSummary{
		Balance:    10000.0,
		Credit:     500.0,
		Profit:     150.0,
		Equity:     10150.0,
		Margin:     2000.0,
		FreeMargin: 8150.0,
		MarginLevel: 507.5,
		Leverage:   100,
		Currency:   "USD",
		IsInvestor: false,
		Type:       mt4grpc.AccountType_AccountType_Demo,
	}
	applyMT4Summary(a, s)

	if a.Balance != 10000.0 {
		t.Errorf("Balance = %f, want 10000.0", a.Balance)
	}
	if a.Equity != 10150.0 {
		t.Errorf("Equity = %f, want 10150.0", a.Equity)
	}
	if a.Leverage != 100 {
		t.Errorf("Leverage = %d, want 100", a.Leverage)
	}
	if a.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", a.Currency)
	}
	if a.AccountType != "demo" {
		t.Errorf("AccountType = %q, want demo", a.AccountType)
	}
	if a.ProfitPercent != 1.5 {
		t.Errorf("ProfitPercent = %f, want 1.5", a.ProfitPercent)
	}
}

func TestApplyMT5Summary(t *testing.T) {
	a := &domain.MTAccount{}
	s := &mt5grpc.AccountSummary{
		Balance:    5000.0,
		Credit:     0,
		Profit:     -200.0,
		Equity:     4800.0,
		Margin:     1000.0,
		FreeMargin: 3800.0,
		MarginLevel: 480.0,
		Leverage:   50,
		Currency:   "EUR",
		IsInvestor: true,
		Type:       "real",
	}
	applyMT5Summary(a, s)

	if a.Balance != 5000.0 {
		t.Errorf("Balance = %f, want 5000.0", a.Balance)
	}
	if a.IsInvestor != true {
		t.Errorf("IsInvestor = %v, want true", a.IsInvestor)
	}
	if a.AccountType != "real" {
		t.Errorf("AccountType = %q, want real", a.AccountType)
	}
	if a.ProfitPercent != -4.0 {
		t.Errorf("ProfitPercent = %f, want -4.0", a.ProfitPercent)
	}
}
