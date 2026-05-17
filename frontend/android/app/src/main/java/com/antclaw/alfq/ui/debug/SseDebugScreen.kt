package com.antclaw.alfq.ui.debug

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
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
                title = { Text(stringResource(R.string.sse_debug_title)) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = stringResource(R.string.sse_debug_back))
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = stringResource(R.string.device_refresh))
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
            DebugCard(stringResource(R.string.sse_debug_base_url), state.baseUrl)

            // Token present
            val tokenColor = if (state.hasToken) {
                MaterialTheme.colorScheme.primary
            } else {
                MaterialTheme.colorScheme.error
            }
            DebugCard(
                stringResource(R.string.sse_debug_access_token),
                if (state.hasToken) stringResource(R.string.sse_debug_token_present) else stringResource(R.string.sse_debug_token_missing),
                valueColor = tokenColor,
            )

            // SSE State
            val sseColor = when (state.sseState) {
                "CONNECTED" -> MaterialTheme.colorScheme.primary
                "CONNECTING" -> MaterialTheme.colorScheme.tertiary
                "DISCONNECTED" -> MaterialTheme.colorScheme.outline
                "ERROR" -> MaterialTheme.colorScheme.error
                else -> MaterialTheme.colorScheme.onSurface
            }
            DebugCard(stringResource(R.string.sse_debug_state), state.sseState, valueColor = sseColor)

            // Last error
            DebugCard(
                stringResource(R.string.sse_debug_last_error),
                state.lastError ?: stringResource(R.string.sse_debug_none),
                valueColor = if (state.lastError != null) MaterialTheme.colorScheme.error
                else MaterialTheme.colorScheme.onSurface,
            )

            // Last device report
            DebugCard(
                stringResource(R.string.sse_debug_last_report),
                state.lastDeviceReportResult ?: stringResource(R.string.sse_debug_no_record),
            )

            Spacer(modifier = Modifier.weight(1f))
            Text(
                stringResource(R.string.sse_debug_footer),
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
