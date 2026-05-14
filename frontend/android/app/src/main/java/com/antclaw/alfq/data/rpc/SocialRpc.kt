package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqFeed
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import com.google.protobuf.MessageLite
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Social RPC 客户端 — 封装 FeedService 全部 RPC。
 */
@Singleton
class SocialRpc @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    // ── Feed ──
    suspend fun getFeed(cursor: String = "", pageSize: Int = 20, filter: String = "all") =
        unary("FeedService/GetFeed", AlfqFeed.GetFeedRequest.newBuilder()
            .setCursor(cursor).setPageSize(pageSize).setFilter(filter).build(),
            AlfqFeed.GetFeedRequest::class, AlfqFeed.FeedResponse::class)

    // ── Post ──
    suspend fun getPost(postId: String) =
        unary("FeedService/GetPost", AlfqFeed.GetPostRequest.newBuilder()
            .setPostId(postId).build(),
            AlfqFeed.GetPostRequest::class, AlfqFeed.Post::class)

    suspend fun createPost(req: AlfqFeed.CreatePostRequest) =
        unary("FeedService/CreatePost", req,
            AlfqFeed.CreatePostRequest::class, AlfqFeed.Post::class)

    // ── Like / Unlike ──
    suspend fun likePost(postId: String) =
        unary("FeedService/LikePost", AlfqFeed.LikePostRequest.newBuilder()
            .setPostId(postId).build(),
            AlfqFeed.LikePostRequest::class, AlfqFeed.Post::class)

    suspend fun unlikePost(postId: String) =
        unary("FeedService/UnlikePost", AlfqFeed.UnlikePostRequest.newBuilder()
            .setPostId(postId).build(),
            AlfqFeed.UnlikePostRequest::class, AlfqFeed.Post::class)

    // ── Comment ──
    suspend fun commentOnPost(postId: String, content: String) =
        unary("FeedService/CommentOnPost", AlfqFeed.CommentRequest.newBuilder()
            .setPostId(postId).setContent(content).build(),
            AlfqFeed.CommentRequest::class, AlfqFeed.Comment::class)

    // ── Share ──
    suspend fun sharePost(postId: String, comment: String = "") =
        unary("FeedService/SharePost", AlfqFeed.SharePostRequest.newBuilder()
            .setPostId(postId).setComment(comment).build(),
            AlfqFeed.SharePostRequest::class, AlfqFeed.Post::class)

    // ── Internal helper ──
    @Suppress("UNCHECKED_CAST")
    private suspend inline fun <reified Req : MessageLite, reified Res : MessageLite> unary(
        method: String,
        req: Req,
        reqClass: kotlin.reflect.KClass<Req>,
        resClass: kotlin.reflect.KClass<Res>,
    ): Res {
        val spec = MethodSpec("antclaw.v1.$method", reqClass, resClass, StreamType.UNARY)
        return client.unary(req, emptyMap(), spec).getOrThrow() as Res
    }
}
