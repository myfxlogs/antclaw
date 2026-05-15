// Package domain contains MT account domain models.
package domain

import "time"

// MTAccountStatus represents the connection status of an MT account.
type MTAccountStatus string

const (
	MTAccountStatusConnecting   MTAccountStatus = "connecting"
	MTAccountStatusConnected    MTAccountStatus = "connected"
	MTAccountStatusDisconnected MTAccountStatus = "disconnected"
	MTAccountStatusError        MTAccountStatus = "error"
	MTAccountStatusDisabled     MTAccountStatus = "disabled"
)

// MTType represents the MetaTrader platform type.
type MTType string

const (
	MTType4 MTType = "MT4"
	MTType5 MTType = "MT5"
)

// MTAccount is the domain model for a MetaTrader trading account.
// Passwords are stored in plaintext because they must be forwarded to
// the MT gRPC gateway in raw form during connection.
type MTAccount struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Login          string          `json:"login"`
	Password       string          `json:"password"`
	MTType         MTType          `json:"mt_type"`
	BrokerCompany  string          `json:"broker_company"`
	BrokerServer   string          `json:"broker_server"`
	BrokerHost     string          `json:"broker_host"`
	Status         MTAccountStatus `json:"status"`
	Token          string          `json:"token,omitempty"`
	Currency       string          `json:"currency"`
	AccountType    string          `json:"account_type"`
	Alias          string          `json:"alias"`
	LastError      string          `json:"last_error,omitempty"`
	IsDisabled     bool            `json:"is_disabled"`
	IsInvestor     bool            `json:"is_investor"`
	Balance        float64         `json:"balance"`
	Credit         float64         `json:"credit"`
	Equity         float64         `json:"equity"`
	Margin         float64         `json:"margin"`
	FreeMargin     float64         `json:"free_margin"`
	MarginLevel    float64         `json:"margin_level"`
	Profit         float64         `json:"profit"`
	ProfitPercent  float64         `json:"profit_percent"`
	Leverage       int32           `json:"leverage"`
	ConnectedAt    *time.Time      `json:"connected_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// MTBrokerSearchResult represents a broker search result.
type MTBrokerSearchResult struct {
	CompanyName string   `json:"company_name"`
	Servers     []string `json:"servers"`
}
