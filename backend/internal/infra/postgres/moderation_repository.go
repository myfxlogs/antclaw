// Package postgres provides the Moderation Repository for admin social governance (A13-P0-04, A13-P0-05).
package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ModerationCaseRow represents a moderation case.
type ModerationCaseRow struct {
	ID         string
	Source     string
	TargetType string
	TargetID   string
	ReporterID *string
	Reason     string
	Priority   string
	Status     string
	AssigneeID *string
	ReviewerID *string
	Notes      *string
	CreatedAt  time.Time
}

// ModerationRepository defines admin operations for social governance.
type ModerationRepository interface {
	// Post listing with admin filters
	ListPostsAdmin(ctx context.Context, keyword, authorID, postType, visibility, status string, cursor *SocialCursor, limit int32) ([]*FeedPostRow, int32, *SocialCursor, error)
	GetPostDetail(ctx context.Context, postID string) (*FeedPostRow, []*FeedCommentRow, int32, error)

	// Comment listing
	ListCommentsAdmin(ctx context.Context, postID, authorID, status string, cursor *SocialCursor, limit int32) ([]*FeedCommentRow, *SocialCursor, error)

	// Content moderation
	UpdateContentStatus(ctx context.Context, targetType, targetID, newStatus string) error
	CreateModerationCase(ctx context.Context, row *ModerationCaseRow) (*ModerationCaseRow, error)

	// Reports
	ListReportsAdmin(ctx context.Context, status, priority, targetType, assigneeID string, cursor *SocialCursor, limit int32) ([]*ModerationCaseRow, int32, *SocialCursor, error)
	GetReportDetail(ctx context.Context, reportID string) (*ModerationCaseRow, error)
	UpdateReportStatus(ctx context.Context, reportID, status, assigneeID, resolvedBy string) error
}

type moderationRepo struct{ pool *pgxpool.Pool }

func NewModerationRepository(pool *pgxpool.Pool) ModerationRepository { return &moderationRepo{pool: pool} }

// ── Posts ──

