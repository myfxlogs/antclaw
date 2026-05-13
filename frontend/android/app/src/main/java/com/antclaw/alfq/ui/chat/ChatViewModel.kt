package com.antclaw.alfq.ui.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqChat
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.ZoneId
import java.time.temporal.ChronoUnit
import javax.inject.Inject

data class ConversationItem(
    val id: String,
    val name: String,
    val lastMessage: String,
    val timeAgo: String,
    val unreadCount: Int,
    val isGroup: Boolean,
)

data class ChatUiState(
    val conversations: List<ConversationItem> = emptyList(),
    val loading: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class ChatViewModel @Inject constructor() : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private fun client() = ConnectTransportProvider.createProtocolClient()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val spec = MethodSpec("antclaw.v1.ChatService/ListConversations",
                    AlfqChat.ListConversationsRequest::class,
                    AlfqChat.ConversationList::class,
                    StreamType.UNARY)
                val resp: ResponseMessage<AlfqChat.ConversationList> =
                    client().unary(AlfqChat.ListConversationsRequest.getDefaultInstance(), emptyMap(), spec)
                val items = resp.getOrThrow().conversationsList.map { c ->
                    ConversationItem(
                        id = c.id,
                        name = c.name,
                        lastMessage = c.lastMessage,
                        timeAgo = formatTimeAgo(c.lastMessageAt),
                        unreadCount = c.unreadCount,
                        isGroup = c.isGroup,
                    )
                }
                _uiState.value = _uiState.value.copy(loading = false, conversations = items)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    loading = false,
                    error = e.message ?: "\u52a0\u8f7d\u5931\u8d25"
                )
            }
        }
    }

    private fun formatTimeAgo(epochSeconds: Long): String {
        if (epochSeconds <= 0) return ""
        val then = Instant.ofEpochSecond(epochSeconds)
        val now = Instant.now()
        val mins = ChronoUnit.MINUTES.between(then, now)
        return when {
            mins < 1 -> "\u521a\u521a"
            mins < 60 -> "${mins}\u5206\u949f\u524d"
            mins < 1440 -> "${mins / 60}\u5c0f\u65f6\u524d"
            else -> "${mins / 1440}\u5929\u524d"
        }
    }
}
