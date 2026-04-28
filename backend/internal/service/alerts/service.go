package alerts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	alertv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxActiveAlertsPerUser = 50

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func validateAlertType(t string) bool {
	switch t {
	case "cot_extreme", "signal_flip", "regime_change", "price_threshold":
		return true
	default:
		return false
	}
}

func validateParamsJSON(s string) error {
	if s == "" {
		return errors.New("params_json required")
	}
	var tmp map[string]any
	if err := json.Unmarshal([]byte(s), &tmp); err != nil {
		return fmt.Errorf("invalid params_json: %w", err)
	}
	return nil
}

func (s *Service) CreateAlert(ctx context.Context, userID, alertType, symbol, paramsJSON string, cooldownSeconds int32) (*alertv1.AlertRule, error) {
	if !validateAlertType(alertType) {
		return nil, fmt.Errorf("invalid alert_type")
	}
	if cooldownSeconds < 60 || cooldownSeconds > 86400 {
		return nil, fmt.Errorf("invalid cooldown_seconds")
	}
	if err := validateParamsJSON(paramsJSON); err != nil {
		return nil, err
	}
	var activeCount int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_signal_alerts WHERE user_id=$1 AND enabled=true AND deleted_at IS NULL`, userID).Scan(&activeCount); err != nil {
		return nil, err
	}
	if activeCount >= MaxActiveAlertsPerUser {
		return nil, fmt.Errorf("alert quota exceeded")
	}

	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO user_signal_alerts(user_id,alert_type,symbol,params,enabled,cooldown_seconds)
VALUES($1,$2,$3,$4::jsonb,true,$5) RETURNING id`, userID, alertType, symbol, paramsJSON, cooldownSeconds).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &alertv1.AlertRule{
		Id: id, UserId: userID, AlertType: alertType, Symbol: symbol, ParamsJson: paramsJSON, Enabled: true, CooldownSeconds: cooldownSeconds,
	}, nil
}

func (s *Service) ListAlerts(ctx context.Context, userID, alertType string) ([]*alertv1.AlertRule, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,user_id,alert_type,symbol,params::text,enabled,COALESCE(EXTRACT(EPOCH FROM last_fired_at)::bigint,0),cooldown_seconds
FROM user_signal_alerts WHERE user_id=$1 AND deleted_at IS NULL AND ($2='' OR alert_type=$2) ORDER BY id DESC`, userID, alertType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (*alertv1.AlertRule, error) {
		var x alertv1.AlertRule
		if err := r.Scan(&x.Id, &x.UserId, &x.AlertType, &x.Symbol, &x.ParamsJson, &x.Enabled, &x.LastFiredAt, &x.CooldownSeconds); err != nil {
			return nil, err
		}
		return &x, nil
	})
}

func (s *Service) UpdateAlert(ctx context.Context, userID string, id int64, paramsJSON string, cooldownSeconds int32) (*alertv1.AlertRule, error) {
	if cooldownSeconds < 60 || cooldownSeconds > 86400 {
		return nil, fmt.Errorf("invalid cooldown_seconds")
	}
	if err := validateParamsJSON(paramsJSON); err != nil {
		return nil, err
	}
	var rule alertv1.AlertRule
	err := s.pool.QueryRow(ctx, `UPDATE user_signal_alerts SET params=$1::jsonb,cooldown_seconds=$2,updated_at=NOW()
WHERE id=$3 AND user_id=$4 AND deleted_at IS NULL
RETURNING id,user_id,alert_type,symbol,params::text,enabled,COALESCE(EXTRACT(EPOCH FROM last_fired_at)::bigint,0),cooldown_seconds`,
		paramsJSON, cooldownSeconds, id, userID).
		Scan(&rule.Id, &rule.UserId, &rule.AlertType, &rule.Symbol, &rule.ParamsJson, &rule.Enabled, &rule.LastFiredAt, &rule.CooldownSeconds)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *Service) DeleteAlert(ctx context.Context, userID string, id int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE user_signal_alerts SET deleted_at=NOW(),updated_at=NOW() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	return err
}

func (s *Service) ToggleAlert(ctx context.Context, userID string, id int64, enabled bool) (*alertv1.AlertRule, error) {
	var rule alertv1.AlertRule
	err := s.pool.QueryRow(ctx, `UPDATE user_signal_alerts SET enabled=$1,updated_at=NOW() WHERE id=$2 AND user_id=$3 AND deleted_at IS NULL
RETURNING id,user_id,alert_type,symbol,params::text,enabled,COALESCE(EXTRACT(EPOCH FROM last_fired_at)::bigint,0),cooldown_seconds`,
		enabled, id, userID).Scan(&rule.Id, &rule.UserId, &rule.AlertType, &rule.Symbol, &rule.ParamsJson, &rule.Enabled, &rule.LastFiredAt, &rule.CooldownSeconds)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *Service) listActiveByType(ctx context.Context, alertType string) ([]*alertv1.AlertRule, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,user_id,alert_type,symbol,params::text,enabled,COALESCE(EXTRACT(EPOCH FROM last_fired_at)::bigint,0),cooldown_seconds
FROM user_signal_alerts WHERE enabled=true AND deleted_at IS NULL AND alert_type=$1`, alertType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (*alertv1.AlertRule, error) {
		var x alertv1.AlertRule
		if err := r.Scan(&x.Id, &x.UserId, &x.AlertType, &x.Symbol, &x.ParamsJson, &x.Enabled, &x.LastFiredAt, &x.CooldownSeconds); err != nil {
			return nil, err
		}
		return &x, nil
	})
}

func (s *Service) markFired(ctx context.Context, id int64, at time.Time) error {
	_, err := s.pool.Exec(ctx, `UPDATE user_signal_alerts SET last_fired_at=$1,updated_at=NOW() WHERE id=$2`, at, id)
	return err
}

// Legacy API compatibility
func (s *Service) ListSubscriptions(ctx context.Context, alertTypeFilter string, activeOnly bool) (*alertv1.ListSubscriptionsResponse, error) {
	return &alertv1.ListSubscriptionsResponse{Subscriptions: []*alertv1.AlertSubscription{}}, nil
}
func (s *Service) Subscribe(ctx context.Context, alertType, pair, condition, threshold, notificationMethod string) (*alertv1.SubscribeResponse, error) {
	return &alertv1.SubscribeResponse{}, nil
}
func (s *Service) Unsubscribe(ctx context.Context, subscriptionID string) (*alertv1.UnsubscribeResponse, error) {
	return &alertv1.UnsubscribeResponse{}, nil
}
func (s *Service) RegisterWebhook(ctx context.Context, url, secret string, eventTypes []string) (*alertv1.RegisterWebhookResponse, error) {
	return &alertv1.RegisterWebhookResponse{}, nil
}
func (s *Service) ListWebhooks(ctx context.Context) (*alertv1.ListWebhooksResponse, error) {
	return &alertv1.ListWebhooksResponse{}, nil
}
