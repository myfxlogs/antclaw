package com.antclaw.alfq.ui.alerts

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel

@Composable
fun AlertScreen(
    viewModel: AlertViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.uiState.collectAsState()
    var showAddDialog by remember { mutableStateOf(false) }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            TextButton(onClick = onBack) { Text("← 返回") }
            Text("我的警报", style = MaterialTheme.typography.titleLarge)
            TextButton(onClick = { showAddDialog = true }) { Text("+ 添加") }
        }

        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else if (state.subscriptions.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("暂无警报订阅", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = { showAddDialog = true }) { Text("添加警报") }
                }
            }
        } else {
            LazyColumn(modifier = Modifier.padding(horizontal = 16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                items(state.subscriptions, key = { it.subscriptionId }) { sub ->
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
                    ) {
                        Row(
                            modifier = Modifier.fillMaxWidth().padding(16.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Column {
                                Text("${sub.pair} ${sub.condition} ${sub.threshold}", style = MaterialTheme.typography.bodyLarge)
                                Text(sub.alertType, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                            }
                            TextButton(onClick = { viewModel.unsubscribe(sub.subscriptionId) }) {
                                Text("删除", color = MaterialTheme.colorScheme.error)
                            }
                        }
                    }
                }
            }
        }
    }

    if (showAddDialog) {
        AddAlertDialog(
            onDismiss = { showAddDialog = false },
            onSubscribe = { type, pair, cond, thresh ->
                viewModel.subscribe(type, pair, cond, thresh)
                showAddDialog = false
            }
        )
    }
}

@Composable
fun AddAlertDialog(
    onDismiss: () -> Unit,
    onSubscribe: (String, String, String, String) -> Unit
) {
    var type by remember { mutableStateOf("signal") }
    var pair by remember { mutableStateOf("EURUSD") }
    var condition by remember { mutableStateOf("above") }
    var threshold by remember { mutableStateOf("") }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("添加警报") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("类型: ", modifier = Modifier.width(80.dp))
                    listOf("signal", "price", "macro").forEach { t ->
                        FilterChip(selected = type == t, onClick = { type = t }, label = { Text(t) })
                        Spacer(modifier = Modifier.width(4.dp))
                    }
                }
                OutlinedTextField(value = pair, onValueChange = { pair = it }, label = { Text("品种") }, singleLine = true)
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("条件: ", modifier = Modifier.width(80.dp))
                    listOf("above", "below", "change").forEach { c ->
                        FilterChip(selected = condition == c, onClick = { condition = c }, label = { Text(c) })
                        Spacer(modifier = Modifier.width(4.dp))
                    }
                }
                OutlinedTextField(value = threshold, onValueChange = { threshold = it }, label = { Text("阈值") }, singleLine = true)
            }
        },
        confirmButton = {
            Button(onClick = { if (threshold.isNotBlank()) onSubscribe(type, pair, condition, threshold) }) {
                Text("订阅")
            }
        },
        dismissButton = { TextButton(onClick = onDismiss) { Text("取消") } }
    )
}
