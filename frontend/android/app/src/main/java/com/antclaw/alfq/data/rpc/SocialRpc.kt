package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqFeed
import antclaw.v1.AlfqTrader
import antclaw.v1.Notification
import antclaw.v1.Search
import antclaw.v1.Trend
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import com.google.protobuf.MessageLite
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Feed RPC 客户端 — 仅封装 FeedService，不包含 Trader / Notification。
 * FeedRpc 专注于内容流，ProfileRpc 负责交易员档案，SearchRpc/TrendRpc 为发现页服务。
 */
@Singleton
class FeedRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun getFeed(cursor: String = "", pageSize: Int = 20, filter: String = "all") =
        unary("FeedService/GetFeed",
            AlfqFeed.GetFeedRequest.newBuilder().setCursor(cursor).setPageSize(pageSize).setFilter(filter).build(),
            AlfqFeed.FeedResponse::class)

    suspend fun getPost(postId: String) =
        unary("FeedService/GetPost",
            AlfqFeed.GetPostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.Post::class)

    suspend fun createPost(req: AlfqFeed.CreatePostRequest) =
        unary("FeedService/CreatePost", req, AlfqFeed.Post::class)

    suspend fun likePost(postId: String) =
        unary("FeedService/LikePost",
            AlfqFeed.LikePostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.Post::class)

    suspend fun unlikePost(postId: String) =
        unary("FeedService/UnlikePost",
            AlfqFeed.UnlikePostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.Post::class)

    suspend fun commentOnPost(req: AlfqFeed.CommentRequest) =
        unary("FeedService/CommentOnPost", req, AlfqFeed.Comment::class)

    suspend fun listComments(postId: String, cursor: String = "", pageSize: Int = 50) =
        unary("FeedService/ListComments",
            AlfqFeed.ListCommentsRequest.newBuilder().setPostId(postId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqFeed.ListCommentsResponse::class)

    suspend fun sharePost(postId: String, comment: String = "") =
        unary("FeedService/SharePost",
            AlfqFeed.SharePostRequest.newBuilder().setPostId(postId).setComment(comment).build(),
            AlfqFeed.Post::class)

    suspend fun listUserPosts(userId: String, cursor: String = "", pageSize: Int = 20, filter: String = "all") =
        unary("FeedService/ListUserPosts",
            AlfqFeed.ListUserPostsRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).setFilter(filter).build(),
            AlfqFeed.FeedResponse::class)

    // ── Internal helper ──

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Res : MessageLite> unary(
        method: String, req: MessageLite, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec(
            "antclaw.v1.$method",
            req::class,
            resClass,
            StreamType.UNARY,
        )
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}

/**
 * Profile RPC 客户端 — TraderService。
 */
@Singleton
class ProfileRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun getProfile(userId: String) =
        unary("TraderService/GetProfile",
            AlfqTrader.GetTraderProfileRequest.newBuilder().setUserId(userId).build(),
            AlfqTrader.TraderProfile::class)

    suspend fun updateProfile(displayName: String) =
        unary("TraderService/UpdateProfile",
            AlfqTrader.UpdateTraderProfileRequest.newBuilder().setDisplayName(displayName).build(),
            AlfqTrader.TraderProfile::class)

    suspend fun follow(targetUserId: String) =
        unary("TraderService/Follow",
            AlfqTrader.FollowRequest.newBuilder().setTargetUserId(targetUserId).build(),
            AlfqTrader.FollowResponse::class)

    suspend fun unfollow(targetUserId: String) =
        unary("TraderService/Unfollow",
            AlfqTrader.UnfollowRequest.newBuilder().setTargetUserId(targetUserId).build(),
            AlfqTrader.FollowResponse::class)

    suspend fun getFollowers(userId: String, cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/GetFollowers",
            AlfqTrader.GetFollowersRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.UserList::class)

    suspend fun getFollowing(userId: String, cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/GetFollowing",
            AlfqTrader.GetFollowingRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.UserList::class)

    suspend fun listRecommendedTraders(cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/ListRecommendedTraders",
            AlfqTrader.ListRecommendedTradersRequest.newBuilder().setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.UserList::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Res : MessageLite> unary(
        method: String, req: MessageLite, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", req::class, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}

/**
 * Notification RPC 客户端 — NotificationService。
 */
@Singleton
class NotificationRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun unreadCount() =
        unary("NotificationService/UnreadCount",
            Notification.UnreadCountRequest.getDefaultInstance(),
            Notification.UnreadCountResponse::class)

    suspend fun listUnread(limit: Int = 20) =
        unary("NotificationService/ListUnread",
            Notification.ListUnreadRequest.newBuilder().setLimit(limit).build(),
            Notification.ListUnreadResponse::class)

    suspend fun markRead(notificationId: String) =
        unary("NotificationService/MarkRead",
            Notification.MarkReadRequest.newBuilder().setNotificationId(notificationId).build(),
            Notification.MarkReadResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Res : MessageLite> unary(
        method: String, req: MessageLite, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", req::class, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}

/**
 * Search RPC 客户端 — SearchService。
 */
@Singleton
class SearchRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun search(query: String, cursor: String = "", pageSize: Int = 10, scopes: List<String> = emptyList()) =
        unary("SearchService/Search",
            Search.SearchRequest.newBuilder().setQuery(query).setCursor(cursor).setPageSize(pageSize).addAllScopes(scopes).build(),
            Search.SearchResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Res : MessageLite> unary(
        method: String, req: MessageLite, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", req::class, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}

/**
 * Trend RPC 客户端 — TrendService。
 */
@Singleton
class TrendRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun listTrendingTopics(window: String = "24h", limit: Int = 10) =
        unary("TrendService/ListTrendingTopics",
            Trend.ListTrendingTopicsRequest.newBuilder().setWindow(window).setLimit(limit).build(),
            Trend.ListTrendingTopicsResponse::class)

    suspend fun listHotSymbols(window: String = "24h", limit: Int = 10) =
        unary("TrendService/ListHotSymbols",
            Trend.ListHotSymbolsRequest.newBuilder().setWindow(window).setLimit(limit).build(),
            Trend.ListHotSymbolsResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Res : MessageLite> unary(
        method: String, req: MessageLite, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", req::class, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}
