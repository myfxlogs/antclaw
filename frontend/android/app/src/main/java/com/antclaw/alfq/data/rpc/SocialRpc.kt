package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqFeed
import antclaw.v1.AlfqTrader
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
 * Feed RPC 客户端 — FeedService。
 */
@Singleton
class FeedRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun getFeed(cursor: String = "", pageSize: Int = 20, filter: String = "all") =
        unary("FeedService/GetFeed",
            AlfqFeed.GetFeedRequest.newBuilder().setCursor(cursor).setPageSize(pageSize).setFilter(filter).build(),
            AlfqFeed.GetFeedRequest::class, AlfqFeed.FeedResponse::class)

    suspend fun getPost(postId: String) =
        unary("FeedService/GetPost",
            AlfqFeed.GetPostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.GetPostRequest::class, AlfqFeed.Post::class)

    suspend fun createPost(req: AlfqFeed.CreatePostRequest) =
        unary("FeedService/CreatePost", req,
            AlfqFeed.CreatePostRequest::class, AlfqFeed.Post::class)

    suspend fun likePost(postId: String) =
        unary("FeedService/LikePost",
            AlfqFeed.LikePostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.LikePostRequest::class, AlfqFeed.Post::class)

    suspend fun unlikePost(postId: String) =
        unary("FeedService/UnlikePost",
            AlfqFeed.UnlikePostRequest.newBuilder().setPostId(postId).build(),
            AlfqFeed.UnlikePostRequest::class, AlfqFeed.Post::class)

    suspend fun commentOnPost(req: AlfqFeed.CommentRequest) =
        unary("FeedService/CommentOnPost", req,
            AlfqFeed.CommentRequest::class, AlfqFeed.Comment::class)

    suspend fun listComments(postId: String, cursor: String = "", pageSize: Int = 50) =
        unary("FeedService/ListComments",
            AlfqFeed.ListCommentsRequest.newBuilder().setPostId(postId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqFeed.ListCommentsRequest::class, AlfqFeed.ListCommentsResponse::class)

    suspend fun sharePost(postId: String, comment: String = "") =
        unary("FeedService/SharePost",
            AlfqFeed.SharePostRequest.newBuilder().setPostId(postId).setComment(comment).build(),
            AlfqFeed.SharePostRequest::class, AlfqFeed.Post::class)

    suspend fun listUserPosts(userId: String, cursor: String = "", pageSize: Int = 20, filter: String = "all") =
        unary("FeedService/ListUserPosts",
            AlfqFeed.ListUserPostsRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).setFilter(filter).build(),
            AlfqFeed.ListUserPostsRequest::class, AlfqFeed.FeedResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Req : MessageLite, reified Res : MessageLite> unary(
        method: String, req: Req, reqClass: kotlin.reflect.KClass<Req>, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", reqClass, resClass, StreamType.UNARY)
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
            AlfqTrader.GetTraderProfileRequest::class, AlfqTrader.TraderProfile::class)

    suspend fun updateProfile(displayName: String) =
        unary("TraderService/UpdateProfile",
            AlfqTrader.UpdateTraderProfileRequest.newBuilder().setDisplayName(displayName).build(),
            AlfqTrader.UpdateTraderProfileRequest::class, AlfqTrader.TraderProfile::class)

    suspend fun follow(targetUserId: String) =
        unary("TraderService/Follow",
            AlfqTrader.FollowRequest.newBuilder().setTargetUserId(targetUserId).build(),
            AlfqTrader.FollowRequest::class, AlfqTrader.FollowResponse::class)

    suspend fun unfollow(targetUserId: String) =
        unary("TraderService/Unfollow",
            AlfqTrader.UnfollowRequest.newBuilder().setTargetUserId(targetUserId).build(),
            AlfqTrader.UnfollowRequest::class, AlfqTrader.FollowResponse::class)

    suspend fun getFollowers(userId: String, cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/GetFollowers",
            AlfqTrader.GetFollowersRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.GetFollowersRequest::class, AlfqTrader.UserList::class)

    suspend fun getFollowing(userId: String, cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/GetFollowing",
            AlfqTrader.GetFollowingRequest.newBuilder().setUserId(userId).setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.GetFollowingRequest::class, AlfqTrader.UserList::class)

    suspend fun listRecommendedTraders(cursor: String = "", pageSize: Int = 20) =
        unary("TraderService/ListRecommendedTraders",
            AlfqTrader.ListRecommendedTradersRequest.newBuilder().setCursor(cursor).setPageSize(pageSize).build(),
            AlfqTrader.ListRecommendedTradersRequest::class, AlfqTrader.UserList::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Req : MessageLite, reified Res : MessageLite> unary(
        method: String, req: Req, reqClass: kotlin.reflect.KClass<Req>, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", reqClass, resClass, StreamType.UNARY)
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
            Search.SearchRequest::class, Search.SearchResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Req : MessageLite, reified Res : MessageLite> unary(
        method: String, req: Req, reqClass: kotlin.reflect.KClass<Req>, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", reqClass, resClass, StreamType.UNARY)
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
            Trend.ListTrendingTopicsRequest::class, Trend.ListTrendingTopicsResponse::class)

    suspend fun listHotSymbols(window: String = "24h", limit: Int = 10) =
        unary("TrendService/ListHotSymbols",
            Trend.ListHotSymbolsRequest.newBuilder().setWindow(window).setLimit(limit).build(),
            Trend.ListHotSymbolsRequest::class, Trend.ListHotSymbolsResponse::class)

    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Req : MessageLite, reified Res : MessageLite> unary(
        method: String, req: Req, reqClass: kotlin.reflect.KClass<Req>, resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", reqClass, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}
