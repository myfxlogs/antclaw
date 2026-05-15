package mt

import (
	"fmt"
	"strconv"
	"strings"

	mt4grpc "github.com/antclaw/antclaw/gen/go/third_party/mt4grpc"
	mt5grpc "github.com/antclaw/antclaw/gen/go/third_party/mt5grpc"

	"github.com/antclaw/antclaw/internal/domain"
)

// parseLoginInt32 converts a login string to int32 for MT4.
func parseLoginInt32(login string) (int32, error) {
	v, err := strconv.ParseInt(login, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid login %q: %w", login, err)
	}
	return int32(v), nil
}

// parseLoginUint64 converts a login string to uint64 for MT5.
func parseLoginUint64(login string) (uint64, error) {
	v, err := strconv.ParseUint(login, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid login %q: %w", login, err)
	}
	return v, nil
}

// parseHostPort splits "host:port" into separate values.
// Default port is 443 if not specified.
func parseHostPort(hostPort string) (host string, port int32) {
	host = hostPort
	port = 443

	if idx := strings.LastIndex(hostPort, ":"); idx > 0 {
		host = hostPort[:idx]
		if p, err := strconv.Atoi(hostPort[idx+1:]); err == nil && p > 0 {
			port = int32(p)
		}
	}
	return
}

// applyMT4Summary fills a domain.MTAccount from an mt4grpc AccountSummary.
func applyMT4Summary(a *domain.MTAccount, s *mt4grpc.AccountSummary) {
	a.Balance = s.Balance
	a.Credit = s.Credit
	a.Profit = s.Profit
	a.Equity = s.Equity
	a.Margin = s.Margin
	a.FreeMargin = s.FreeMargin
	a.MarginLevel = s.MarginLevel
	a.Leverage = int32(s.Leverage)
	a.Currency = s.Currency
	a.IsInvestor = s.IsInvestor
	a.ProfitPercent = calcProfitPercent(a.Profit, a.Balance)

	switch s.Type {
	case mt4grpc.AccountType_AccountType_Demo:
		a.AccountType = "demo"
	case mt4grpc.AccountType_AccountType_Contest:
		a.AccountType = "contest"
	default:
		a.AccountType = "real"
	}
}

// applyMT5Summary fills a domain.MTAccount from an mt5grpc AccountSummary.
func applyMT5Summary(a *domain.MTAccount, s *mt5grpc.AccountSummary) {
	a.Balance = s.Balance
	a.Credit = s.Credit
	a.Profit = s.Profit
	a.Equity = s.Equity
	a.Margin = s.Margin
	a.FreeMargin = s.FreeMargin
	a.MarginLevel = s.MarginLevel
	a.Leverage = int32(s.Leverage)
	a.Currency = s.Currency
	a.IsInvestor = s.IsInvestor
	a.ProfitPercent = calcProfitPercent(a.Profit, a.Balance)
	a.AccountType = strings.ToLower(s.Type)
}

// calcProfitPercent computes profit as a percentage of balance.
func calcProfitPercent(profit, balance float64) float64 {
	if balance == 0 {
		return 0
	}
	return (profit / balance) * 100
}

// convertMT4Companies converts mt4grpc.Company results to domain types.
func convertMT4Companies(companies []*mt4grpc.Company) []domain.MTBrokerSearchResult {
	var out []domain.MTBrokerSearchResult
	for _, c := range companies {
		r := domain.MTBrokerSearchResult{
			CompanyName: c.CompanyName,
		}
		for _, srv := range c.Results {
			r.Servers = append(r.Servers, srv.Access...)
		}
		out = append(out, r)
	}
	return out
}

// convertMT5Companies converts mt5grpc.Company results to domain types.
func convertMT5Companies(companies []*mt5grpc.Company) []domain.MTBrokerSearchResult {
	var out []domain.MTBrokerSearchResult
	for _, c := range companies {
		r := domain.MTBrokerSearchResult{
			CompanyName: c.CompanyName,
		}
		for _, srv := range c.Results {
			r.Servers = append(r.Servers, srv.Access...)
		}
		out = append(out, r)
	}
	return out
}
