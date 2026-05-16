// Package postgres provides the Feed Repository for social post operations.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FeedPostRow is the flat DB row used to map alfq_posts.
type FeedPostRow struct {
	ID               string
	AuthorID         string
	AuthorName       string
	Content          string
	PostType         string
	SignalPair       string
	SignalDirection  string
	SignalConfidence int32
	Visibility       string
	CircleID         *string
	LikeCount        int32
	CommentCount     int32
	ShareCount       int32
	CreatedAt        time.Time
	OriginalPostID   *string
	// Ranked fields (computed, not stored in DB)
	Score          float64
	AuthorCredScore float64
	VisibilityScore float64
}

// FeedCommentRow is the flat DB row for alfq_comments.
type FeedCommentRow struct {
	ID              string
	PostID          string
	AuthorID        string
	AuthorName      string
	Content         string
	CreatedAt       time.Time
	ParentCommentID *string
}

// FeedRepository defines data operations for the social feed.
type FeedRepository interface {
	CreatePost(ctx context.Context, row *FeedPostRow) (*FeedPostRow, error)
	GetPost(ctx context.Context, postID string, currentUserID string) (*FeedPostRow, []string, error)
	CheckPostExists(ctx context.Context, postID string) (bool, error)
	CheckPostVisibility(ctx context.Context, postID, currentUserID string) (bool, error)

	GetFeed(ctx context.Context, filter string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error)
	GetFollowingFeed(ctx context.Context, userID string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error)
	ListUserPosts(ctx context.Context, userID, filter string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error)

	LikePost(ctx context.Context, postID, userID string) error
	UnlikePost(ctx context.Context, postID, userID string) error
	GetLikedByUser(ctx context.Context, postID, userID string) (bool, error)

	CreateComment(ctx context.Context, row *FeedCommentRow) (*FeedCommentRow, error)
	ListComments(ctx context.Context, postID string, cursor *SocialCursor, limit int32) ([]*FeedCommentRow, *SocialCursor, error)
	CheckCommentPostID(ctx context.Context, commentID, postID string) (bool, error)

	GetUserName(ctx context.Context, userID string) (string, error)
}

type feedRepo struct{ pool *pgxpool.Pool }

func NewFeedRepository(pool *pgxpool.Pool) FeedRepository { return &feedRepo{pool: pool} }

// ----- shared SQL helpers -----

const feedSelectColumns = `p.id, p.author_id, p.author_name, p.content, p.post_type,
	p.signal_pair, p.signal_direction, p.signal_confidence, p.visibility, p.circle_id,
	(SELECT COUNT(*) FROM alfq_likes WHERE post_id = p.id)::int4 AS like_count,
	(SELECT COUNT(*) FROM alfq_comments WHERE post_id = p.id)::int4 AS comment_count,
	(SELECT COUNT(*) FROM alfq_posts WHERE original_post_id = p.id AND post_type = 'share')::int4 AS share_count,
	p.created_at, p.original_post_id`

// feedInsertReturningColumns is used for INSERT RETURNING (no table alias allowed).
const feedInsertReturningColumns = `id, author_id, author_name, content, post_type,
	signal_pair, signal_direction, signal_confidence, visibility, circle_id,
	0::int4 AS like_count, 0::int4 AS comment_count, 0::int4 AS share_count,
	NOW() AS created_at, original_post_id`

func scanFeedPost(scanner interface{ Scan(dest ...interface{}) error }) (*FeedPostRow, error) {
	var r FeedPostRow
	var circleID, originalPostID *string
	err := scanner.Scan(
		&r.ID, &r.AuthorID, &r.AuthorName, &r.Content, &r.PostType,
		&r.SignalPair, &r.SignalDirection, &r.SignalConfidence, &r.Visibility, &circleID,
		&r.LikeCount, &r.CommentCount, &r.ShareCount,
		&r.CreatedAt, &originalPostID,
	)
	r.CircleID = circleID
	r.OriginalPostID = originalPostID
	return &r, err
}

// appendFeedFilter conditionally adds a post_type filter.
func appendFeedFilter(query, filter string) string {
	switch filter {
	case "signals_only":
		return query + ` AND p.post_type = 'signal_card'`
	case "posts_only":
		return query + ` AND p.post_type = 'text'`
	case "shares":
		return query + ` AND p.post_type = 'share'`
	}
	return query
}



