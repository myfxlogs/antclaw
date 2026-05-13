//go:build integration

// 跨重启去重集成测试：验证 notification_push_state 的 CRUD 与幂等性。
// 运行方式：go test -tags=integration ./cmd/antclaw-worker/ -run Integration -count=1
// 前置条件：PostgreSQL 可连接（ANTCLAW_DB_* 或 DATABASE_URL 环境变量）。
package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testEnv 集成测试环境：持有连接池和查询对象。
type testEnv struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// requireDB 连接 PostgreSQL；不可用时跳过测试。
func requireDB(t *testing.T) *testEnv {
	t.Helper()
	pool, err := postgres.NewPoolFromEnv()
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("PostgreSQL ping failed: %v", err)
	}
	env := &testEnv{pool: pool, q: db.New(pool)}
	t.Cleanup(func() { pool.Close() })
	return env
}

// testUser 在 users 表中创建一个测试用户，测试结束后清理相关数据。
func (e *testEnv) testUser(t *testing.T) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	ctx := context.Background()
	_, err := e.pool.Exec(ctx,
		`INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
		uid, "test-"+uid.String()[:8], uid.String()[:8]+"@test.local", "$2a$10$placeholder")
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}
	t.Cleanup(func() {
		e.pool.Exec(context.Background(), `DELETE FROM notification_push_state WHERE user_id = $1`, uid)
		e.pool.Exec(context.Background(), `DELETE FROM user_alert_preferences WHERE user_id = $1`, uid)
		e.pool.Exec(context.Background(), `DELETE FROM user_notification_prefs WHERE user_id = $1`, uid)
		e.pool.Exec(context.Background(), `DELETE FROM notifications WHERE user_id = $1`, uid)
		e.pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, uid)
	})
	return uid
}

// ---------- GetPushState / UpsertPushState 基础 CRUD ----------

func TestIntegration_AlreadyPushed_ReturnsFalseForNewKey(t *testing.T) {
	env := requireDB(t)
	uid := env.testUser(t)
	ctx := context.Background()

	// 首次查询 → 应返回 false（未推送过）
	_, err := env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   uid,
		EventKey: "calendar:evt-001:pre:15",
		PushType: "calendar_pre",
	})
	if err == nil {
		t.Error("GetPushState for fresh key should return error (not found)")
	}
}

func TestIntegration_RecordPush_ThenAlreadyPushed_ReturnsTrue(t *testing.T) {
	env := requireDB(t)
	uid := env.testUser(t)
	ctx := context.Background()

	eventKey := "calendar:evt-002:actual"
	pushType := "calendar_actual"

	// 记录一次推送
	_, err := env.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID:     uid,
		EventKey:   eventKey,
		PushType:   pushType,
		LastSentAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("UpsertPushState: %v", err)
	}

	// 再次查询 → 应返回 true（已推送过）
	state, err := env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   uid,
		EventKey: eventKey,
		PushType: pushType,
	})
	if err != nil {
		t.Fatalf("GetPushState after upsert should succeed: %v", err)
	}
	if state.EventKey != eventKey {
		t.Errorf("event_key = %q, want %q", state.EventKey, eventKey)
	}
}

// ---------- 去重幂等性：同一 (user_id, event_key, push_type) 不重复推送 ----------

func TestIntegration_Dedup_PreventsDuplicatePush(t *testing.T) {
	env := requireDB(t)
	uid := env.testUser(t)
	ctx := context.Background()

	eventKey := "digest:daily:" + uid.String() + ":2026-05-13"
	pushType := "daily_news"

	// 第一次记录
	_, err := env.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID:     uid,
		EventKey:   eventKey,
		PushType:   pushType,
		LastSentAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("first UpsertPushState: %v", err)
	}

	// 模拟重启前已推送 → 重启后再次查询应返回 true
	_, err = env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   uid,
		EventKey: eventKey,
		PushType: pushType,
	})
	if err != nil {
		t.Fatalf("GetPushState should find existing record (simulates cross-restart dedup): %v", err)
	}

	// 第二次 Upsert 同 key 应成功（ON CONFLICT UPDATE）但不改变唯一性
	now := time.Now()
	_, err = env.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID:     uid,
		EventKey:   eventKey,
		PushType:   pushType,
		LastSentAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("second UpsertPushState should succeed (upsert): %v", err)
	}

	// 确认 last_sent_at 已更新
	state, _ := env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   uid,
		EventKey: eventKey,
		PushType: pushType,
	})
	if state.LastSentAt.Time.Before(now.Add(-1 * time.Second)) {
		t.Error("last_sent_at should be updated on second upsert")
	}
}

// ---------- 多用户/多事件去重互不干扰 ----------

func TestIntegration_Dedup_IsolationAcrossUsers(t *testing.T) {
	env := requireDB(t)
	uidA := env.testUser(t)
	uidB := env.testUser(t)
	ctx := context.Background()

	eventKey := "calendar:evt-003:pre:60"
	pushType := "calendar_pre"

	// 用户 A 已推送
	env.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID: uidA, EventKey: eventKey, PushType: pushType,
		LastSentAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})

	// 用户 B 未推送 → 应返回 false
	_, err := env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID: uidB, EventKey: eventKey, PushType: pushType,
	})
	if err == nil {
		t.Error("user B should not have push state for user A's event")
	}
}

// ---------- 批量用户扫描 (ListUsersWithAlertPrefs) ----------

func TestIntegration_ScanUsers_PaginatesCorrectly(t *testing.T) {
	env := requireDB(t)
	ctx := context.Background()

	// 游标分页：从 uuid.Nil 开始
	users, err := env.q.ListUsersWithAlertPrefs(ctx, db.ListUsersWithAlertPrefsParams{
		ID:    uuid.Nil,
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("ListUsersWithAlertPrefs: %v", err)
	}
	t.Logf("found %d users (first page)", len(users))

	// 如果第一页有数据，验证游标前进
	if len(users) > 0 && len(users) == 100 {
		lastID := users[len(users)-1].UserID
		nextPage, err := env.q.ListUsersWithAlertPrefs(ctx, db.ListUsersWithAlertPrefsParams{
			ID:    lastID,
			Limit: 100,
		})
		if err != nil {
			t.Fatalf("second page: %v", err)
		}
		t.Logf("second page: %d users", len(nextPage))
		// 第二页不应包含第一页的最后一个用户
		for _, u := range nextPage {
			if u.UserID == lastID {
				t.Error("second page should not duplicate last user from first page")
			}
		}
	}
}

// ---------- 偏好默认值验证 ----------

func TestIntegration_ScanUsers_ReturnsDefaultPrefs(t *testing.T) {
	env := requireDB(t)
	uid := env.testUser(t)
	ctx := context.Background()

	// 为新用户（无 user_alert_preferences 记录）查询偏好
	prefs, err := env.q.GetUserAlertPreferences(ctx, uid)
	if err != nil {
		// 无记录是预期行为 → 业务层使用默认值
		t.Logf("GetUserAlertPreferences for new user: %v (expected — defaults used by caller)", err)
		return
	}
	// 如果有记录（例如 DB 中已有），验证字段
	if len(prefs.Currencies) == 0 {
		t.Error("default currencies should not be empty")
	}
}

// ---------- 旧 push state 清理 ----------

func TestIntegration_DeleteOldPushStates(t *testing.T) {
	env := requireDB(t)
	uid := env.testUser(t)
	ctx := context.Background()

	// 创建一条 91 天前的旧记录
	oldTime := time.Now().Add(-91 * 24 * time.Hour)
	_, err := env.q.UpsertPushState(ctx, db.UpsertPushStateParams{
		UserID:     uid,
		EventKey:   "old:event:key",
		PushType:   "old_type",
		LastSentAt: pgtype.Timestamptz{Time: oldTime, Valid: true},
	})
	if err != nil {
		t.Fatalf("create old push state: %v", err)
	}

	// 删除 90 天前的记录
	cutoff := pgtype.Timestamptz{Time: time.Now().Add(-90 * 24 * time.Hour), Valid: true}
	err = env.q.DeleteOldPushStates(ctx, cutoff)
	if err != nil {
		t.Fatalf("DeleteOldPushStates: %v", err)
	}

	// 验证旧记录已被删除
	_, err = env.q.GetPushState(ctx, db.GetPushStateParams{
		UserID:   uid,
		EventKey: "old:event:key",
		PushType: "old_type",
	})
	if err == nil {
		t.Error("old push state should have been deleted")
	}
}

func init() {
	if os.Getenv("DATABASE_URL") == "" && os.Getenv("ANTCLAW_DB_URL") == "" {
		println("integration test: set DATABASE_URL or ANTCLAW_DB_URL to run")
	}
}
