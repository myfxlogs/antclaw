package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	adminv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/service/audit"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Service implements Admin business logic with real PostgreSQL storage.
type Service struct {
	queries   *db.Queries
	auditSvc  *audit.AuditService
	redis     *redis.Client
}

// NewService creates a new AdminService with database connection.
func NewService(queries *db.Queries, auditSvc *audit.AuditService, redis *redis.Client) *Service {
	return &Service{
		queries:   queries,
		auditSvc:  auditSvc,
		redis:     redis,
	}
}

// ListUsers lists users with pagination from PostgreSQL.
func (s *Service) ListUsers(ctx context.Context, cursor string, pageSize int32, emailFilter, roleFilter string, bannedOnly bool) (*adminv1.ListUsersResponse, error) {
	// Parse cursor as offset
	var offset int32
	if cursor != "" {
		fmt.Sscanf(cursor, "%d", &offset)
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Convert bannedOnly to status filter
	status := ""
	if bannedOnly {
		status = "banned"
	}

	params := db.ListUsersParams{
		Column1: emailFilter,
		Column2: roleFilter,
		Column3: status != "",
		Limit:   pageSize,
		Offset:  offset,
	}

	dbUsers, err := s.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.queries.CountUsers(ctx)
	if err != nil {
		total = int64(len(dbUsers))
	}

	var users []*adminv1.User
	for _, u := range dbUsers {
		users = append(users, dbUserToProto(u))
	}

	nextCursor := ""
	if len(users) >= int(pageSize) {
		nextCursor = fmt.Sprintf("%d", offset+pageSize)
	}

	return &adminv1.ListUsersResponse{
		Users:      users,
		Total:      int32(total),
		NextCursor: nextCursor,
	}, nil
}

// dbUserToProto converts db.User to proto User.
func dbUserToProto(u db.User) *adminv1.User {
	user := &adminv1.User{
		UserId:        u.ID.String(),
		Email:         u.Email,
		Username:      derefString(u.Username),
		DisplayName:   derefString(u.DisplayName),
		Roles:         []string{u.Role},
		EmailVerified: u.EmailVerifiedAt.Valid,
		CreatedAt:     u.CreatedAt.Time.Unix(),
		UpdatedAt:     u.UpdatedAt.Time.Unix(),
	}
	// Status may not exist in proto, skip it
	return user
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// parseUUID parses a string UUID to uuid.UUID.
func parseUUID(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}
	}
	return u
}