// executePaginatedFeed runs a feed query and returns posts, liked-by arrays, and next cursor.
func (r *feedRepo) executePaginatedFeed(ctx context.Context, query string, args []interface{}, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var posts []*FeedPostRow
	for rows.Next() {
		row, err := scanFeedPost(rows)
		if err != nil {
			return nil, nil, nil, err
		}
		posts = append(posts, row)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}
	hasMore := len(posts) > int(limit)
	if hasMore {
		posts = posts[:limit]
	}
	var nextCursor *SocialCursor
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = &SocialCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	likedBy, err := r.loadLikedByBatch(ctx, posts, currentUserID)
	if err != nil {
		return nil, nil, nil, err
	}
	return posts, likedBy, nextCursor, nil
}

func (r *feedRepo) loadLikedByBatch(ctx context.Context, posts []*FeedPostRow, currentUserID string) ([][]string, error) {
	if currentUserID == "" || len(posts) == 0 {
		return make([][]string, len(posts)), nil
	}
	ids := make([]string, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	rows, err := r.pool.Query(ctx,
		`SELECT post_id FROM alfq_likes WHERE user_id=$1 AND post_id = ANY($2)`,
		currentUserID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	likedSet := make(map[string]bool, len(posts))
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return nil, err
		}
		likedSet[pid] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([][]string, len(posts))
	for i, p := range posts {
		if likedSet[p.ID] {
			result[i] = []string{currentUserID}
		}
	}
	return result, nil
}

// ----- Post CRUD -----

func (r *feedRepo) CreatePost(ctx context.Context, row *FeedPostRow) (*FeedPostRow, error) {
	return scanFeedPost(r.pool.QueryRow(ctx, `
		INSERT INTO alfq_posts (author_id, author_name, content, post_type,
			signal_pair, signal_direction, signal_confidence, visibility, circle_id, original_post_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+feedInsertReturningColumns,
		row.AuthorID, row.AuthorName, row.Content, row.PostType,
		row.SignalPair, row.SignalDirection, row.SignalConfidence, row.Visibility, row.CircleID, row.OriginalPostID,
	))
}

func (r *feedRepo) GetPost(ctx context.Context, postID string, currentUserID string) (*FeedPostRow, []string, error) {
	row, err := scanFeedPost(r.pool.QueryRow(ctx,
		`SELECT `+feedSelectColumns+` FROM alfq_posts p WHERE p.id = $1`, postID))
	if err != nil {
		return nil, nil, err
	}
	var likedBy []string
	if currentUserID != "" {
		liked, err := r.GetLikedByUser(ctx, postID, currentUserID)
		if err != nil {
			return nil, nil, fmt.Errorf("get liked by user: %w", err)
		}
		if liked {
			likedBy = []string{currentUserID}
		}
	}
	return row, likedBy, nil
}

func (r *feedRepo) GetUserName(ctx context.Context, userID string) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(display_name, email) FROM users WHERE id=$1`, userID).Scan(&name)
	return name, err
}

func (r *feedRepo) CheckPostExists(ctx context.Context, postID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM alfq_posts WHERE id=$1)`, postID).Scan(&exists)
	return exists, err
}

// CheckPostVisibility returns true if currentUserID can view the post.
// P0: public always visible; followers check for followers-only; circle returns false.
func (r *feedRepo) CheckPostVisibility(ctx context.Context, postID, currentUserID string) (bool, error) {
	var visibility, authorID string
	err := r.pool.QueryRow(ctx,
		`SELECT visibility, author_id FROM alfq_posts WHERE id=$1`, postID).Scan(&visibility, &authorID)
	if err != nil {
		return false, err
	}
	if visibility == "public" {
		return true, nil
	}
	if currentUserID == "" {
		return false, nil
	}
	if authorID == currentUserID {
		return true, nil
	}
	if visibility == "followers" {
		var follows bool
		err := r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM alfq_follows WHERE follower_id=$1 AND following_id=$2)`,
			currentUserID, authorID).Scan(&follows)
		return follows, err
	}
	// circle: P0 unsupported
	return false, nil
}

// ----- Feed listing -----

