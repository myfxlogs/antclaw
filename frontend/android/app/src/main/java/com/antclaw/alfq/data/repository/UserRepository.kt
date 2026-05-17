package com.antclaw.alfq.data.repository

import com.antclaw.alfq.data.rpc.UserRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(private val userRpc: UserRpc) {
    data class Me(val userId: String, val displayName: String, val username: String)
    suspend fun getMe(): Me {
        val me = userRpc.getMe().user
        return Me(me.userId, me.displayName.ifEmpty { me.username }, me.username)
    }
}
