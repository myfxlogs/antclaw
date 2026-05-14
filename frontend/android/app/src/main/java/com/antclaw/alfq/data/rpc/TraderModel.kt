package com.antclaw.alfq.data.rpc

data class TraderProfile(
    val display_name: String,
    val bio: String,
    val tier: String,
    val win_rate: Double,
    val profit_factor: Double,
    val sharpe_ratio: Double,
    val total_trades: Int,
    val follower_count: Int,
    val following_count: Int,
)

data class FollowResp(val follower_count: Int)
