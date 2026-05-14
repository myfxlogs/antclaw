package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqFeed
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Feed RPC 客户端，封装 FeedService 的调用
 */
@Singleton
class FeedRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    /**
     * 创建帖子
     */
    suspend fun createPost(req: AlfqFeed.CreatePostRequest) {
        val spec = MethodSpec(
            path = "antclaw.v1.FeedService/CreatePost",
            requestClass = AlfqFeed.CreatePostRequest::class,
            responseClass = AlfqFeed.Post::class,
            streamType = StreamType.UNARY,
        )
        client.unary(req, emptyMap(), spec).getOrThrow()
    }
}