// Package postgres provides MT account repository implementation.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/domain"
)

// MTAccountRepository provides persistence for MT trading accounts.
type MTAccountRepository struct {
	pool *pgxpool.Pool
}

// NewMTAccountRepository creates a new MT account repository.
func NewMTAccountRepository(pool *pgxpool.Pool) *MTAccountRepository {
	return &MTAccountRepository{pool: pool}
}

// Create inserts a new MT account.
func (r *MTAccountRepository) Create(ctx context.Context, a *domain.MTAccount) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO mt_accounts (
			id, user_id, login, password, mt_type,
			broker_company, broker_server, broker_host,
			status, currency, account_type, alias,
			is_disabled, is_investor, connected_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15
		)`,
		a.ID, a.UserID, a.Login, a.Password, string(a.MTType),
		a.BrokerCompany, a.BrokerServer, a.BrokerHost,
		string(a.Status), a.Currency, a.AccountType, a.Alias,
		a.IsDisabled, a.IsInvestor, a.ConnectedAt,
	)
	if err != nil {
		return fmt.Errorf("create mt account: %w", err)
	}
	return nil
}

// GetByID retrieves an MT account by its ID.
func (r *MTAccountRepository) GetByID(ctx context.Context, id string) (*domain.MTAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, login, password, mt_type,
			broker_company, broker_server, broker_host,
			status, token, currency, account_type, alias,
			last_error, is_disabled, is_investor,
			balance, credit, equity, margin, free_margin, margin_level,
			profit, profit_percent, leverage,
			connected_at, created_at, updated_at
		FROM mt_accounts WHERE id = $1`, id)
	return r.scanAccount(row)
}

// GetByUserIDAndLogin retrieves an account for a specific user and login.
func (r *MTAccountRepository) GetByUserIDAndLogin(ctx context.Context, userID, login, mtType string) (*domain.MTAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, login, password, mt_type,
			broker_company, broker_server, broker_host,
			status, token, currency, account_type, alias,
			last_error, is_disabled, is_investor,
			balance, credit, equity, margin, free_margin, margin_level,
			profit, profit_percent, leverage,
			connected_at, created_at, updated_at
		FROM mt_accounts
		WHERE user_id = $1 AND login = $2 AND mt_type = $3`, userID, login, mtType)
	return r.scanAccount(row)
}

// ListByUserID returns all MT accounts for a given user.
func (r *MTAccountRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.MTAccount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, login, password, mt_type,
			broker_company, broker_server, broker_host,
			status, token, currency, account_type, alias,
			last_error, is_disabled, is_investor,
			balance, credit, equity, margin, free_margin, margin_level,
			profit, profit_percent, leverage,
			connected_at, created_at, updated_at
		FROM mt_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list mt accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.MTAccount
	for rows.Next() {
		a, err := scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// Update saves changes to an existing MT account.
func (r *MTAccountRepository) Update(ctx context.Context, a *domain.MTAccount) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE mt_accounts SET
			status = $2,
			token = $3,
			currency = $4,
			account_type = $5,
			alias = $6,
			last_error = $7,
			is_disabled = $8,
			is_investor = $9,
			balance = $10,
			credit = $11,
			equity = $12,
			margin = $13,
			free_margin = $14,
			margin_level = $15,
			profit = $16,
			profit_percent = $17,
			leverage = $18,
			connected_at = $19,
			updated_at = NOW()
		WHERE id = $1`,
		a.ID, string(a.Status), a.Token,
		a.Currency, a.AccountType, a.Alias,
		a.LastError, a.IsDisabled, a.IsInvestor,
		a.Balance, a.Credit, a.Equity, a.Margin, a.FreeMargin, a.MarginLevel,
		a.Profit, a.ProfitPercent, a.Leverage,
		a.ConnectedAt,
	)
	if err != nil {
		return fmt.Errorf("update mt account: %w", err)
	}
	return nil
}

// UpdateStatus changes the connection status of an account.
func (r *MTAccountRepository) UpdateStatus(ctx context.Context, id string, status domain.MTAccountStatus, lastError string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE mt_accounts SET status = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1`, id, string(status), lastError)
	if err != nil {
		return fmt.Errorf("update mt account status: %w", err)
	}
	return nil
}

// UpdateAccountSummary refreshes the balance/equity/margin fields from MT data.
func (r *MTAccountRepository) UpdateAccountSummary(ctx context.Context, a *domain.MTAccount) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE mt_accounts SET
			balance = $2, credit = $3, equity = $4,
			margin = $5, free_margin = $6, margin_level = $7,
			profit = $8, profit_percent = $9, leverage = $10,
			currency = $11, account_type = $12, is_investor = $13,
			status = $14, connected_at = $15, updated_at = NOW()
		WHERE id = $1`,
		a.ID,
		a.Balance, a.Credit, a.Equity,
		a.Margin, a.FreeMargin, a.MarginLevel,
		a.Profit, a.ProfitPercent, a.Leverage,
		a.Currency, a.AccountType, a.IsInvestor,
		string(a.Status), a.ConnectedAt,
	)
	if err != nil {
		return fmt.Errorf("update mt account summary: %w", err)
	}
	return nil
}

// Delete removes an MT account by ID.
func (r *MTAccountRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM mt_accounts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete mt account: %w", err)
	}
	return nil
}

// scanAccount scans a single row into a domain.MTAccount.
func (r *MTAccountRepository) scanAccount(row pgx.Row) (*domain.MTAccount, error) {
	a := &domain.MTAccount{}
	var mtTypeStr string
	var statusStr string

	err := row.Scan(
		&a.ID, &a.UserID, &a.Login, &a.Password, &mtTypeStr,
		&a.BrokerCompany, &a.BrokerServer, &a.BrokerHost,
		&statusStr, &a.Token, &a.Currency, &a.AccountType, &a.Alias,
		&a.LastError, &a.IsDisabled, &a.IsInvestor,
		&a.Balance, &a.Credit, &a.Equity, &a.Margin, &a.FreeMargin, &a.MarginLevel,
		&a.Profit, &a.ProfitPercent, &a.Leverage,
		&a.ConnectedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("mt account not found: %w", err)
		}
		return nil, fmt.Errorf("scan mt account: %w", err)
	}
	a.MTType = domain.MTType(mtTypeStr)
	a.Status = domain.MTAccountStatus(statusStr)
	return a, nil
}

// scanAccountRow scans from pgx.Rows (for list queries).
func scanAccountRow(row pgx.Rows) (*domain.MTAccount, error) {
	a := &domain.MTAccount{}
	var mtTypeStr string
	var statusStr string

	err := row.Scan(
		&a.ID, &a.UserID, &a.Login, &a.Password, &mtTypeStr,
		&a.BrokerCompany, &a.BrokerServer, &a.BrokerHost,
		&statusStr, &a.Token, &a.Currency, &a.AccountType, &a.Alias,
		&a.LastError, &a.IsDisabled, &a.IsInvestor,
		&a.Balance, &a.Credit, &a.Equity, &a.Margin, &a.FreeMargin, &a.MarginLevel,
		&a.Profit, &a.ProfitPercent, &a.Leverage,
		&a.ConnectedAt, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan mt account row: %w", err)
	}
	a.MTType = domain.MTType(mtTypeStr)
	a.Status = domain.MTAccountStatus(statusStr)
	return a, nil
}
