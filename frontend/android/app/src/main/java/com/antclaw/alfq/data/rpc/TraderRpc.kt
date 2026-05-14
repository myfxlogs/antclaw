package com.antclaw.alfq.data.rpc

import antclaw.v1.AlfqTrader
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject

class TraderRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    private val getProfileMethod = MethodSpec(
        path = "antclaw.v1.TraderService/GetProfile",
        requestClass = AlfqTrader.GetTraderProfileRequest::class,
        responseClass = AlfqTrader.TraderProfile::class,
        streamType = StreamType.UNARY,
    )
    private val followMethod = MethodSpec(
        path = "antclaw.v1.TraderService/Follow",
        requestClass = AlfqTrader.FollowRequest::class,
        responseClass = AlfqTrader.FollowResponse::class,
        streamType = StreamType.UNARY,
    )
    private val unfollowMethod = MethodSpec(
        path = "antclaw.v1.TraderService/Unfollow",
        requestClass = AlfqTrader.UnfollowRequest::class,
        responseClass = AlfqTrader.FollowResponse::class,
        streamType = StreamType.UNARY,
    )

    suspend fun getProfile(userId: String): TraderProfile {
        val request = AlfqTrader.GetTraderProfileRequest.newBuilder().setUserId(userId).build()
        val p = client.unary<AlfqTrader.GetTraderProfileRequest, AlfqTrader.TraderProfile>(
            request, emptyMap(), getProfileMethod
        ).getOrThrow()
        return TraderProfile(display_name = p.displayName, bio = p.bio, tier = p.tier,
            win_rate = p.winRate, profit_factor = p.profitFactor,
            sharpe_ratio = p.sharpeRatio, total_trades = p.totalTrades,
            follower_count = p.followerCount, following_count = p.followingCount)
    }

    suspend fun follow(userId: String): FollowResp {
        val request = AlfqTrader.FollowRequest.newBuilder().setTargetUserId(userId).build()
        val resp = client.unary<AlfqTrader.FollowRequest, AlfqTrader.FollowResponse>(
            request, emptyMap(), followMethod
        ).getOrThrow()
        return FollowResp(follower_count = resp.followerCount)
    }

    suspend fun unfollow(userId: String): FollowResp {
        val request = AlfqTrader.UnfollowRequest.newBuilder().setTargetUserId(userId).build()
        val resp = client.unary<AlfqTrader.UnfollowRequest, AlfqTrader.FollowResponse>(
            request, emptyMap(), unfollowMethod
        ).getOrThrow()
        return FollowResp(follower_count = resp.followerCount)
    }
}
