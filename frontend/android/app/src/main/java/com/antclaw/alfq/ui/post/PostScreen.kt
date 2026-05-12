package com.antclaw.alfq.ui.post

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.CreatePostReq
import com.antclaw.alfq.data.rpc.FeedRpcClient
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class PostViewModel @Inject constructor() : ViewModel() {
    private val client = FeedRpcClient()
    val posted = MutableStateFlow(false)

    fun post(content: String, signalPair: String, signalDirection: String, signalConfidence: Int, visibility: String) {
        viewModelScope.launch {
            try {
                client.createPost(CreatePostReq(
                    content = content,
                    post_type = if (signalPair.isNotBlank()) "signal_card" else "text",
                    signal_pair = signalPair,
                    signal_direction = signalDirection,
                    signal_confidence = signalConfidence,
                    visibility = visibility
                ))
                posted.value = true
            } catch (_: Exception) { }
        }
    }
}

@Composable
fun PostScreen(viewModel: PostViewModel = hiltViewModel()) {
    var content by remember { mutableStateOf("") }
    var signalPair by remember { mutableStateOf("") }
    var visibility by remember { mutableStateOf("public") }
    var showSuccess by remember { mutableStateOf(false) }

    val posted by viewModel.posted.collectAsState()

    LaunchedEffect(posted) { if (posted) { showSuccess = true; content = "" } }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Text("发布", style = MaterialTheme.typography.headlineMedium, color = MaterialTheme.colorScheme.primary)
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(
            value = content,
            onValueChange = { content = it },
            label = { Text("分享你的交易观点...") },
            modifier = Modifier.fillMaxWidth().weight(0.4f),
            maxLines = 8
        )

        Spacer(modifier = Modifier.height(12.dp))
        OutlinedTextField(
            value = signalPair,
            onValueChange = { signalPair = it },
            label = { Text("引用信号（可选，如 EURUSD）") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )

        Spacer(modifier = Modifier.height(12.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            listOf("public" to "公开", "followers" to "关注者", "circle" to "圈子").forEach { (v, label) ->
                FilterChip(selected = visibility == v, onClick = { visibility = v }, label = { Text(label) })
            }
        }

        Spacer(modifier = Modifier.height(16.dp))
        Button(
            onClick = { viewModel.post(content, signalPair, "", 0, visibility) },
            enabled = content.isNotBlank(),
            modifier = Modifier.fillMaxWidth()
        ) { Text("发布") }

        if (showSuccess) {
            Spacer(modifier = Modifier.height(8.dp))
            Text("发布成功！", color = MaterialTheme.colorScheme.primary)
        }
    }
}
