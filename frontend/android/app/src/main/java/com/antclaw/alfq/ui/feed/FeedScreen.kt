package com.antclaw.alfq.ui.feed

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.social.UiEvent
import com.antclaw.alfq.ui.theme.*
import kotlinx.coroutines.flow.distinctUntilChanged

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    viewModel: FeedViewModel = hiltViewModel(),
    notificationCount: Int = 0,
    onPostClick: (String) -> Unit = {},
    onAuthorClick: (String) -> Unit = {},
    onNotificationClick: () -> Unit = {},
    onSearchClick: () -> Unit = {},
    onMeClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }
    val context = LocalContext.current

    LaunchedEffect(Unit) {
        viewModel.uiEvent.collect { event ->
            when (event) {
                is UiEvent.Snackbar -> snackbarHostState.showSnackbar(event.message)
                is UiEvent.SnackbarRes -> snackbarHostState.showSnackbar(context.resources.getString(event.resId))
                else -> {}
            }
        }
    }

    val tabs = listOf(
        HomeFeedTab.FOLLOWING to stringResource(R.string.feed_tab_following),
        HomeFeedTab.RECOMMENDED to stringResource(R.string.feed_tab_recommended),
        HomeFeedTab.SIGNALS to stringResource(R.string.feed_tab_signals),
    )

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
        TopAppBar(
            title = { Text(stringResource(R.string.app_name), fontWeight = FontWeight.Bold) },
            navigationIcon = {
                IconButton(onClick = onMeClick) {
                    Surface(Modifier.size(32.dp), CircleShape, color = MaterialTheme.colorScheme.primaryContainer) {
                        Box(contentAlignment = Alignment.Center) {
                            Text(stringResource(R.string.app_name).take(1).uppercase(), fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.onPrimaryContainer)
                        }
                    }
                }
            },
            actions = {
                IconButton(onClick = onSearchClick) {
                    Icon(Icons.Default.Search, contentDescription = stringResource(R.string.feed_search))
                }
                BadgedBox(badge = { if (notificationCount > 0) Badge { Text("$notificationCount") } }) {
                    IconButton(onClick = onNotificationClick) {
                        Icon(Icons.Default.Notifications, contentDescription = stringResource(R.string.feed_notifications))
                    }
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
        )
    }) { padding ->
        Column(Modifier.padding(padding)) {
            TabRow(
                selectedTabIndex = viewModel.currentTab.ordinal,
                containerColor = MaterialTheme.colorScheme.background,
            ) {
                tabs.forEachIndexed { idx, (tab: HomeFeedTab, label: String) ->
                    Tab(
                        selected = viewModel.currentTab == tab,
                        onClick = { viewModel.selectTab(tab) },
                        text = {
                            Text(label,
                                fontWeight = if (viewModel.currentTab == tab) FontWeight.Bold else FontWeight.Normal,
                                color = if (viewModel.currentTab == tab) MaterialTheme.colorScheme.onSurface
                                        else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                        },
                    )
                }
            }

            FeedBody(
                state = state,
                onRetry = { viewModel.retryLoad() },
                onLoadMore = { viewModel.loadMore() },
                onPostClick = onPostClick,
                onAuthorClick = onAuthorClick,
                onLikeClick = { id -> viewModel.toggleLike(id) },
                onShareClick = { id -> viewModel.sharePost(id) },
                onRetryAppend = { viewModel.retryLoadMore() },
            )
        }
    }
}

/**
 * Feed 内容区 — 按首屏四态分流：Loading / Error / Empty / Success（含分页）
 */
@Composable
private fun FeedBody(
    state: TimelineState,
    onRetry: () -> Unit,
    onLoadMore: () -> Unit,
    onPostClick: (String) -> Unit,
    onAuthorClick: (String) -> Unit,
    onLikeClick: (String) -> Unit,
    onShareClick: (String) -> Unit,
    onRetryAppend: () -> Unit,
) {
    // 预先读取字符串资源，避免在非 @Composable 上下文中调用
    val loadingDesc = stringResource(R.string.common_loading)
    val commonError = stringResource(R.string.common_error)
    val feedEmptyTitle = stringResource(R.string.feed_empty_title)
    val feedRetry = stringResource(R.string.feed_retry)
    val commonRetry = stringResource(R.string.common_retry)

    when (state.initialPhase) {
        InitialPhase.Loading -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(
                    modifier = Modifier.semantics { contentDescription = loadingDesc }
                )
            }
        }

        InitialPhase.Error -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(
                        text = stringResource(state.initialError?.userMessageRes ?: R.string.common_error),
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodyMedium,
                    )
                    Spacer(Modifier.height(SpacingSm))
                    Button(onClick = onRetry) {
                        Text(feedRetry)
                    }
                }
            }
        }

        InitialPhase.Empty -> {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text(
                    text = feedEmptyTitle,
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.semantics { contentDescription = feedEmptyTitle },
                )
            }
        }

        InitialPhase.Success, InitialPhase.Idle -> {
            val listState = rememberLazyListState()
            LaunchedEffect(listState) {
                snapshotFlow { listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index }
                    .distinctUntilChanged()
                    .collect { lastVisibleIndex ->
                        if (lastVisibleIndex != null && lastVisibleIndex >= state.posts.size - 3 && state.hasMore) {
                            onLoadMore()
                        }
                    }
            }

            LazyColumn(state = listState) {
                items(state.posts, key = { it.postId }) { post ->
                    PostCard(
                        post = post,
                        onPostClick = { onPostClick(post.postId) },
                        onLikeClick = { onLikeClick(post.postId) },
                        onShareClick = { onShareClick(post.postId) },
                        onAuthorClick = { onAuthorClick(post.authorId) },
                    )
                }
                // 追加加载指示器
                if (state.appendPhase == AppendPhase.Loading) {
                    item {
                        Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                            CircularProgressIndicator(Modifier.size(24.dp))
                        }
                    }
                }
                // 追加错误：不清空已有列表，显示错误和重试
                if (state.appendPhase == AppendPhase.Error) {
                    item {
                        Box(
                            Modifier.fillMaxWidth().padding(16.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Text(
                                    text = stringResource(state.appendError?.userMessageRes ?: R.string.common_error),
                                    color = MaterialTheme.colorScheme.error,
                                    style = MaterialTheme.typography.bodySmall,
                                )
                                Spacer(Modifier.height(4.dp))
                                TextButton(onClick = onRetryAppend) {
                                    Text(commonRetry)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
