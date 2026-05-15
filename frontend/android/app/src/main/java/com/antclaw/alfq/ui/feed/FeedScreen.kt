package com.antclaw.alfq.ui.feed

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
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
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.UiEvent
import com.antclaw.alfq.ui.theme.*

// ── FeedScreen (≤60 lines) ──

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    notificationCount: Int = 0,
    onSignalClick: (pair: String) -> Unit = {},
    onPostClick: (postId: String) -> Unit = {},
    onAuthorClick: (userId: String) -> Unit = {},
    onNotificationClick: () -> Unit = {},
    onSearchClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(Unit) {
        viewModel.uiEvent.collect { event ->
            if (event is UiEvent.Snackbar) snackbarHostState.showSnackbar(event.message)
        }
    }

    val shouldLoadMore = remember {
        derivedStateOf {
            val layoutInfo = listState.layoutInfo
            val totalItems = layoutInfo.totalItemsCount
            val lastVisible = layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            totalItems > 0 && lastVisible >= totalItems - 3 && state.hasMore && !state.isAppending
        }
    }
    LaunchedEffect(shouldLoadMore.value) {
        if (shouldLoadMore.value) viewModel.loadMore()
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbarHostState) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            FeedTopBar(notificationCount, onSearchClick, onNotificationClick)
            HorizontalDivider(color = MaterialTheme.colorScheme.outline)
            FeedTabs(state.currentTab, onTabClick = { viewModel.load(it) })
            HorizontalDivider(color = MaterialTheme.colorScheme.outline)
            FeedContent(
                posts = state.posts,
                isLoading = state.isLoading,
                isAppending = state.isAppending,
                error = state.error,
                appendError = state.appendError,
                listState = listState,
                onRetry = { viewModel.load() },
                onAppendRetry = { viewModel.loadMore() },
                onPostClick = onPostClick,
                onAuthorClick = onAuthorClick,
                onLikeClick = { viewModel.toggleLike(it) },
                onShareClick = { viewModel.sharePost(it) },
            )
        }
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
private fun FeedTabs(currentTab: HomeFeedTab, onTabClick: (HomeFeedTab) -> Unit) {
    val tabs = HomeFeedTab.entries
    ScrollableTabRow(
        selectedTabIndex = currentTab.ordinal,
        containerColor = MaterialTheme.colorScheme.surface,
        contentColor = MaterialTheme.colorScheme.onSurface,
        edgePadding = SpacingMd,
        indicator = { tabPositions ->
            TabRowDefaults.SecondaryIndicator(
                modifier = Modifier.tabIndicatorOffset(tabPositions[currentTab.ordinal]),
                color = MaterialTheme.colorScheme.primary
            )
        },
        modifier = Modifier.shadow(2.dp)
    ) {
        tabs.forEach { tab ->
            Tab(
                selected = currentTab == tab,
                onClick = { onTabClick(tab) },
                text = {
                    Text(
                        text = when (tab) {
                            HomeFeedTab.RECOMMENDED -> stringResource(R.string.feed_tab_recommended)
                            HomeFeedTab.SIGNALS -> stringResource(R.string.feed_tab_signals)
                            HomeFeedTab.LATEST -> stringResource(R.string.feed_tab_latest)
                        },
                        fontWeight = if (currentTab == tab) FontWeight.Bold else FontWeight.Normal,
                    )
                },
            )
        }
    }
}

@Composable
private fun FeedContent(
    posts: List<PostUi>,
    isLoading: Boolean,
    isAppending: Boolean,
    error: String?,
    appendError: String?,
    listState: androidx.compose.foundation.lazy.LazyListState,
    onRetry: () -> Unit,
    onAppendRetry: () -> Unit,
    onPostClick: (String) -> Unit,
    onAuthorClick: (String) -> Unit,
    onLikeClick: (String) -> Unit,
    onShareClick: (String) -> Unit,
) {
    when {
        isLoading && posts.isEmpty() ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        error != null ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(error, color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(SpacingSm))
                    TextButton(onClick = onRetry) { Text(stringResource(R.string.feed_retry)) }
                }
            }
        posts.isEmpty() ->
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(stringResource(R.string.feed_empty_title), color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.height(SpacingSm))
                    TextButton(onClick = onRetry) { Text(stringResource(R.string.feed_empty_action)) }
                }
            }
        else ->
            LazyColumn(
                state = listState,
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingSm),
                verticalArrangement = Arrangement.spacedBy(SpacingMd),
            ) {
                items(posts, key = { it.postId }) { post ->
                    PostCard(
                        post = post,
                        onLikeClick = { onLikeClick(post.postId) },
                        onCommentClick = { onPostClick(post.postId) },
                        onShareClick = { onShareClick(post.postId) },
                        onCardClick = { onPostClick(post.postId) },
                        onAuthorClick = onAuthorClick,
                    )
                }
                if (isAppending) {
                    item {
                        Box(Modifier.fillMaxWidth().padding(vertical = SpacingMd), contentAlignment = Alignment.Center) {
                            CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
                        }
                    }
                }
                if (appendError != null) {
                    item {
                        TextButton(onClick = onAppendRetry, modifier = Modifier.fillMaxWidth()) {
                            Text(stringResource(R.string.feed_retry))
                        }
                    }
                }
            }
    }
}
