// Package postgres provides the Trader Repository for profile / follow operations.
package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TraderProfileRow is the flat DB row for a trader profile.
type TraderProfileRow struct {
	UserID         string
	DisplayName    string
	Bio            string
	Tier           string
	ShowWinRate    bool
	ShowProfitFact bool
	ShowSharpe     bool
	ShowTotalTrad  bool
	WinRate        float64
	ProfitFactor   float64
	SharpeRatio    float64
	TotalTrades    int32
	FollowerCount  int32
	FollowingCount int32
	CreatedAt      time.Time
	IsFollowing    bool
}

// UserInfoRow carries a user info for follower/following lists.
type UserInfoRow struct {
	UserID        string
	DisplayName   string
	Tier          string
	FollowerCount int32
}

// TraderRepository defines data operations for trader profiles and follows.
type TraderRepository interface {
	GetProfile(ctx context.Context, userID string) (*TraderProfileRow, error)
	UpdateProfile(ctx context.Context, userID string, row *TraderProfileRow) error
	CheckUserExists(ctx context.Context, userID string) (bool, error)
	GetUserName(ctx context.Context, userID string) (string, error)

	Follow(ctx context.Context, followerID, followingID string) error
	Unfollow(ctx context.Context, followerID, followingID string) error
	CheckFollowExists(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowerCount(ctx context.Context, userID string) (int32, error)
	GetFollowingCount(ctx context.Context, userID string) (int32, error)

	GetFollowers(ctx context.Context, userID string, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error)
	GetFollowing(ctx context.Context, userID string, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error)
	ListRecommendedTraders(ctx context.Context, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error)

	IsFollowing(ctx context.Context, currentUserID, targetUserID string) (bool, error)
}

type traderRepo struct{ pool *pgxpool.Pool }

func NewTraderRepository(pool *pgxpool.Pool) TraderRepository { return &traderRepo{pool: pool} }

// ----- shared helpers -----

const userInfoSelect = `u.id, COALESCE(NULLIF(u.display_name,''), u.username, u.code_id, 'User-' || LEFT(u.id::text, 8)), 'normal',
	(SELECT COUNT(*) FROM alfq_follows WHERE following_id = u.id)::int4`

func scanUserInfo(scanner interface{ Scan(dest ...interface{}) error }) (*UserInfoRow, error) {
	u := &UserInfoRow{}
	err := scanner.Scan(&u.UserID, &u.DisplayName, &u.Tier, &u.FollowerCount)
	return u, err
}

// executePaginatedUserList runs a follow-list query and returns users + next cursor.
func (r *traderRepo) executePaginatedUserList(ctx context.Context, query string, args []interface{}, limit int32) ([]*UserInfoRow, *SocialCursor, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var users []*UserInfoRow
	for rows.Next() {
		u, err := scanUserInfo(rows)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	hasMore := len(users) > int(limit)
	if hasMore {
		users = users[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(users) > 0 {
		last := users[len(users)-1]
		nextCursor = &SocialCursor{CreatedAt: time.Time{}, ID: last.UserID}
	}
	return users, nextCursor, nil
}

// ----- Profile -----

const profileSelect = `id, ` + PublicDisplayNameExpr + `, COALESCE(bio,''), 'normal',
	COALESCE(show_win_rate, false), COALESCE(show_profit_factor, false), COALESCE(show_sharpe, false), COALESCE(show_total_trades, false),
	0.0, 0.0, 0.0, 0,
	(SELECT COUNT(*) FROM alfq_follows WHERE following_id = $1)::int4,
	(SELECT COUNT(*) FROM alfq_follows WHERE follower_id = $1)::int4,
	NOW()`

func (r *traderRepo) GetProfile(ctx context.Context, userID string) (*TraderProfileRow, error) {
	row := &TraderProfileRow{}
	err := r.pool.QueryRow(ctx,
		`SELECT `+profileSelect+` FROM users WHERE id = $1`, userID).Scan(
		&row.UserID, &row.DisplayName, &row.Bio, &row.Tier,
		&row.ShowWinRate, &row.ShowProfitFact, &row.ShowSharpe, &row.ShowTotalTrad,
		&row.WinRate, &row.ProfitFactor, &row.SharpeRatio, &row.TotalTrades,
		&row.FollowerCount, &row.FollowingCount,
		&row.CreatedAt,
	)
	return row, err
}

func (r *traderRepo) UpdateProfile(ctx context.Context, userID string, row *TraderProfileRow) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET
			display_name = COALESCE(NULLIF($1,''), display_name),
			bio = COALESCE(NULLIF($2,''), bio),
			show_win_rate = $3,
			show_profit_factor = $4,
			show_sharpe = $5,
			show_total_trades = $6
		WHERE id = $7`,
		row.DisplayName, row.Bio, row.ShowWinRate, row.ShowProfitFact, row.ShowSharpe, row.ShowTotalTrad, userID)
	return err
}

func (r *traderRepo) CheckUserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists)
	return exists, err
}

func (r *traderRepo) GetUserName(ctx context.Context, userID string) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx,
		`SELECT `+PublicDisplayNameExpr+` FROM users WHERE id=$1`, userID).Scan(&name)
	return name, err
}

// ----- Follow -----

func (r *traderRepo) Follow(ctx context.Context, followerID, followingID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO alfq_follows (follower_id, following_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		followerID, followingID)
	return err
}

func (r *traderRepo) Unfollow(ctx context.Context, followerID, followingID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM alfq_follows WHERE follower_id=$1 AND following_id=$2`,
		followerID, followingID)
	return err
}

func (r *traderRepo) CheckFollowExists(ctx context.Context, followerID, followingID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM alfq_follows WHERE follower_id=$1 AND following_id=$2)`,
		followerID, followingID).Scan(&exists)
	return exists, err
}

func (r *traderRepo) GetFollowerCount(ctx context.Context, userID string) (int32, error) {
	var cnt int32
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)::int4 FROM alfq_follows WHERE following_id=$1`, userID).Scan(&cnt)
	return cnt, err
}