func (r *moderationRepo) ListPostsAdmin(ctx context.Context, keyword, authorID, postType, visibility, status string, cursor *SocialCursor, limit int32) ([]*FeedPostRow, int32, *SocialCursor, error) {
	query := `SELECT ` + feedSelectColumns + `, COALESCE(p.status, 'active') AS post_status,
		(SELECT COUNT(*)::int4 FROM alfq_reports WHERE post_id = p.id AND status = 'pending') AS report_count
		FROM alfq_posts p LEFT JOIN alfq_post_stats s ON s.post_id = p.id WHERE 1=1`
	args := []interface{}{limit + 1}
	base := 2

	if keyword != "" {
		query += ` AND p.content ILIKE '%' || $` + itoa(base) + ` || '%'`
		args = append(args, keyword)
		base++
	}
	if authorID != "" {
		query += ` AND p.author_id = $` + itoa(base)
		args = append(args, authorID)
		base++
	}
	if postType != "" {
		query += ` AND p.post_type = $` + itoa(base)
		args = append(args, postType)
		base++
	}
	if visibility != "" {
		query += ` AND p.visibility = $` + itoa(base)
		args = append(args, visibility)
		base++
	}
	if status != "" {
		query += ` AND COALESCE(p.status, 'active') = $` + itoa(base)
		args = append(args, status)
		base++
	}
	query, args = AppendCursor(query, args, cursor, "p.created_at, p.id", CursorDesc, base)
	query += ` ORDER BY p.created_at DESC, p.id DESC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()

	var posts []*FeedPostRow
	for rows.Next() {
		p, err := scanFeedPostWithReport(rows)
		if err != nil {
			return nil, 0, nil, err
		}
		posts = append(posts, p)
	}

	// Count total (simplified: count without cursor pagination)
	var total int32
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM alfq_posts WHERE 1=1`).Scan(&total)

	hasMore := len(posts) > int(limit)
	if hasMore {
		posts = posts[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = &SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return posts, total, nextCursor, rows.Err()
}

func scanFeedPostWithReport(scanner interface{ Scan(dest ...interface{}) error }) (*FeedPostRow, error) {
	var r FeedPostRow
	var circleID, originalPostID *string
	var status string
	err := scanner.Scan(
		&r.ID, &r.AuthorID, &r.AuthorName, &r.Content, &r.PostType,
		&r.SignalPair, &r.SignalDirection, &r.SignalConfidence, &r.Visibility, &circleID,
		&r.LikeCount, &r.CommentCount, &r.ShareCount,
		&r.CreatedAt, &originalPostID,
		&status, &r.Score, // report_count stored in Score for reuse
	)
	r.CircleID = circleID
	r.OriginalPostID = originalPostID
	return &r, err
}

func (r *moderationRepo) GetPostDetail(ctx context.Context, postID string) (*FeedPostRow, []*FeedCommentRow, int32, error) {
	row, err := scanFeedPost(r.pool.QueryRow(ctx,
		`SELECT `+feedSelectColumns+` FROM alfq_posts p LEFT JOIN alfq_post_stats s ON s.post_id = p.id WHERE p.id = $1`, postID))
	if err != nil {
		return nil, nil, 0, err
	}

	commentRows, err2 := r.pool.Query(ctx,
		`SELECT id, post_id, author_id, author_name, content, created_at, parent_comment_id
		 FROM alfq_comments WHERE post_id = $1 AND COALESCE(status, 'active') = 'active' ORDER BY created_at ASC LIMIT 100`,
		postID)
	if err2 != nil {
		return row, nil, 0, nil
	}
	defer commentRows.Close()
	var comments []*FeedCommentRow
	for commentRows.Next() {
		c := &FeedCommentRow{}
		_ = commentRows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt, &c.ParentCommentID)
		comments = append(comments, c)
	}

	var reportCount int32
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*)::int4 FROM alfq_reports WHERE post_id = $1 AND status = 'pending'`, postID).Scan(&reportCount)
	return row, comments, reportCount, nil
}

// ── Comments Admin ──

func (r *moderationRepo) ListCommentsAdmin(ctx context.Context, postID, authorID, status string, cursor *SocialCursor, limit int32) ([]*FeedCommentRow, *SocialCursor, error) {
	args := []interface{}{limit + 1}
	base := 2
	query := `SELECT c.id, c.post_id, c.author_id, c.author_name, c.content, c.created_at, c.parent_comment_id
		FROM alfq_comments c WHERE 1=1`
	if postID != "" {
		query += ` AND c.post_id = $` + itoa(base)
		args = append(args, postID)
		base++
	}
	if authorID != "" {
		query += ` AND c.author_id = $` + itoa(base)
		args = append(args, authorID)
		base++
	}
	if status != "" {
		query += ` AND COALESCE(c.status, 'active') = $` + itoa(base)
		args = append(args, status)
		base++
	}
	query, args = AppendCursor(query, args, cursor, "c.created_at, c.id", CursorAsc, base)
	query += ` ORDER BY c.created_at ASC, c.id ASC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var comments []*FeedCommentRow
	for rows.Next() {
		c := &FeedCommentRow{}
		_ = rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt, &c.ParentCommentID)
		comments = append(comments, c)
	}
	hasMore := len(comments) > int(limit)
	if hasMore {
		comments = comments[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(comments) > 0 {
		last := comments[len(comments)-1]
		nextCursor = &SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return comments, nextCursor, rows.Err()
}

// ── Moderation ──

func (r *moderationRepo) UpdateContentStatus(ctx context.Context, targetType, targetID, newStatus string) error {
	table := "alfq_posts"
	if targetType == "comment" {
		table = "alfq_comments"
	}
	_, err := r.pool.Exec(ctx, `UPDATE `+table+` SET status = $1 WHERE id = $2`, newStatus, targetID)
	return err
}

func (r *moderationRepo) CreateModerationCase(ctx context.Context, row *ModerationCaseRow) (*ModerationCaseRow, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO moderation_cases (source, target_type, target_id, reporter_id, reason, priority, status, assignee_id, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at`,
		row.Source, row.TargetType, row.TargetID, row.ReporterID, row.Reason, row.Priority, row.Status, row.AssigneeID, row.Notes,
	).Scan(&row.ID, &row.CreatedAt)
	return row, err
}

// ── Reports ──

func (r *moderationRepo) ListReportsAdmin(ctx context.Context, status, priority, targetType, assigneeID string, cursor *SocialCursor, limit int32) ([]*ModerationCaseRow, int32, *SocialCursor, error) {
	query := `SELECT c.id, c.source, c.target_type, c.target_id, c.reporter_id::text, c.reason, c.priority, c.status, c.assignee_id::text, c.notes, c.created_at
		FROM moderation_cases c WHERE 1=1`
	args := []interface{}{limit + 1}
	base := 2

	if status != "" {
		query += ` AND c.status = $` + itoa(base)
		args = append(args, status)
		base++
	}
	if priority != "" {
		query += ` AND c.priority = $` + itoa(base)
		args = append(args, priority)
		base++
	}
	if targetType != "" {
		query += ` AND c.target_type = $` + itoa(base)
		args = append(args, targetType)
		base++
	}
	if assigneeID != "" {
		query += ` AND c.assignee_id::text = $` + itoa(base)
		args = append(args, assigneeID)
		base++
	}
	query += ` ORDER BY
		CASE c.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 ELSE 4 END,
		c.created_at DESC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()
	var cases []*ModerationCaseRow
	for rows.Next() {
		m := &ModerationCaseRow{}
		_ = rows.Scan(&m.ID, &m.Source, &m.TargetType, &m.TargetID, &m.ReporterID, &m.Reason, &m.Priority, &m.Status, &m.AssigneeID, &m.Notes, &m.CreatedAt)
		cases = append(cases, m)
	}
	var total int32
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM moderation_cases`).Scan(&total)

	hasMore := len(cases) > int(limit)
	if hasMore {
		cases = cases[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(cases) > 0 {
		last := cases[len(cases)-1]
		nextCursor = &SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return cases, total, nextCursor, rows.Err()
}

func (r *moderationRepo) GetReportDetail(ctx context.Context, reportID string) (*ModerationCaseRow, error) {
	m := &ModerationCaseRow{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, source, target_type, target_id, reporter_id::text, reason, priority, status, assignee_id::text, notes, created_at
		 FROM moderation_cases WHERE id = $1`, reportID,
	).Scan(&m.ID, &m.Source, &m.TargetType, &m.TargetID, &m.ReporterID, &m.Reason, &m.Priority, &m.Status, &m.AssigneeID, &m.Notes, &m.CreatedAt)
	return m, err
}

func (r *moderationRepo) UpdateReportStatus(ctx context.Context, reportID, status, assigneeID, resolvedBy string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE moderation_cases SET status = $1, assignee_id = NULLIF($2,'')::uuid, reviewer_id = NULLIF($3,'')::uuid, updated_at = NOW() WHERE id = $4`,
		status, assigneeID, resolvedBy, reportID)
	return err
}

// helper: int to string for parameter numbering
func itoa(n int) string {
	if n <= 0 {
		return "1"
	}
	var digits []byte
	for x := n; x > 0; x /= 10 {
		digits = append([]byte{byte('0' + x%10)}, digits...)
	}
	return string(digits)
}
