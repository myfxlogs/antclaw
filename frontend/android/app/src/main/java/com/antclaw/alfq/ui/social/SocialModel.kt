package com.antclaw.alfq.ui.social

import java.time.Instant

// ── Feed Tab ──

enum class FeedTab { FOLLOWING, FOR_YOU }

// ── Feed State ──

data class SocialFeedState(
    val posts: List<PostUi> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val error: String? = null,
    val hasMore: Boolean = true,
    val currentTab: FeedTab = FeedTab.FOLLOWING,
    val nextCursor: String? = null,
)

// ── Post UI Model ──

data class PostUi(
    val postId: String,
    val authorId: String,
    val authorName: String,
    val authorAvatar: String? = null,
    val content: String,
    val postType: PostType,
    val signalCard: SignalCardUi? = null,
    val chartShare: ChartShareUi? = null,
    val visibility: PostVisibility = PostVisibility.PUBLIC,
    val likeCount: Int = 0,
    val commentCount: Int = 0,
    val shareCount: Int = 0,
    val viewCount: Int = 0,
    val isLiked: Boolean = false,
    val isBookmarked: Boolean = false,
    val createdAt: Instant = Instant.EPOCH,
    val originalPostId: String? = null,
)

// ── Comment UI Model ──

data class CommentUi(
    val commentId: String,
    val postId: String,
    val authorId: String,
    val authorName: String,
    val content: String,
    val parentCommentId: String? = null,
    val createdAt: Instant = Instant.EPOCH,
    val replies: List<CommentUi> = emptyList(),
)

// ── Signal Card (embedded in PostUi for signal_card type) ──

data class SignalCardUi(
    val pair: String,
    val direction: String,   // bullish / bearish / neutral
    val confidence: Int,     // 0-100
)

// ── Chart Share (embedded in PostUi for chart_share type) ──

data class ChartShareUi(
    val pair: String,
    val chartUrl: String? = null,
)

// ── One-time UI Events ──

sealed class UiEvent {
    data class Snackbar(val message: String) : UiEvent()
    data class Navigate(val route: String) : UiEvent()
}

// ── Enums ──

enum class PostType { TEXT, SIGNAL_CARD, CHART_SHARE, SHARE }

enum class PostVisibility { PUBLIC, FOLLOWERS_ONLY, CIRCLE_ONLY }

// ── Trader Profile UI Model ──

data class TraderProfileUi(
    val userId: String,
    val displayName: String,
    val bio: String = "",
    val tier: String = "normal",
    val followerCount: Int = 0,
    val followingCount: Int = 0,
    val isFollowing: Boolean = false,
    val showWinRate: Boolean = false,
    val showProfitFactor: Boolean = false,
    val showSharpe: Boolean = false,
    val showTotalTrades: Boolean = false,
    val winRate: Double = 0.0,
    val profitFactor: Double = 0.0,
    val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0,
    val createdAt: Instant = Instant.EPOCH,
)

// ── User Info (follower/following list item) ──

data class UserInfoUi(
    val userId: String,
    val displayName: String,
    val tier: String = "normal",
    val followerCount: Int = 0,
)
