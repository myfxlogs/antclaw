package com.antclaw.alfq.ui.chat

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ConvUi(val id: String, val name: String, val isGroup: Boolean, val lastMessage: String, val timeAgo: String, val unreadCount: Int = 0)
data class ChatUiState(val conversations: List<ConvUi> = emptyList(), val loading: Boolean = true)

@HiltViewModel
class ChatViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    init { viewModelScope.launch { _uiState.value = ChatUiState(demoConvs, false) } }

    companion object {
        val demoConvs = listOf(
            ConvUi("1", "Alex Chen", false, "这个位置值得关注...", "2m前", 2),
            ConvUi("2", "EURUSD 交易圈", true, "李：收到信号分享", "1h前"),
        )
    }
}

@Composable
fun ChatScreen(viewModel: ChatViewModel = hiltViewModel(), onBack: () -> Unit) {
    val state by viewModel.uiState.collectAsState()
    Column(modifier = Modifier.fillMaxSize()) {
        Row(modifier = Modifier.fillMaxWidth().padding(16.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text("←") }
            Text("消息", style = MaterialTheme.typography.titleLarge)
            TextButton(onClick = { }) { Text("✎") }
        }
        if (state.loading) Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        else LazyColumn(Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            items(state.conversations) { conv ->
                Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Row(Modifier.padding(16.dp), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column(Modifier.weight(1f)) {
                            Row { Text(conv.name, style = MaterialTheme.typography.bodyLarge); if (conv.isGroup) Text(" (群)", style = MaterialTheme.typography.bodySmall) }
                            Text(conv.lastMessage, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                        }
                        Column(horizontalAlignment = Alignment.End) {
                            Text(conv.timeAgo, style = MaterialTheme.typography.labelSmall)
                            if (conv.unreadCount > 0) Badge { Text("${conv.unreadCount}") }
                        }
                    }
                }
            }
        }
    }
}
