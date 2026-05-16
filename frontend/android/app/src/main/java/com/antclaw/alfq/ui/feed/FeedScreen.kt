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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.UiEvent
import com.antclaw.alfq.ui.theme.*

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
    val tabs = listOf(
        HomeFeedTab.FOLLOWING to "关注",
        HomeFeedTab.RECOMMENDED to "推荐",
        HomeFeedTab.SIGNALS to "信号",
    )

    Scaffold(topBar = {
        TopAppBar(
            title = { Text("AntClaw", fontWeight = FontWeight.Bold) },
            navigationIcon = {
                IconButton(onClick = onMeClick) {
                    Surface(Modifier.size(32.dp), CircleShape, color = MaterialTheme.colorScheme.primaryContainer) {
                        Box(contentAlignment = Alignment.Center) {
                            Text("A", fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.onPrimaryContainer)
                        }
                    }
                }
            },
            actions = {
                IconButton(onClick = onSearchClick) { Icon(Icons.Default.Search, "搜索") }
                BadgedBox(badge = { if (notificationCount > 0) Badge { Text("$notificationCount") } }) {
                    IconButton(onClick = onNotificationClick) { Icon(Icons.Default.Notifications, "通知") }
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

            when (state.phase) {
                AsyncPhase.Loading -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
                else -> {
                    val listState = rememberLazyListState()
                    LaunchedEffect(listState) {
                        snapshotFlow { listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index }
                            .collect { if (it != null && it >= state.posts.size - 3) viewModel.loadMore() }
                    }
                    LazyColumn(state = listState) {
                        items(state.posts, key = { it.postId }) { post ->
                            PostCard(
                                post = post,
                                onPostClick = { onPostClick(post.postId) },
                                onLikeClick = { viewModel.toggleLike(post.postId) },
                                onShareClick = { viewModel.sharePost(post.postId) },
                                onAuthorClick = { onAuthorClick(post.authorId) },
                            )
                        }
                        if (state.hasMore && state.posts.isNotEmpty()) {
                            item {
                                Box(Modifier.fillMaxWidth().padding(16.dp), contentAlignment = Alignment.Center) {
                                    CircularProgressIndicator(Modifier.size(24.dp))
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
