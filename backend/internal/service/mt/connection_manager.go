// Package mt provides MetaTrader account connection management.
package mt

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/antclaw/antclaw/internal/domain"
	"github.com/antclaw/antclaw/internal/infra/apiclient/mt4"
	"github.com/antclaw/antclaw/internal/infra/apiclient/mt5"
)

// activeConn holds an active MT gateway connection.
type activeConn struct {
	accountID string
	mtType    domain.MTType
	token     string
	connected bool
	lastError error
	mu        sync.RWMutex
}

// ConnectionManager manages MT4/MT5 gateway connections for multiple accounts.
// It maps account IDs to active tokens and provides Connect/Disconnect/IsConnected.
type ConnectionManager struct {
	mt4Client *mt4.Client
	mt5Client *mt5.Client

	conns map[string]*activeConn // accountID -> connection
	mu    sync.RWMutex
}

// NewConnectionManager creates a new ConnectionManager.
// mt4BaseURL and mt5BaseURL are the MT gateway addresses, e.g. "http://mt4-gateway:8080".
func NewConnectionManager(mt4BaseURL, mt5BaseURL string) *ConnectionManager {
	return &ConnectionManager{
		mt4Client: mt4.NewClient(mt4BaseURL),
		mt5Client: mt5.NewClient(mt5BaseURL),
		conns:     make(map[string]*activeConn),
	}
}

// Connect establishes a connection to the MT account and returns the session token.
// It stores the token for subsequent API calls.
func (m *ConnectionManager) Connect(ctx context.Context, account *domain.MTAccount) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing connection if any
	if existing, ok := m.conns[account.ID]; ok {
		m.disconnectLocked(ctx, existing)
	}

	var token string
	var err error

	switch account.MTType {
	case domain.MTType4:
		login, parseErr := parseLoginInt32(account.Login)
		if parseErr != nil {
			return "", fmt.Errorf("mt4 invalid login %q: %w", account.Login, parseErr)
		}
		host, port := parseHostPort(account.BrokerHost)
		token, err = m.mt4Client.Connect(ctx, login, account.Password, host, port)
		if err != nil {
			return "", fmt.Errorf("mt4 connect: %w", err)
		}
	case domain.MTType5:
		login, parseErr := parseLoginUint64(account.Login)
		if parseErr != nil {
			return "", fmt.Errorf("mt5 invalid login %q: %w", account.Login, parseErr)
		}
		host, port := parseHostPort(account.BrokerHost)
		token, err = m.mt5Client.Connect(ctx, login, account.Password, host, port)
		if err != nil {
			// Fallback to ConnectEx using server name
			token, err = m.mt5Client.ConnectEx(ctx, login, account.Password, account.BrokerServer)
			if err != nil {
				return "", fmt.Errorf("mt5 connect: %w", err)
			}
		}
	default:
		return "", fmt.Errorf("unsupported MT type: %s", account.MTType)
	}

	conn := &activeConn{
		accountID: account.ID,
		mtType:    account.MTType,
		token:     token,
		connected: true,
	}
	m.conns[account.ID] = conn

	return token, nil
}

// Disconnect closes the connection for the given account.
func (m *ConnectionManager) Disconnect(ctx context.Context, accountID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.conns[accountID]
	if !ok {
		return nil // already disconnected
	}
	return m.disconnectLocked(ctx, conn)
}

// disconnectLocked closes a connection; caller must hold m.mu.
func (m *ConnectionManager) disconnectLocked(ctx context.Context, conn *activeConn) error {
	var err error
	switch conn.mtType {
	case domain.MTType4:
		err = m.mt4Client.Disconnect(ctx, conn.token)
	case domain.MTType5:
		err = m.mt5Client.Disconnect(ctx, conn.token)
	}
	conn.connected = false
	conn.lastError = err
	delete(m.conns, conn.accountID)
	if err != nil {
		return fmt.Errorf("disconnect %s: %w", conn.accountID, err)
	}
	return nil
}

// IsConnected returns whether the account has an active connection.
func (m *ConnectionManager) IsConnected(accountID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[accountID]
	return ok && conn.connected
}

// GetToken returns the session token for an active connection.
func (m *ConnectionManager) GetToken(accountID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[accountID]
	if !ok || !conn.connected {
		return "", false
	}
	return conn.token, true
}

// GetAccountSummary fetches the account summary from the MT gateway.
func (m *ConnectionManager) GetAccountSummary(ctx context.Context, account *domain.MTAccount) (*domain.MTAccount, error) {
	token, ok := m.GetToken(account.ID)
	if !ok {
		return nil, fmt.Errorf("account %s not connected", account.ID)
	}

	switch account.MTType {
	case domain.MTType4:
		summary, err := m.mt4Client.AccountSummary(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("mt4 account summary: %w", err)
		}
		applyMT4Summary(account, summary)
	case domain.MTType5:
		summary, err := m.mt5Client.AccountSummary(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("mt5 account summary: %w", err)
		}
		applyMT5Summary(account, summary)
	default:
		return nil, fmt.Errorf("unsupported MT type: %s", account.MTType)
	}

	account.Status = domain.MTAccountStatusConnected
	now := time.Now()
	account.ConnectedAt = &now
	return account, nil
}

// SearchBroker searches for brokers using the MT gateway.
func (m *ConnectionManager) SearchBroker(ctx context.Context, mtType domain.MTType, company string) ([]domain.MTBrokerSearchResult, error) {
	switch mtType {
	case domain.MTType4:
		companies, err := m.mt4Client.Search(ctx, company)
		if err != nil {
			return nil, fmt.Errorf("mt4 broker search: %w", err)
		}
		return convertMT4Companies(companies), nil
	case domain.MTType5:
		companies, err := m.mt5Client.Search(ctx, company)
		if err != nil {
			return nil, fmt.Errorf("mt5 broker search: %w", err)
		}
		return convertMT5Companies(companies), nil
	default:
		return nil, fmt.Errorf("unsupported MT type: %s", mtType)
	}
}

// CheckHealth verifies all active connections are alive.
// Returns a map of accountID -> error for unhealthy connections.
func (m *ConnectionManager) CheckHealth(ctx context.Context) map[string]error {
	m.mu.RLock()
	conns := make([]*activeConn, 0, len(m.conns))
	for _, c := range m.conns {
		conns = append(conns, c)
	}
	m.mu.RUnlock()

	unhealthy := make(map[string]error)
	for _, conn := range conns {
		conn.mu.RLock()
		token := conn.token
		mtType := conn.mtType
		conn.mu.RUnlock()

		var err error
		switch mtType {
		case domain.MTType4:
			_, err = m.mt4Client.CheckConnect(ctx, token)
		case domain.MTType5:
			_, err = m.mt5Client.CheckConnect(ctx, token)
		}
		if err != nil {
			conn.mu.Lock()
			conn.connected = false
			conn.lastError = err
			conn.mu.Unlock()
			unhealthy[conn.accountID] = err
			log.Printf("[mt] connection unhealthy %s: %v", conn.accountID, err)
		}
	}
	return unhealthy
}

// StartHealthCheck starts a background goroutine that periodically checks connections.
func (m *ConnectionManager) StartHealthCheck(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.CheckHealth(ctx)
			}
		}
	}()
}

// Shutdown disconnects all active connections.
func (m *ConnectionManager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, conn := range m.conns {
		if err := m.disconnectLocked(ctx, conn); err != nil {
			log.Printf("[mt] shutdown disconnect %s: %v", id, err)
		}
	}
}
