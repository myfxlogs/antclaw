package com.antclaw.alfq.data.repository

import antclaw.v1.AlfqTrader
import antclaw.v1.Alerts
import antclaw.v1.UserOuterClass
import antclaw.v1.Signals
import antclaw.v1.Price
import antclaw.v1.AlfqChat
import com.antclaw.alfq.data.rpc.AlertRpc
import com.antclaw.alfq.data.rpc.ChatRpc
import com.antclaw.alfq.data.rpc.ProfileRpc
import com.antclaw.alfq.data.rpc.SignalRpc
import com.antclaw.alfq.data.rpc.PriceRpc
import com.antclaw.alfq.data.rpc.UserRpc
import javax.inject.Inject
import javax.inject.Singleton

/** AlertRepository */
@Singleton
class AlertRepository @Inject constructor(private val rpc: AlertRpc) {
    data class Sub(val id: String, val pair: String, val condition: String, val threshold: String, val type: String, val active: Boolean)
    suspend fun listSubscriptions(): List<Sub> = rpc.listSubscriptions().subscriptionsList.map { s ->
        Sub(s.subscriptionId, s.pair, s.condition, s.threshold, s.alertType, s.active)
    }
    suspend fun subscribe(type: String, pair: String, condition: String, threshold: String) {
        rpc.subscribe(Alerts.SubscribeRequest.newBuilder().setAlertType(type).setPair(pair).setCondition(condition).setThreshold(threshold).build())
    }
    suspend fun unsubscribe(id: String) { rpc.unsubscribe(id) }
}

/** ChatRepository */
@Singleton
class ChatRepository @Inject constructor(private val rpc: ChatRpc) {
    data class Conv(val id: String, val name: String, val lastMessage: String, val lastMessageAt: Long, val unreadCount: Int, val isGroup: Boolean)
    suspend fun listConversations(): List<Conv> = rpc.listConversations().conversationsList.map { c ->
        Conv(c.id, c.name, c.lastMessage, c.lastMessageAt, c.unreadCount, c.isGroup)
    }
}

/** UserRepository */
@Singleton
class UserRepository @Inject constructor(private val userRpc: UserRpc) {
    data class Me(val userId: String, val displayName: String, val username: String)
    suspend fun getMe(): Me {
        val me = userRpc.getMe().user
        return Me(me.userId, me.displayName.ifEmpty { me.username }, me.username)
    }
}

/** ProfileRepository */
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

/** DiscoverRepository */
@Singleton
class DiscoverRepository @Inject constructor(
    private val profileRpc: ProfileRpc,
    private val searchRpc: com.antclaw.alfq.data.rpc.SearchRpc,
    private val trendRpc: com.antclaw.alfq.data.rpc.TrendRpc,
) {
    data class Trader(val userId: String, val displayName: String, val tier: String, val followerCount: Int)
    suspend fun listFollowing(userId: String = ""): List<Trader> =
        profileRpc.getFollowing(userId, pageSize = 20).usersList.map { Trader(it.userId, it.displayName, it.tier, it.followerCount) }
}

/** SignalRepository */
@Singleton
class SignalRepository @Inject constructor(
    private val signalRpc: SignalRpc,
    private val priceRpc: PriceRpc,
) {
    data class SignalDetail(val direction: String, val confidence: Int, val price: String)
    suspend fun getDetail(pair: String): SignalDetail {
        val signal = signalRpc.getUnified(pair)
        val price = priceRpc.getPrice(pair)
        return SignalDetail(
            direction = if (signal.hasSignal()) signal.signal.direction else "neutral",
            confidence = if (signal.hasSignal()) ((signal.signal.confidence) * 100).toInt() else 0,
            price = price.current.ifEmpty { "--" },
        )
    }
}
