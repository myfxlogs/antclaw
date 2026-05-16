package feed

import (
	"math"
	"sort"
	"time"

	"github.com/antclaw/antclaw/internal/infra/postgres"
)

const (
	recencyDecay   = 0.1  // λ: 时间衰减系数
	engagementNorm = 3600 // 归一化到1小时
)

// RankPosts reorders a list of posts by a weighted score:
// 40% recency (exponential decay), 30% engagement density, 30% author credibility.
func RankPosts(posts []*postgres.FeedPostRow) []*postgres.FeedPostRow {
	now := time.Now()
	for _, p := range posts {
		hoursAgo := now.Sub(p.CreatedAt).Hours()
		engagement := (float64(p.LikeCount+p.CommentCount+p.ShareCount) / (hoursAgo + 1)) * engagementNorm
		p.Score = 0.4*math.Exp(-recencyDecay*hoursAgo) +
			0.3*engagement +
			0.3*p.AuthorCredScore
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Score > posts[j].Score })
	return posts
}
