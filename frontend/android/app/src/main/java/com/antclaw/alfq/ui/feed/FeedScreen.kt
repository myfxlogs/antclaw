package com.antclaw.alfq.ui.feed

import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.SignalCard
import com.antclaw.alfq.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    notificationCount: Int = 0,
    onSignalClick: (pair: String) -> Unit = {},
    onNotificationClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        // Top Bar - 液态玻璃风格
        TopAppBar(
            title = {
                Image(
                    painter = painterResource(id = com.antclaw.alfq.R.drawable.app_logo),
                    contentDescription = "AlfQ",
                    modifier = Modifier.size(32.dp)
                )
            },
            actions = {
                IconButton(onClick = {}) {
                    Icon(Icons.Default.Search, contentDescription = "搜索", tint = MaterialTheme.colorScheme.onSurface)
                }
            IconButton(onClick = onNotificationClick) {
                BadgedBox(badge = {
                    if (notificationCount > 0) {
                        Badge { Text(if (notificationCount > 99) "99+" else notificationCount.toString()) }
                    }
                }) {
                    Icon(Icons.Default.Notifications, contentDescription = "通知", tint = MaterialTheme.colorScheme.onSurface)
                }
            }
            },
            colors = TopAppBarDefaults.topAppBarColors(
                containerColor = GlassSurface.copy(alpha = 0.8f)
            ),
            modifier = Modifier.shadow(4.dp)
        )
        HorizontalDivider(color = MaterialTheme.colorScheme.outline)

        // Tab Row - 液态玻璃风格（热门在前）
        ScrollableTabRow(
            selectedTabIndex = 0,
            containerColor = GlassSurface.copy(alpha = 0.6f),
            contentColor = MaterialTheme.colorScheme.onSurface,
            edgePadding = SpacingMd,
            indicator = { tabPositions ->
                TabRowDefaults.SecondaryIndicator(
                    modifier = Modifier.tabIndicatorOffset(tabPositions[0]),
                    color = MaterialTheme.colorScheme.primary
                )
            },
            modifier = Modifier.shadow(2.dp)
        ) {
            Tab(selected = true, onClick = {}, text = { Text("热门", fontWeight = FontWeight.Bold) })
            Tab(selected = false, onClick = {}, text = { Text("推荐", fontWeight = FontWeight.Normal) })
            Tab(selected = false, onClick = {}, text = { Text("信号", fontWeight = FontWeight.Normal) })
            Tab(selected = false, onClick = {}, text = { Text("关注", fontWeight = FontWeight.Normal) })
        }

        HorizontalDivider(color = MaterialTheme.colorScheme.outline)

        // Signal Bar (horizontal scroll)
        if (state.signalBar.isNotEmpty()) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = SpacingMd, vertical = SpacingSm),
                horizontalArrangement = Arrangement.spacedBy(SpacingSm)
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
                    Spacer(modifier = Modifier.height(SpacingSm))
                    TextButton(onClick = { viewModel.load() }) { Text("重试") }
                }
            }
        } else if (state.cards.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("关注交易员，获取实时信号", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(modifier = Modifier.height(SpacingSm))
                    Button(onClick = { /* Navigate to discover */ }) {
                        Text("去发现")
                    }
                }
            }
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingSm),
                verticalArrangement = Arrangement.spacedBy(SpacingMd)
            ) {
                items(state.cards) { card ->
                    SignalCard(card, onDetailClick = {
                        card.pair?.let { onSignalClick(it) }
                    })
                }
            }
        }
    }
}

@Composable
fun SignalChip(item: SignalBarItem, onClick: () -> Unit) {
    val textColor = when (item.direction) {
        "bullish" -> BullGreen
        "bearish" -> BearRed
        else -> MaterialTheme.colorScheme.onSurface
    }
    val directionIcon = when (item.direction) {
        "bullish" -> "↗"
        "bearish" -> "↘"
        else -> "→"
    }

    Surface(
        onClick = onClick,
        shape = MaterialTheme.shapes.small,
        color = MaterialTheme.colorScheme.surfaceVariant,
        modifier = Modifier
    ) {
        Column(
            modifier = Modifier.padding(horizontal = SpacingSm, vertical = SpacingXs),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(item.pair, style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold, color = textColor)
            Text(
                "$directionIcon ${item.confidence}%",
                style = MaterialTheme.typography.labelMedium, color = textColor
            )
            Text(item.price, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
        }
    }
}
