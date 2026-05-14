package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqFeed
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject

data class CreatePostReq(
    val content: String,
    val post_type: String,
    val signal_pair: String,
    val signal_direction: String,
    val signal_confidence: Int,
    val visibility: String,
)

class FeedRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    private val createPostMethod = MethodSpec(
        path = "antclaw.v1.FeedService/CreatePost",
        requestClass = AlfqFeed.CreatePostRequest::class,
        responseClass = AlfqFeed.Post::class,
        streamType = StreamType.UNARY,
    )

    suspend fun createPost(req: CreatePostReq) {
        val request = AlfqFeed.CreatePostRequest.newBuilder()
            .setContent(req.content)
            .setPostType(req.post_type)
            .setSignalPair(req.signal_pair)
            .setSignalDirection(req.signal_direction)
            .setSignalConfidence(req.signal_confidence)
            .setVisibility(req.visibility)
            .build()

        client.unary<AlfqFeed.CreatePostRequest, AlfqFeed.Post>(
            request, emptyMap(), createPostMethod
        ).getOrThrow()
    }
}