func (r *feedRepo) GetFeed(ctx context.Context, filter string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error) {
	args := []interface{}{limit + 1}
	query := `SELECT ` + feedSelectColumns + ` FROM alfq_posts p WHERE p.visibility = 'public'`
	query = appendFeedFilter(query, filter)
	query, args = AppendCursor(query, args, cursor, "p.created_at, p.id", CursorDesc, 2)
	query += ` ORDER BY p.created_at DESC, p.id DESC LIMIT $1`
	return r.executePaginatedFeed(ctx, query, args, limit, currentUserID)
}

func (r *feedRepo) GetFollowingFeed(ctx context.Context, userID string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error) {
	args := []interface{}{limit + 1, userID}
	query := `SELECT ` + feedSelectColumns + ` FROM alfq_posts p
		WHERE p.author_id IN (SELECT following_id FROM alfq_follows WHERE follower_id = $2)
		  AND p.visibility = 'public'`
	query, args = AppendCursor(query, args, cursor, "p.created_at, p.id", CursorDesc, 2)
	query += ` ORDER BY p.created_at DESC, p.id DESC LIMIT $1`
	return r.executePaginatedFeed(ctx, query, args, limit, currentUserID)
}

func (r *feedRepo) ListUserPosts(ctx context.Context, userID, filter string, cursor *SocialCursor, limit int32, currentUserID string) ([]*FeedPostRow, [][]string, *SocialCursor, error) {
	args := []interface{}{limit + 1, userID}
	base := 3
	query := `SELECT ` + feedSelectColumns + ` FROM alfq_posts p WHERE p.author_id = $2`
	// Visibility: author sees all; authenticated visitors see public + followers(if following); unauthenticated see only public.
	if currentUserID == "" {
		query += ` AND p.visibility = 'public'`
	} else if currentUserID != userID {
		query += ` AND (p.visibility = 'public' OR (p.visibility = 'followers' AND EXISTS(SELECT 1 FROM alfq_follows WHERE follower_id = $3 AND following_id = $2)))`
		args = append(args, currentUserID)
		base = 4
	}
	query = appendFeedFilter(query, filter)
	query, args = AppendCursor(query, args, cursor, "p.created_at, p.id", CursorDesc, base)
	query += ` ORDER BY p.created_at DESC, p.id DESC LIMIT $1`
	return r.executePaginatedFeed(ctx, query, args, limit, currentUserID)
}

// ----- Likes -----

func (r *feedRepo) LikePost(ctx context.Context, postID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO alfq_likes (post_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, postID, userID)
	return err
}

func (r *feedRepo) UnlikePost(ctx context.Context, postID, userID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM alfq_likes WHERE post_id=$1 AND user_id=$2`, postID, userID)
	return err
}

func (r *feedRepo) GetLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM alfq_likes WHERE post_id=$1 AND user_id=$2)`, postID, userID).Scan(&exists)
	return exists, err
}

// ----- Comments -----

func (r *feedRepo) CreateComment(ctx context.Context, row *FeedCommentRow) (*FeedCommentRow, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO alfq_comments (post_id, author_id, author_name, content, parent_comment_id)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		row.PostID, row.AuthorID, row.AuthorName, row.Content, row.ParentCommentID,
	).Scan(&row.ID, &row.CreatedAt)
	return row, err
}

func (r *feedRepo) ListComments(ctx context.Context, postID string, cursor *SocialCursor, limit int32) ([]*FeedCommentRow, *SocialCursor, error) {
	args := []interface{}{limit + 1, postID}
	query := `SELECT id, post_id, author_id, author_name, content, created_at, parent_comment_id
		FROM alfq_comments WHERE post_id = $2`
	query, args = AppendCursor(query, args, cursor, "created_at, id", CursorAsc, 3)
	query += ` ORDER BY created_at ASC, id ASC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var comments []*FeedCommentRow
	for rows.Next() {
		c := &FeedCommentRow{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.AuthorName, &c.Content, &c.CreatedAt, &c.ParentCommentID); err != nil {
			return nil, nil, err
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
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
	return comments, nextCursor, nil
}

// CheckCommentPostID returns true if commentID exists and belongs to postID.
func (r *feedRepo) CheckCommentPostID(ctx context.Context, commentID, postID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM alfq_comments WHERE id=$1 AND post_id=$2)`,
		commentID, postID).Scan(&exists)
	return exists, err
}