// SetRole sets user roles.
func (s *Service) SetRole(ctx context.Context, userID string, roles []string) (*adminv1.SetRoleResponse, error) {
	role := "user"
	if len(roles) > 0 {
		role = roles[0]
	}

	err := s.queries.UpdateUserRole(ctx, db.UpdateUserRoleParams{
		ID:   parseUUID(userID),
		Role: role,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &adminv1.SetRoleResponse{}, nil
}

// Ban bans a user.
func (s *Service) Ban(ctx context.Context, userID, reason string, expiresAt int64) (*adminv1.BanResponse, error) {
	err := s.queries.BanUser(ctx, parseUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to ban user: %w", err)
	}
	return &adminv1.BanResponse{}, nil
}

// Unban unbans a user.
func (s *Service) Unban(ctx context.Context, userID string) (*adminv1.UnbanResponse, error) {
	err := s.queries.UnbanUser(ctx, parseUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to unban user: %w", err)
	}
	return &adminv1.UnbanResponse{}, nil
}

// RunJob 通过 Redis Pub/Sub (channel "jobs:trigger") 通知 worker 立即执行指定 job。
// jobName 必须等于 worker 注册的 jobID（如 "macro-sync"）。
func (s *Service) RunJob(ctx context.Context, jobName string, params map[string]string) (*adminv1.RunJobResponse, error) {
	if jobName == "" {
		return nil, fmt.Errorf("missing job_name")
	}
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}
	payload, _ := json.Marshal(map[string]string{"job_id": jobName})
	if err := s.redis.Publish(ctx, "jobs:trigger", payload).Err(); err != nil {
		return nil, fmt.Errorf("publish trigger: %w", err)
	}
	return &adminv1.RunJobResponse{
		JobId:  jobName,
		Status: "triggered",
	}, nil
}

// ListJobs lists scheduled jobs with real status from Redis.
func (s *Service) ListJobs(ctx context.Context, statusFilter string) (*adminv1.ListJobsResponse, error) {
	jobConfigs := []struct {
		id       string
		name     string
		interval time.Duration
	}{
		{"calendar-sync", "财经日历采集 (每小时)", 1 * time.Hour},
		{"macro-sync", "宏观数据采集 (每4小时)", 4 * time.Hour},
		{"actuals-update", "实际值更新 (每30分钟)", 30 * time.Minute},
		{"cot-sync", "COT持仓采集 (每6小时)", 6 * time.Hour},
		{"price-sync", "价格数据采集 (每6小时)", 6 * time.Hour},
		{"sentiment-sync", "情绪数据采集 (每小时)", 1 * time.Hour},
		{"onchain-sync", "链上数据采集 (每小时)", 1 * time.Hour},
		{"intraday-sync", "分时价格采集 (Yahoo 5min)", 5 * time.Minute},
		{"defi-sync", "DeFi数据采集 (DefiLlama)", 1 * time.Hour},
		{"vix-term-sync", "VIX期限结构采集", 1 * time.Hour},
		{"dvol-sync", "DVOL采集 (Deribit)", 1 * time.Hour},
		{"cot-analysis", "COT分析 (Index/Z-score/百分位)", 6 * time.Hour},
		{"macro-regime", "宏观状态分类 (Risk-On/Off)", 4 * time.Hour},
		{"flow-divergence", "资金流向背离分析", 1 * time.Hour},
		{"volume-profile", "成交量分布分析 (POC/VAH/VAL)", 1 * time.Hour},
	}

	var jobs []*adminv1.Job
	now := time.Now()

	for _, cfg := range jobConfigs {
		if statusFilter != "" {
			// Filter will be applied below after we get real status
		}

		status, lastRun, nextRun, lastErr := s.getJobStatus(ctx, cfg.id, cfg.interval, now)

		if statusFilter != "" && status != statusFilter {
			continue
		}

		jobs = append(jobs, &adminv1.Job{
			JobId:     cfg.id,
			JobName:   cfg.name,
			Status:    status,
			LastRun:   lastRun.Format(time.RFC3339),
			NextRun:   nextRun.Format(time.RFC3339),
			Enabled:   s.isJobEnabled(ctx, cfg.id),
			LastError: lastErr,
		})
	}
	return &adminv1.ListJobsResponse{Jobs: jobs}, nil
}

// isJobEnabled 查询 Redis 中的 job 启用状态。默认启用（无 key 时）。
func (s *Service) isJobEnabled(ctx context.Context, jobID string) bool {
	if s.redis == nil {
		return true
	}
	val, err := s.redis.Get(ctx, fmt.Sprintf("jobs:enabled:%s", jobID)).Result()
	if err != nil {
		return true
	}
	return val != "false"
}

// SetJobEnabled 设置 job 启用/禁用状态，持久化到 Redis。
func (s *Service) SetJobEnabled(ctx context.Context, jobID string, enabled bool) (*adminv1.SetJobEnabledResponse, error) {
	if jobID == "" {
		return nil, fmt.Errorf("missing job_id")
	}
	if s.redis == nil {
		return nil, fmt.Errorf("redis not available")
	}
	val := "true"
	if !enabled {
		val = "false"
	}
	if err := s.redis.Set(ctx, fmt.Sprintf("jobs:enabled:%s", jobID), val, 0).Err(); err != nil {
		return nil, fmt.Errorf("set job enabled: %w", err)
	}
	return &adminv1.SetJobEnabledResponse{JobId: jobID, Enabled: enabled}, nil
}

// getJobStatus retrieves the last known status of a job from Redis.
func (s *Service) getJobStatus(ctx context.Context, jobID string, interval time.Duration, now time.Time) (status string, lastRun, nextRun time.Time, lastErr string) {
	defaultStatus := "unknown"
	lastRun = now.Add(-interval)
	nextRun = now

	if s.redis != nil {
		key := fmt.Sprintf("jobs:status:%s", jobID)
		data, err := s.redis.Get(ctx, key).Result()
		if err == nil && data != "" {
			var evt struct {
				Status     string `json:"status"`
				StartedAt  int64  `json:"started_at"`
				FinishedAt int64  `json:"finished_at"`
				Error      string `json:"error"`
			}
			if err := json.Unmarshal([]byte(data), &evt); err == nil {
				defaultStatus = evt.Status
				if evt.FinishedAt > 0 {
					lastRun = time.Unix(evt.FinishedAt, 0)
				} else if evt.StartedAt > 0 {
					lastRun = time.Unix(evt.StartedAt, 0)
				}
				if evt.Status == "running" {
					nextRun = now
				} else {
					nextRun = lastRun.Add(interval)
				}
				return defaultStatus, lastRun, nextRun, evt.Error
			}
		}
	}

	// Fallback: estimate based on interval
	schedule := now.Truncate(interval).Add(interval)
	if schedule.Before(now) {
		schedule = schedule.Add(interval)
	}
	return defaultStatus, lastRun, schedule, ""
}

// ListAuditLogs lists audit logs from PostgreSQL.
func (s *Service) ListAuditLogs(ctx context.Context, cursor string, pageSize int32, userIDFilter, actionFilter string) (*adminv1.ListAuditLogsResponse, error) {
	if pageSize <= 0 || pageSize > 1000 {
		pageSize = 100
	}

	dbLogs, err := s.queries.ListAuditLogs(ctx, db.ListAuditLogsParams{
		Limit:  pageSize,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	var logs []*adminv1.AuditLogEntry
	for _, l := range dbLogs {
		userID := ""
		if l.UserID.Valid {
			userID = uuid.UUID(l.UserID.Bytes).String()
		}
		logs = append(logs, &adminv1.AuditLogEntry{
			LogId:     fmt.Sprintf("%d", l.ID),
			UserId:    userID,
			Action:    l.Action,
			Resource:  l.Resource,
			Details:   l.Details,
			IpAddress: derefString(l.IpAddress),
			CreatedAt: l.CreatedAt.Time.Unix(),
		})
	}

	return &adminv1.ListAuditLogsResponse{
		Entries: logs,
	}, nil
}

// ListWebhookDeliveries lists webhook deliveries.
func (s *Service) ListWebhookDeliveries(ctx context.Context, cursor string, pageSize int32, webhookIDFilter string) (*adminv1.ListWebhookDeliveriesResponse, error) {
	// Webhook deliveries not yet implemented in database - return empty
	return &adminv1.ListWebhookDeliveriesResponse{
		Deliveries: []*adminv1.WebhookDelivery{},
	}, nil
}

// ForceLogout forces logout of a user from all sessions.
func (s *Service) ForceLogout(ctx context.Context, userID string) (*adminv1.ForceLogoutResponse, error) {
	// Revoke all user sessions
	err := s.queries.RevokeAllUserSessions(ctx, parseUUID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to revoke sessions: %w", err)
	}
	return &adminv1.ForceLogoutResponse{}, nil
}

// AdminResetUserPassword 管理员直接重置用户密码（Argon2id 落库）。
func (s *Service) AdminResetUserPassword(ctx context.Context, userID, newPassword string) (*adminv1.AdminResetUserPasswordResponse, error) {
	if len(newPassword) < 8 {
		return nil, fmt.Errorf("%w: need at least 8 characters", ErrPasswordPolicy)
	}
	id := parseUUID(userID)
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid user_id")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	if err := s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: hash,
	}); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}
	if s.auditSvc != nil {
		_, _ = s.auditSvc.Log(ctx, audit.AuditEntry{
			Action:   "admin_reset_password",
			Resource: "user:" + userID,
			Details:  "password reset by admin",
		})
	}
	return &adminv1.AdminResetUserPasswordResponse{}, nil
}