func (r *traderRepo) GetFollowingCount(ctx context.Context, userID string) (int32, error) {
	var cnt int32
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*)::int4 FROM alfq_follows WHERE follower_id=$1`, userID).Scan(&cnt)
	return cnt, err
}

func (r *traderRepo) IsFollowing(ctx context.Context, currentUserID, targetUserID string) (bool, error) {
	if currentUserID == "" || targetUserID == "" {
		return false, nil
	}
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM alfq_follows WHERE follower_id=$1 AND following_id=$2)`,
		currentUserID, targetUserID).Scan(&exists)
	return exists, err
}

// ----- Paginated lists -----

func (r *traderRepo) GetFollowers(ctx context.Context, userID string, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error) {
	args := []interface{}{limit + 1, userID}
	query := `SELECT ` + userInfoSelect + `
		FROM alfq_follows f JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = $2`
	query, args = AppendCursor(query, args, cursor, "f.created_at, f.follower_id", CursorDesc, 3)
	query += ` ORDER BY f.created_at DESC, f.follower_id DESC LIMIT $1`
	return r.executePaginatedUserList(ctx, query, args, limit)
}

func (r *traderRepo) GetFollowing(ctx context.Context, userID string, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error) {
	args := []interface{}{limit + 1, userID}
	query := `SELECT ` + userInfoSelect + `
		FROM alfq_follows f JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = $2`
	query, args = AppendCursor(query, args, cursor, "f.created_at, f.following_id", CursorDesc, 3)
	query += ` ORDER BY f.created_at DESC, f.following_id DESC LIMIT $1`
	return r.executePaginatedUserList(ctx, query, args, limit)
}

// ListRecommendedTraders returns traders ordered by follower_count DESC, user_id DESC (P2).
// Algorithm: simple popularity rank (document option 1).
// Cursor reuses SocialCursor — follower_count stored as Unix seconds in CreatedAt.
func (r *traderRepo) ListRecommendedTraders(ctx context.Context, cursor *SocialCursor, limit int32) ([]*UserInfoRow, *SocialCursor, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	query := `SELECT ` + userInfoSelect + `
		FROM users u`
	args := []interface{}{limit + 1}
	if cursor != nil && cursor.ID != "" {
		query += ` WHERE ((SELECT COUNT(*) FROM alfq_follows WHERE following_id = u.id), u.id) < ($2, $3)`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}
	query += ` ORDER BY (SELECT COUNT(*) FROM alfq_follows WHERE following_id = u.id) DESC, u.id DESC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var users []*UserInfoRow
	for rows.Next() {
		u, err := scanUserInfo(rows)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	hasMore := len(users) > int(limit)
	if hasMore {
		users = users[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(users) > 0 {
		last := users[len(users)-1]
		nextCursor = &SocialCursor{
			CreatedAt: time.Unix(int64(last.FollowerCount), 0),
			ID:        last.UserID,
		}
	}
	return users, nextCursor, nil
}