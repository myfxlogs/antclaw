package com.antclaw.alfq.data.repository

import antclaw.v1.AlfqFeed
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.FeedRpc
import com.antclaw.alfq.ui.social.CommentUi
import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.PostVisibility
import com.antclaw.alfq.ui.social.SignalCardUi
import com.antclaw.alfq.ui.social.TraderProfileUi
import com.antclaw.alfq.ui.social.UserInfoUi
import com.antclaw.alfq.data.rpc.ProfileRpc
import java.time.Instant
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 社交数据仓库 — Proto ↔ UI 映射，本地缓存策略。
 * P0 负责 Feed 内容流和社交交互，P1 扩展搜索/趋势。
 */
@Singleton
class SocialRepository @Inject constructor(
    private val feedRpc: FeedRpc,
    private val profileRpc: ProfileRpc,
    private val tokenStore: TokenStore,
) {
    private val currentUserId get() = tokenStore.getUserId().orEmpty()

    // ══════ Feed ══════

    suspend fun getFeed(cursor: String = "", pageSize: Int = 20, filter: String = "all"): Pair<List<PostUi>, String?> {
        val resp = feedRpc.getFeed(cursor, pageSize, filter)
        val posts = resp.postsList.map { it.toPostUi(currentUserId) }
        return posts to resp.nextCursor.takeIf { it.isNotBlank() }
    }

    // ══════ Post ══════

    suspend fun getPost(postId: String): PostUi =
        feedRpc.getPost(postId).toPostUi(currentUserId)

    suspend fun createPost(content: String, signalPair: String = "", signalDirection: String = "",
                           signalConfidence: Int = 0, visibility: String = "public"): PostUi {
        val req = AlfqFeed.CreatePostRequest.newBuilder()
            .setContent(content)
            .setPostType(if (signalPair.isNotBlank()) "signal_card" else "text")
            .setSignalPair(signalPair)
            .setSignalDirection(signalDirection)
            .setSignalConfidence(signalConfidence)
            .setVisibility(visibility)
            .build()
        return feedRpc.createPost(req).toPostUi(currentUserId)
    }

    suspend fun listUserPosts(userId: String, cursor: String = "", pageSize: Int = 20,
                              filter: String = "all"): Pair<List<PostUi>, String?> {
        val resp = feedRpc.listUserPosts(userId, cursor, pageSize, filter)
        val posts = resp.postsList.map { it.toPostUi(currentUserId) }
        return posts to resp.nextCursor.takeIf { it.isNotBlank() }
    }

    // ══════ Like / Unlike ══════

    suspend fun likePost(postId: String): PostUi =
        feedRpc.likePost(postId).toPostUi(currentUserId)

    suspend fun unlikePost(postId: String): PostUi =
        feedRpc.unlikePost(postId).toPostUi(currentUserId)

    // ══════ Comment ══════

    suspend fun commentOnPost(postId: String, content: String, parentCommentId: String? = null): CommentUi {
        val req = AlfqFeed.CommentRequest.newBuilder()
            .setPostId(postId).setContent(content)
        if (!parentCommentId.isNullOrBlank()) req.parentCommentId = parentCommentId
        return feedRpc.commentOnPost(req.build()).toCommentUi()
    }

    suspend fun listComments(postId: String, cursor: String = "", pageSize: Int = 50): Pair<List<CommentUi>, String?> {
        val resp = feedRpc.listComments(postId, cursor, pageSize)
        val comments = resp.commentsList.map { it.toCommentUi() }
        return comments to resp.nextCursor.takeIf { it.isNotBlank() }
    }

    // ══════ Share ══════

    suspend fun sharePost(postId: String, comment: String = ""): PostUi =
        feedRpc.sharePost(postId, comment).toPostUi(currentUserId)

    // ══════ Profile ══════

    suspend fun getProfile(userId: String): TraderProfileUi {
        val p = profileRpc.getProfile(userId)
        return TraderProfileUi(
            userId = p.userId,
            displayName = p.displayName,
            bio = p.bio,
            tier = p.tier,
            followerCount = p.followerCount,
            followingCount = p.followingCount,
            isFollowing = p.isFollowing,
            showWinRate = p.showWinRate,
            showProfitFactor = p.showProfitFactor,
            showSharpe = p.showSharpe,
            showTotalTrades = p.showTotalTrades,
            winRate = p.winRate,
            profitFactor = p.profitFactor,
            sharpeRatio = p.sharpeRatio,
            totalTrades = p.totalTrades,
            createdAt = Instant.ofEpochSecond(p.createdAt),
        )
    }

    suspend fun follow(userId: String): Int {
        val resp = profileRpc.follow(userId)
        return resp.followerCount
    }

    suspend fun unfollow(userId: String): Int {
        val resp = profileRpc.unfollow(userId)
        return resp.followerCount
    }

    // ══════ Proto → UI Mapping ══════

    private fun AlfqFeed.Post.toPostUi(uid: String): PostUi = PostUi(
        postId = id,
        authorId = authorId,
        authorName = authorName,
        content = content,
        postType = mapPostType(postType),
        signalCard = if (postType == "signal_card") SignalCardUi(
            pair = signalPair, direction = signalDirection, confidence = signalConfidence,
        ) else null,
        visibility = mapVisibility(visibility),
        likeCount = likeCount,
        commentCount = commentCount,
        shareCount = shareCount,
        isLiked = uid.isNotBlank() && likedByList.contains(uid),
        createdAt = Instant.ofEpochSecond(createdAt),
        originalPostId = originalPostId.takeIf { it.isNotBlank() },
    )

    private fun AlfqFeed.Comment.toCommentUi(): CommentUi = CommentUi(
        commentId = id,
        postId = postId,
        authorId = authorId,
        authorName = authorName,
        content = content,
        parentCommentId = parentCommentId.takeIf { it.isNotBlank() },
        createdAt = Instant.ofEpochSecond(createdAt),
    )
}

// ── Mapping helpers ──

private fun mapPostType(raw: String): PostType = when (raw) {
    "signal_card" -> PostType.SIGNAL_CARD
    "chart_share" -> PostType.CHART_SHARE
    "share" -> PostType.SHARE
    else -> PostType.TEXT
}

private fun mapVisibility(raw: String): PostVisibility = when (raw) {
    "followers" -> PostVisibility.FOLLOWERS_ONLY
    "circle" -> PostVisibility.CIRCLE_ONLY
    else -> PostVisibility.PUBLIC
}
