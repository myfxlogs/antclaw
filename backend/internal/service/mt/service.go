// Package mt provides MetaTrader account service.
// Implements the business logic for MT4/MT5 account binding,
// connection management, and data queries.
package mt

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/antclaw/antclaw/internal/domain"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// Service implements the MT account business logic.
// Dependencies follow the Handler → Service → Repository chain.
type Service struct {
	repo      *postgres.MTAccountRepository
	connMgr   *ConnectionManager
}

// NewService creates a new MT Service.
func NewService(repo *postgres.MTAccountRepository, connMgr *ConnectionManager) *Service {
	return &Service{
		repo:    repo,
		connMgr: connMgr,
	}
}

// CreateAccount binds a new MT trading account for the given user.
// Flow: validate → persist → test connection → rollback on failure.
func (s *Service) CreateAccount(ctx context.Context, userID string, req CreateAccountRequest) (*domain.MTAccount, error) {
	// Validate required fields
	if req.Login == "" || req.Password == "" {
		return nil, fmt.Errorf("login and password are required")
	}
	if req.MTType != domain.MTType4 && req.MTType != domain.MTType5 {
		return nil, fmt.Errorf("invalid mt_type: %s (must be MT4 or MT5)", req.MTType)
	}

	// Check for duplicate (same user, login, mt_type)
	existing, _ := s.repo.GetByUserIDAndLogin(ctx, userID, req.Login, string(req.MTType))
	if existing != nil {
		return nil, fmt.Errorf("account already bound: login=%s type=%s", req.Login, req.MTType)
	}

	// Build domain model
	now := time.Now()
	account := &domain.MTAccount{
		ID:            uuid.New().String(),
		UserID:        userID,
		Login:         req.Login,
		Password:      req.Password,
		MTType:        req.MTType,
		BrokerCompany: req.BrokerCompany,
		BrokerServer:  req.BrokerServer,
		BrokerHost:    req.BrokerHost,
		Status:        domain.MTAccountStatusConnecting,
		Alias:         req.Alias,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Persist
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("create mt account: %w", err)
	}

	// Test connection to MT gateway
	token, err := s.connMgr.Connect(ctx, account)
	if err != nil {
		// Rollback: remove account on connection failure
		s.repo.Delete(ctx, account.ID)
		s.repo.UpdateStatus(ctx, account.ID, domain.MTAccountStatusError, err.Error())
		return nil, fmt.Errorf("connect to MT gateway: %w", err)
	}

	// Fetch account summary to populate fields
	account, err = s.connMgr.GetAccountSummary(ctx, account)
	if err != nil {
		// Still keep the account, but mark with error
		s.repo.UpdateStatus(ctx, account.ID, domain.MTAccountStatusError, err.Error())
		s.connMgr.Disconnect(ctx, account.ID)
		return nil, fmt.Errorf("fetch account summary: %w", err)
	}

	// Persist the updated summary
	account.Token = token
	if err := s.repo.UpdateAccountSummary(ctx, account); err != nil {
		return nil, fmt.Errorf("update account summary: %w", err)
	}

	return account, nil
}

// DeleteAccount removes an MT account binding.
func (s *Service) DeleteAccount(ctx context.Context, userID, accountID string) error {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.UserID != userID {
		return fmt.Errorf("account %s does not belong to user %s", accountID, userID)
	}

	// Disconnect if connected
	if s.connMgr.IsConnected(accountID) {
		s.connMgr.Disconnect(ctx, accountID)
	}

	return s.repo.Delete(ctx, accountID)
}

// GetAccount returns a single MT account.
func (s *Service) GetAccount(ctx context.Context, userID, accountID string) (*domain.MTAccount, error) {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.UserID != userID {
		return nil, fmt.Errorf("account %s does not belong to user %s", accountID, userID)
	}
	return account, nil
}

// ListAccounts returns all MT accounts for a user.
func (s *Service) ListAccounts(ctx context.Context, userID string) ([]*domain.MTAccount, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// ConnectAccount establishes an active connection to the MT gateway.
func (s *Service) ConnectAccount(ctx context.Context, userID, accountID string) (*domain.MTAccount, error) {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.UserID != userID {
		return nil, fmt.Errorf("account %s does not belong to user %s", accountID, userID)
	}

	token, err := s.connMgr.Connect(ctx, account)
	if err != nil {
		s.repo.UpdateStatus(ctx, accountID, domain.MTAccountStatusError, err.Error())
		return nil, fmt.Errorf("connect to MT gateway: %w", err)
	}

	// Fetch fresh summary
	account, err = s.connMgr.GetAccountSummary(ctx, account)
	if err != nil {
		s.repo.UpdateStatus(ctx, accountID, domain.MTAccountStatusError, err.Error())
		return nil, fmt.Errorf("fetch account summary: %w", err)
	}

	account.Token = token
	if err := s.repo.UpdateAccountSummary(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// DisconnectAccount closes the MT gateway connection.
func (s *Service) DisconnectAccount(ctx context.Context, userID, accountID string) error {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account.UserID != userID {
		return fmt.Errorf("account %s does not belong to user %s", accountID, userID)
	}

	if err := s.connMgr.Disconnect(ctx, accountID); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, accountID, domain.MTAccountStatusDisconnected, "")
}

// GetAccountInfo refreshes and returns the current account summary from MT.
func (s *Service) GetAccountInfo(ctx context.Context, userID, accountID string) (*domain.MTAccount, error) {
	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.UserID != userID {
		return nil, fmt.Errorf("account %s does not belong to user %s", accountID, userID)
	}

	// Connect if not connected
	if !s.connMgr.IsConnected(accountID) {
		_, err := s.connMgr.Connect(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("connect to MT gateway: %w", err)
		}
	}

	account, err = s.connMgr.GetAccountSummary(ctx, account)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateAccountSummary(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// SearchBroker searches for MT brokers by company name.
func (s *Service) SearchBroker(ctx context.Context, mtType domain.MTType, company string) ([]domain.MTBrokerSearchResult, error) {
	if mtType != domain.MTType4 && mtType != domain.MTType5 {
		return nil, fmt.Errorf("invalid mt_type: %s (must be MT4 or MT5)", mtType)
	}
	return s.connMgr.SearchBroker(ctx, mtType, company)
}

// CreateAccountRequest is the input for creating an MT account.
type CreateAccountRequest struct {
	Login         string
	Password      string
	MTType        domain.MTType
	BrokerCompany string
	BrokerServer  string
	BrokerHost    string
	Alias         string
}
