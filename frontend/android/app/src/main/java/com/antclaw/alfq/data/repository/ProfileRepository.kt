package com.antclaw.alfq.data.repository

import com.antclaw.alfq.data.rpc.ProfileRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ProfileRepository @Inject constructor(private val profileRpc: ProfileRpc) {
    data class Profile(val displayName: String, val bio: String, val tier: String,
                       val winRate: Double, val profitFactor: Double, val sharpeRatio: Double,
                       val totalTrades: Int, val followerCount: Int, val followingCount: Int)
    suspend fun getProfile(userId: String): Profile {
        val p = profileRpc.getProfile(userId)
        return Profile(p.displayName, p.bio, p.tier, p.winRate, p.profitFactor, p.sharpeRatio, p.totalTrades, p.followerCount, p.followingCount)
    }
    suspend fun follow(userId: String) = profileRpc.follow(userId).followerCount
    suspend fun unfollow(userId: String) = profileRpc.unfollow(userId).followerCount
}
