package com.antclaw.alfq.ui.feed

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.BullGreen
import com.antclaw.alfq.ui.theme.BearRed

@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    onSignalClick: (pair: String) -> Unit = {}
) {
    val state by viewModel.uiState.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        // Signal Bar (horizontal scroll)
        if (state.signalBar.isNotEmpty()) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                state.signalBar.forEach { item ->
                    SignalChip(item, onClick = { onSignalClick(item.pair) })
                }
            }
            HorizontalDivider(color = MaterialTheme.colorScheme.outline.copy(alpha = 0.2f))
        }

        // Feed Content
        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (state.error != null) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(state.error!!, color = MaterialTheme.colorScheme.error)
                    Spacer(modifier = Modifier.height(8.dp))
                    TextButton(onClick = { viewModel.load() }) { Text("重试") }
                }
            }
        } else if (state.cards.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("关注交易员，获取实时信号", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
            }
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(horizontal = 16.dp, vertical = 8.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                items(state.cards) { card ->
                    FeedCardView(card, onDetailClick = {
                        card.pair?.let { onSignalClick(it) }
                    })
                }
            }
        }
    }
}

@Composable
fun SignalChip(item: SignalBarItem, onClick: () -> Unit) {
    val bgColor = when (item.direction) {
        "bullish" -> BullGreen.copy(alpha = 0.15f)
        "bearish" -> BearRed.copy(alpha = 0.15f)
        else -> MaterialTheme.colorScheme.surfaceVariant
    }
    val textColor = when (item.direction) {
        "bullish" -> BullGreen
        "bearish" -> BearRed
        else -> MaterialTheme.colorScheme.onSurface
    }

    Surface(
        onClick = onClick,
        shape = MaterialTheme.shapes.small,
        color = bgColor,
        modifier = Modifier
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(item.pair, style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold, color = textColor)
            Text(
                "${if (item.direction == "bullish") "↗" else if (item.direction == "bearish") "↘" else "→"} ${item.confidence}%",
                style = MaterialTheme.typography.labelMedium, color = textColor
            )
            Text(item.price, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
        }
    }
}

@Composable
fun FeedCardView(card: FeedCard, onDetailClick: () -> Unit) {
    val bgColor = when (card.direction) {
        "bullish" -> BullGreen.copy(alpha = 0.08f)
        "bearish" -> BearRed.copy(alpha = 0.08f)
        else -> MaterialTheme.colorScheme.surfaceVariant
    }
    val accentColor = when (card.direction) {
        "bullish" -> BullGreen
        "bearish" -> BearRed
        else -> MaterialTheme.colorScheme.onSurface
    }

    Card(
        onClick = onDetailClick,
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = bgColor)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(card.author, style = MaterialTheme.typography.bodySmall, color = accentColor)
                }
                Text(card.timeAgo, style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f))
            }
            if (card.pair != null) {
                Spacer(modifier = Modifier.height(8.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(card.pair, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                    Spacer(modifier = Modifier.width(8.dp))
                    val arrow = when (card.direction) { "bullish" -> "↗" "bearish" -> "↘" else -> "→" }
                    Text(arrow, style = MaterialTheme.typography.titleMedium, color = accentColor)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("${card.confidence}%", style = MaterialTheme.typography.titleMedium, color = accentColor)
                }
            }
            if (card.content.isNotBlank()) {
                Spacer(modifier = Modifier.height(4.dp))
                Text(card.content, style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
            }
        }
    }
}
