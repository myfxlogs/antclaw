package com.antclaw.alfq.ui.discover

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

data class CircleUi(val id: String, val name: String, val symbol: String, val memberCount: Int)

@HiltViewModel
class DiscoverViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(DiscoverUiState())
    val uiState: StateFlow<DiscoverUiState> = _uiState.asStateFlow()
    init { _uiState.value = DiscoverUiState() }
}

data class DiscoverUiState(
    val circles: List<CircleUi> = listOf(
        CircleUi("1", "EURUSD 日内交易", "EURUSD", 1200),
        CircleUi("2", "黄金交易员", "XAUUSD", 890),
        CircleUi("3", "加密猎人", "BTCUSD", 650),
    ),
    val traders: List<String> = listOf("Alex Chen 🟢", "Li Wei 🔵", "Zhang San"),
)

@Composable
fun DiscoverScreen(viewModel: DiscoverViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsState()
    LazyColumn(modifier = Modifier.fillMaxSize().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        item { Text("热门交易员", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold) }
        items(state.traders.take(3)) { name ->
            Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                Row(Modifier.padding(16.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                    Text(name, style = MaterialTheme.typography.bodyLarge)
                    OutlinedButton(onClick = { }) { Text("关注") }
                }
            }
        }
        item { Spacer(Modifier.height(8.dp)); Text("热门圈子", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold) }
        items(state.circles) { c ->
            Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                Column(Modifier.padding(16.dp)) {
                    Text(c.name, style = MaterialTheme.typography.bodyLarge)
                    Text("${c.symbol} · ${c.memberCount} 成员", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                }
            }
        }
    }
}
