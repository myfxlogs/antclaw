package feed

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

type FilterDecision struct {
	Allow    bool
	Drop     bool
	DownRank float64
	Reason   string
}

type FilterRule struct {
	Name     string
	Priority int
	Apply    func(ctx context.Context, pg *pgxpool.Pool, post *postgres.FeedPostRow) FilterDecision
}

var defaultRules = []FilterRule{
	{Name: "low_content", Priority: 0, Apply: checkLowContent},
	{Name: "report_penalty", Priority: 1, Apply: checkReportPenalty},
}

func checkLowContent(_ context.Context, _ *pgxpool.Pool, post *postgres.FeedPostRow) FilterDecision {
	if len(post.Content) < 5 && post.PostType == "text" {
		return FilterDecision{Allow: true, DownRank: 0.5, Reason: "low_content"}
	}
	return FilterDecision{Allow: true}
}

func checkReportPenalty(ctx context.Context, pg *pgxpool.Pool, post *postgres.FeedPostRow) FilterDecision {
	if pg == nil {
		return FilterDecision{Allow: true}
	}
	var count int
	if err := pg.QueryRow(ctx, `SELECT COUNT(*) FROM alfq_reports WHERE post_id = $1 AND status = 'pending'`, post.ID).Scan(&count); err != nil {
		return FilterDecision{Allow: true}
	}
	if count >= 3 {
		return FilterDecision{Allow: true, DownRank: 0.3, Reason: "reported"}
	}
	return FilterDecision{Allow: true}
}

// ApplyFilters runs all default visibility rules against a list of posts.
// Posts that are dropped are excluded; others may have their VisibilityScore multiplied.
func ApplyFilters(ctx context.Context, pg *pgxpool.Pool, posts []*postgres.FeedPostRow) []*postgres.FeedPostRow {
	filtered := make([]*postgres.FeedPostRow, 0, len(posts))
	for _, p := range posts {
		allow := true
		if p.VisibilityScore == 0 {
			p.VisibilityScore = 1.0
		}
		for _, rule := range defaultRules {
			d := rule.Apply(ctx, pg, p)
			if d.Drop {
				allow = false
				break
			}
			p.VisibilityScore *= d.DownRank
		}
		if allow {
			p.Score *= p.VisibilityScore
			filtered = append(filtered, p)
		}
	}
	return filtered
}
