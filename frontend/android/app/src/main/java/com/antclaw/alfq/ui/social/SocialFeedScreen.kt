package com.antclaw.alfq.ui.social

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

/**
 * 社交 Feed 主页面 — 关注 / 推荐 双 Tab 切换。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SocialFeedScreen(
    viewModel: SocialFeedViewModel = hiltViewModel(),
    onPostClick: (postId: String) -> Unit = {},
    onPostCreate: () -> Unit = {},
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()

    // Load more when reaching the bottom
    val shouldLoadMore = remember {
        derivedStateOf {
            val layoutInfo = listState.layoutInfo
            val totalItems = layoutInfo.totalItemsCount
            val lastVisible = layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: 0
            totalItems > 0 && lastVisible >= totalItems - 3 && state.hasMore && !state.isLoading
        }
    }
    LaunchedEffect(shouldLoadMore.value) {
        if (shouldLoadMore.value) viewModel.loadMore()
    }

    Scaffold(
        topBar = {
            Column {
                TopAppBar(
                    title = { Text(stringResource(R.string.social_title), fontWeight = FontWeight.Bold) },
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = MaterialTheme.colorScheme.background,
                    ),
                )
                // ── Tab Row ──
                TabRow(
                    selectedTabIndex = state.currentTab.ordinal,
                    containerColor = MaterialTheme.colorScheme.background,
                    contentColor = MaterialTheme.colorScheme.onSurface,
                ) {
                    FeedTab.entries.forEach { tab ->
                        Tab(
                            selected = state.currentTab == tab,
                            onClick = { viewModel.loadFeed(tab) },
                            text = {
                                Text(
                                    text = when (tab) {
                                        FeedTab.FOLLOWING -> stringResource(R.string.social_tab_following)
                                        FeedTab.FOR_YOU -> stringResource(R.string.social_tab_for_you)
                                    },
                                    fontWeight = if (state.currentTab == tab) FontWeight.Bold else FontWeight.Normal,
                                )
                            },
                        )
                    }
                }
                HorizontalDivider(color = MaterialTheme.colorScheme.outline.copy(alpha = 0.3f))
            }
        },
        floatingActionButton = {
            FloatingActionButton(
                onClick = onPostCreate,
                containerColor = MaterialTheme.colorScheme.primary,
            ) {
                Icon(Icons.Default.Add, contentDescription = stringResource(R.string.social_new_post), tint = MaterialTheme.colorScheme.onPrimary)
            }
        },
    ) { padding ->
        when {
            state.isLoading && state.posts.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator()
                }
            }
            state.error != null && state.posts.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(state.error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(SpacingSm))
                        TextButton(onClick = { viewModel.refresh() }) {
                            Text(stringResource(R.string.common_retry))
                        }
                    }
                }
            }
            state.posts.isEmpty() -> {
                Box(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = if (state.currentTab == FeedTab.FOLLOWING)
                                stringResource(R.string.social_empty_following)
                            else stringResource(R.string.social_empty_for_you),
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                }
            }
            else -> {
                LazyColumn(
                    state = listState,
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingSm),
                    verticalArrangement = Arrangement.spacedBy(SpacingMd),
                ) {
                    items(state.posts, key = { it.postId }) { post ->
                        PostCard(
                            post = post,
                            onLikeClick = { viewModel.toggleLike(post.postId) },
                            onCommentClick = { onPostClick(post.postId) },
                            onShareClick = { viewModel.sharePost(post.postId) },
                            onCardClick = { onPostClick(post.postId) },
                        )
                    }
                    if (state.hasMore && state.posts.isNotEmpty()) {
                        item {
                            Box(
                                modifier = Modifier.fillMaxWidth().padding(vertical = SpacingMd),
                                contentAlignment = Alignment.Center,
                            ) {
                                CircularProgressIndicator(modifier = Modifier.size(24.dp), strokeWidth = 2.dp)
                            }
                        }
                    }
                }
            }
        }
    }
}
