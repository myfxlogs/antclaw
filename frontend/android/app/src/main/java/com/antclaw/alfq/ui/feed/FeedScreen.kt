package com.antclaw.alfq.ui.feed

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
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.SignalCard
import com.antclaw.alfq.ui.theme.*

// ── FeedScreen (≤60 lines) ──

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    notificationCount: Int = 0,
    onSignalClick: (pair: String) -> Unit = {},
    onNotificationClick: () -> Unit = {},
    onSearchClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        FeedTopBar(notificationCount, onSearchClick, onNotificationClick)
        HorizontalDivider(color = MaterialTheme.colorScheme.outline)
        FeedTabs()
        HorizontalDivider(color = MaterialTheme.colorScheme.outline)
        if (state.signalBar.isNotEmpty()) {
            SignalBar(state.signalBar, onSignalClick)
            HorizontalDivider(color = MaterialTheme.colorScheme.outline.copy(alpha = 0.2f))
        }
        FeedContent(state.cards, state.loading, state.error, onRetry = { viewModel.load() }, onSignalClick)
    }
}

// ── Sub-composables ──

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun FeedTopBar(notificationCount: Int, onSearch: () -> Unit, onNotify: () -> Unit) {
    TopAppBar(
        title = {
            Text("α", style = MaterialTheme.typography.titleLarge,
                color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Bold)
        },
        actions = {
            IconButton(onClick = onSearch) {
                Icon(Icons.Default.Search, contentDescription = stringResource(R.string.feed_search),
                    tint = MaterialTheme.colorScheme.onSurface)
            }
            IconButton(onClick = onNotify) {
                BadgedBox(badge = {
                    if (notificationCount > 0) {
                        Badge { Text(if (notificationCount > 99) "99+" else notificationCount.toString()) }
                    }
                }) {
                    Icon(Icons.Default.Notifications, contentDescription = stringResource(R.string.feed_notifications),
                        tint = MaterialTheme.colorScheme.onSurface)
                }
            }
        },
        colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        modifier = Modifier.shadow(4.dp)
    )
}

@Composable
private fun FeedTabs() {
    ScrollableTabRow(
        selectedTabIndex = 0,
        containerColor = MaterialTheme.colorScheme.surface,
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
        Tab(selected = true, onClick = {}, text = { Text(stringResource(R.string.feed_tab_hot), fontWeight = FontWeight.Bold) })
        Tab(selected = false, onClick = {}, text = { Text(stringResource(R.string.feed_tab_recommended), fontWeight = FontWeight.Normal) })
        Tab(selected = false, onClick = {}, text = { Text(stringResource(R.string.feed_tab_signals), fontWeight = FontWeight.Normal) })
        Tab(selected = false, onClick = {}, text = { Text(stringResource(R.string.feed_tab_following), fontWeight = FontWeight.Normal) })
    }
}

@Composable
private fun FeedContent(
    cards: List<FeedCard>, loading: Boolean, error: String?,
    onRetry: () -> Unit, onSignalClick: (String) -> Unit,
) {
    when {
        loading ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        error != null ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(error, color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(SpacingSm))
                    TextButton(onClick = onRetry) { Text(stringResource(R.string.feed_retry)) }
                }
            }
        cards.isEmpty() ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(stringResource(R.string.feed_empty_title), color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.height(SpacingSm))
                    Button(onClick = {}) { Text(stringResource(R.string.feed_empty_action)) }
                }
            }
        else ->
            LazyColumn(
                Modifier.fillMaxSize(),
                contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingSm),
                verticalArrangement = Arrangement.spacedBy(SpacingMd),
            ) {
                items(cards, key = { it.id }) { card ->
                    SignalCard(card, onDetailClick = { card.pair?.let { onSignalClick(it) } })
                }
            }
    }
}

@Composable
private fun SignalBar(items: List<SignalBarItem>, onSignalClick: (String) -> Unit) {
    Row(
        Modifier.fillMaxWidth().horizontalScroll(rememberScrollState())
            .padding(horizontal = SpacingMd, vertical = SpacingSm),
        horizontalArrangement = Arrangement.spacedBy(SpacingSm),
    ) {
        items.forEach { item -> SignalChip(item, onClick = { onSignalClick(item.pair) }) }
    }
}

@Composable
fun SignalChip(item: SignalBarItem, onClick: () -> Unit) {
    val textColor = when (item.direction) {
        "bullish" -> BullGreen; "bearish" -> BearRed
        else -> MaterialTheme.colorScheme.onSurface
    }
    val directionIcon = when (item.direction) {
        "bullish" -> stringResource(R.string.direction_bullish)
        "bearish" -> stringResource(R.string.direction_bearish)
        else -> stringResource(R.string.direction_neutral)
    }
    Surface(onClick = onClick, shape = MaterialTheme.shapes.small, color = MaterialTheme.colorScheme.surfaceVariant) {
        Column(Modifier.padding(horizontal = SpacingSm, vertical = SpacingXs),
            horizontalAlignment = Alignment.CenterHorizontally) {
            Text(item.pair, style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold, color = textColor)
            Text("$directionIcon ${item.confidence}%", style = MaterialTheme.typography.labelMedium, color = textColor)
            Text(item.price, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
        }
    }
}
