package com.antclaw.alfq.data.repository

import antclaw.v1.AlfqFeed
import com.antclaw.alfq.data.rpc.SocialRpc
import com.antclaw.alfq.ui.social.CommentUi
import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.PostVisibility
import com.antclaw.alfq.ui.social.SignalCardUi
import java.time.Instant
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 社交数据仓库 — Proto ↔ UI 映射，本地缓存策略。
 */
@Singleton
class SocialRepository @Inject constructor(
    private val rpc: SocialRpc,
) {
    // ── Feed ──

    suspend fun getFeed(cursor: String = "", pageSize: Int = 20): Pair<List<PostUi>, String?> {
        val resp = rpc.getFeed(cursor, pageSize)
        val posts = resp.postsList.map { it.toPostUi() }
        val nextCursor = resp.nextCursor.ifEmpty { null }
        return posts to nextCursor
    }

    // ── Post ──

    suspend fun getPost(postId: String): PostUi =
        rpc.getPost(postId).toPostUi()

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
        return rpc.createPost(req).toPostUi()
    }

    // ── Like / Unlike ──

    suspend fun likePost(postId: String): PostUi =
        rpc.likePost(postId).toPostUi()

    suspend fun unlikePost(postId: String): PostUi =
        rpc.unlikePost(postId).toPostUi()

    // ── Comment ──

    suspend fun commentOnPost(postId: String, content: String): CommentUi =
        rpc.commentOnPost(postId, content).toCommentUi()

    // ── Share ──

    suspend fun sharePost(postId: String, comment: String = ""): PostUi =
        rpc.sharePost(postId, comment).toPostUi()

    // ── Proto → UI Mapping ──

    private fun AlfqFeed.Post.toPostUi(): PostUi = PostUi(
        postId = id,
        authorId = authorId,
        authorName = authorName,
        content = content,
        postType = mapPostType(postType),
        signalCard = if (postType == "signal_card") SignalCardUi(
            pair = signalPair,
            direction = signalDirection,
            confidence = signalConfidence,
        ) else null,
        visibility = mapVisibility(visibility),
        likeCount = likeCount,
        commentCount = commentCount,
        shareCount = shareCount,
        isLiked = false, // resolved by ViewModel against current user
        createdAt = Instant.ofEpochSecond(createdAt),
    )

    private fun AlfqFeed.Comment.toCommentUi(): CommentUi = CommentUi(
        commentId = id,
        postId = postId,
        authorId = authorId,
        authorName = authorName,
        content = content,
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
