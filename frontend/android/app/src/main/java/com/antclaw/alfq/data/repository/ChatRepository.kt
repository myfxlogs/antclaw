package com.antclaw.alfq.data.repository

import com.antclaw.alfq.data.rpc.ChatRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChatRepository @Inject constructor(private val rpc: ChatRpc) {
    data class Conv(val id: String, val name: String, val lastMessage: String, val lastMessageAt: Long, val unreadCount: Int, val isGroup: Boolean)
    suspend fun listConversations(): List<Conv> = rpc.listConversations().conversationsList.map { c ->
        Conv(c.id, c.name, c.lastMessage, c.lastMessageAt, c.unreadCount, c.isGroup)
    }
}
