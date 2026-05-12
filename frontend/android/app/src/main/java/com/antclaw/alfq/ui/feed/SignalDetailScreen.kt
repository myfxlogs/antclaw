package com.antclaw.alfq.ui.feed

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.BullGreen
import com.antclaw.alfq.ui.theme.BearRed

@Composable
fun SignalDetailScreen(
    pair: String,
    viewModel: SignalDetailViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.uiState.collectAsState()

    LaunchedEffect(pair) { viewModel.load(pair) }

    Column(modifier = Modifier.fillMaxSize()) {
        // Header
        Row(
            modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            TextButton(onClick = onBack) { Text("← 返回") }
            Text("$pair 信号详情", style = MaterialTheme.typography.titleLarge)
            TextButton(onClick = { viewModel.load(pair) }) { Text("刷新") }
        }

        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (state.error != null) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text(state.error!!, color = MaterialTheme.colorScheme.error)
            }
        } else {
            Column(
                modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                // Price + Direction
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween
                        ) {
                            Text(state.price, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                            val arrow = when (state.direction) { "bullish" -> "↗" "bearish" -> "↘" else -> "→" }
                            val accent = when (state.direction) { "bullish" -> BullGreen "bearish" -> BearRed else -> Color.Gray }
                            Text("$arrow  ${state.confidence}%", style = MaterialTheme.typography.headlineSmall, color = accent)
                        }
                    }
                }

                // K-line Sparkline
                if (state.barCloses.isNotEmpty()) {
                    Card(
                        modifier = Modifier.fillMaxWidth().height(160.dp),
                        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
                    ) {
                        Box(modifier = Modifier.fillMaxSize().padding(12.dp)) {
                            val bullColor = BullGreen
                            val bearColor = BearRed
                            SparklineChart(
                                values = state.barCloses,
                                modifier = Modifier.fillMaxSize(),
                                lineColor = if (state.direction == "bullish") bullColor else bearColor
                            )
                        }
                    }
                }

                // Factor decomposition
                Card(
                    modifier = Modifier.fillMaxWidth(),
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
                ) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text("因子分解", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                        Spacer(modifier = Modifier.height(12.dp))
                        state.factors.forEach { factor ->
                            FactorRow(factor)
                            Spacer(modifier = Modifier.height(8.dp))
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Button(
                        onClick = { /* TODO: Subscribe alert */ },
                        modifier = Modifier.weight(1f)
                    ) { Text("订阅警报") }
                    OutlinedButton(
                        onClick = { /* TODO: Post discussion */ },
                        modifier = Modifier.weight(1f)
                    ) { Text("发帖讨论") }
                }
            }
        }
    }
}

@Composable
fun FactorRow(factor: FactorItem) {
    val barColor = when {
        factor.name == "TA" -> BullGreen
        factor.name == "Macro" -> BearRed
        else -> MaterialTheme.colorScheme.primary
    }
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(factor.name, modifier = Modifier.width(80.dp), style = MaterialTheme.typography.bodyMedium)
        LinearProgressIndicator(
            progress = { factor.value.toFloat() },
            modifier = Modifier.weight(1f).height(8.dp),
            color = barColor,
            trackColor = MaterialTheme.colorScheme.surfaceVariant,
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(String.format("%.2f", factor.value), style = MaterialTheme.typography.bodySmall)
    }
}

@Composable
fun SparklineChart(
    values: List<Float>,
    modifier: Modifier = Modifier,
    lineColor: Color = BullGreen
) {
    if (values.isEmpty()) return
    val min = values.min()
    val max = values.max()
    val range = (max - min).coerceAtLeast(0.01f)

    Canvas(modifier = modifier) {
        val width = size.width
        val height = size.height
        val stepX = width / (values.size - 1).coerceAtLeast(1)

        val path = Path()
        values.forEachIndexed { i, v ->
            val x = i * stepX
            val y = height - ((v - min) / range) * height * 0.85f - height * 0.075f
            if (i == 0) path.moveTo(x, y) else path.lineTo(x, y)
        }
        drawPath(path, color = lineColor, style = Stroke(width = 2.dp.toPx(), cap = StrokeCap.Round))
    }
}
