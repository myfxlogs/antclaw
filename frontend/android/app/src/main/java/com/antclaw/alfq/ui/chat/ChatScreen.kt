package com.antclaw.alfq.ui.chat

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R

@Composable
fun ChatScreen(viewModel: ChatViewModel = hiltViewModel(), onBack: () -> Unit) {
    val state by viewModel.uiState.collectAsState()
    Column(modifier = Modifier.fillMaxSize()) {
        Row(modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) }
            Text(stringResource(R.string.chat_title), style = MaterialTheme.typography.titleLarge)
        }
        if (state.loading) Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        else if (state.error != null) Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Text(state.error!!, color = MaterialTheme.colorScheme.error)
                TextButton(onClick = { viewModel.load() }) { Text(stringResource(R.string.common_retry)) }
            }
        }
        else LazyColumn(Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            items(state.conversations) { conv ->
                Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Row(Modifier.padding(16.dp), horizontalArrangement = Arrangement.SpaceBetween) {
                        Column(Modifier.weight(1f)) {
                            Row { Text(conv.name, style = MaterialTheme.typography.bodyLarge); if (conv.isGroup) Text(stringResource(R.string.chat_group_suffix), style = MaterialTheme.typography.bodySmall) }
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
