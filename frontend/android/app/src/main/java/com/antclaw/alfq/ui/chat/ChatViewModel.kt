package com.antclaw.alfq.ui.chat

import android.content.Context
import android.text.format.DateUtils
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.ChatRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
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
    @ApplicationContext private val appContext: Context,
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
                _uiState.value = ChatUiState(error = e.message ?: appContext.getString(com.antclaw.alfq.R.string.common_error))
            }
        }
    }

    private fun formatTime(epochSeconds: Long): String {
        if (epochSeconds <= 0) return ""
        return DateUtils.getRelativeTimeSpanString(
            epochSeconds * 1000,
            System.currentTimeMillis(),
            DateUtils.MINUTE_IN_MILLIS
        ).toString()
    }
}
