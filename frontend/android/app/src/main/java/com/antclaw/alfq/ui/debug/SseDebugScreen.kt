package com.antclaw.alfq.ui.debug

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.data.sse.SseDebugViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SseDebugScreen(
    onBack: () -> Unit,
    viewModel: SseDebugViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("SSE 连接诊断") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "刷新")
                    }
                },
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            // Base URL
            DebugCard("Base URL", state.baseUrl)

            // Token present
            val tokenColor = if (state.hasToken) {
                MaterialTheme.colorScheme.primary
            } else {
                MaterialTheme.colorScheme.error
            }
            DebugCard("Access Token", if (state.hasToken) "已设置" else "未设置", valueColor = tokenColor)

            // SSE State
            val sseColor = when (state.sseState) {
                "CONNECTED" -> MaterialTheme.colorScheme.primary
                "CONNECTING" -> MaterialTheme.colorScheme.tertiary
                "DISCONNECTED" -> MaterialTheme.colorScheme.outline
                "ERROR" -> MaterialTheme.colorScheme.error
                else -> MaterialTheme.colorScheme.onSurface
            }
            DebugCard("SSE 连接状态", state.sseState, valueColor = sseColor)

            // Last error
            DebugCard(
                "最近连接错误",
                state.lastError ?: "无",
                valueColor = if (state.lastError != null) MaterialTheme.colorScheme.error
                else MaterialTheme.colorScheme.onSurface,
            )

            // Last device report
            DebugCard(
                "最近设备上报",
                state.lastDeviceReportResult ?: "无记录",
            )

            Spacer(modifier = Modifier.weight(1f))
            Text(
                "此页面仅供联调和运维诊断使用。",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
            )
        }
    }
}

@Composable
private fun DebugCard(
    label: String,
    value: String,
    valueColor: androidx.compose.ui.graphics.Color = MaterialTheme.colorScheme.onSurface,
) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                label,
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                value,
                style = MaterialTheme.typography.bodyLarge.copy(
                    fontFamily = FontFamily.Monospace,
                    fontWeight = FontWeight.Medium,
                ),
                color = valueColor,
            )
        }
    }
}
