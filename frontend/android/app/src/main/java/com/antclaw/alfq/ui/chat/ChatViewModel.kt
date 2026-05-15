package com.antclaw.alfq.ui.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.ChatRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.temporal.ChronoUnit
import javax.inject.Inject

data class ConversationItem(
    val id: String, val name: String, val lastMessage: String,
    val timeAgo: String, val unreadCount: Int, val isGroup: Boolean,
)
data class ChatUiState(
    val conversations: List<ConversationItem> = emptyList(),
    val loading: Boolean = false, val error: String? = null,
)

@HiltViewModel
class ChatViewModel @Inject constructor(
    private val chatRepo: ChatRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = ChatUiState(loading = true)
            try {
                val items = chatRepo.listConversations().map { c ->
                    ConversationItem(c.id, c.name, c.lastMessage, formatTime(c.lastMessageAt), c.unreadCount, c.isGroup)
                }
                _uiState.value = ChatUiState(conversations = items)
            } catch (e: Exception) {
                _uiState.value = ChatUiState(error = e.message ?: "加载失败")
            }
        }
    }

    private fun formatTime(epochSeconds: Long): String {
        if (epochSeconds <= 0) return ""
        val mins = ChronoUnit.MINUTES.between(Instant.ofEpochSecond(epochSeconds), Instant.now())
        return when { mins < 1 -> "刚刚"; mins < 60 -> "${mins}分钟前"; mins < 1440 -> "${mins / 60}小时前"; else -> "${mins / 1440}天前" }
    }
}
