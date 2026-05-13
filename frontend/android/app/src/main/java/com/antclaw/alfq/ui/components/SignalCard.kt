package com.antclaw.alfq.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.ui.feed.FeedCard
import com.antclaw.alfq.ui.theme.BullGreen
import com.antclaw.alfq.ui.theme.BearRed
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

@Composable
fun SignalCard(card: FeedCard, onDetailClick: () -> Unit) {
    Card(
        onClick = onDetailClick,
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
    ) {
        Column(modifier = Modifier.padding(SpacingMd)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(card.author, style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(card.timeAgo, style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                }
            }

            if (card.pair != null) {
                Spacer(modifier = Modifier.height(SpacingSm))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(card.pair, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(card.direction, style = MaterialTheme.typography.labelSmall,
                        color = when (card.direction) {
                            "bullish" -> BullGreen
                            "bearish" -> BearRed
                            else -> MaterialTheme.colorScheme.onSurface
                        })
                    Spacer(modifier = Modifier.width(8.dp))
                    Text("${card.confidence}%", style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.primary)
                }
            }

            if (card.content.isNotBlank()) {
                Spacer(modifier = Modifier.height(SpacingSm))
                Text(card.content, style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.8f))
            }
        }
    }
}
