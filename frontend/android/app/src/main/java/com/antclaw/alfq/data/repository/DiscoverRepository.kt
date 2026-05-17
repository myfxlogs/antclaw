package com.antclaw.alfq.data.repository

import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.data.rpc.SearchRpc
import com.antclaw.alfq.data.rpc.TrendRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class DiscoverRepository @Inject constructor(
    private val profileRpc: ProfileRpc,
    private val searchRpc: SearchRpc,
    private val trendRpc: TrendRpc,
) {
    data class Trader(val userId: String, val displayName: String, val tier: String, val followerCount: Int)
    suspend fun listFollowing(userId: String = ""): List<Trader> =
        profileRpc.getFollowing(userId, pageSize = 20).usersList.map { Trader(it.userId, it.displayName, it.tier, it.followerCount) }
}
