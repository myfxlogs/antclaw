package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.components.TraderStatRow
import com.antclaw.alfq.ui.feed.AsyncPhase
import com.antclaw.alfq.ui.social.UiEvent

@Composable
fun ProfileScreen(
    userId: String = "me",
    viewModel: ProfileViewModel = hiltViewModel(),
    onBack: () -> Unit = {},
    onPostClick: (postId: String) -> Unit = {},
    onSettingsClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsState()
    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(userId) { viewModel.load(userId) }
    LaunchedEffect(Unit) {
        viewModel.uiEvent.collect { event ->
            if (event is UiEvent.Snackbar) snackbarHostState.showSnackbar(event.message)
        }
    }

    Scaffold(snackbarHost = { SnackbarHost(snackbarHostState) }) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            // Top bar
            Row(Modifier.fillMaxWidth().padding(horizontal = 4.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                IconButton(onClick = onBack) {
                    Icon(Icons.Default.ArrowBack, contentDescription = stringResource(R.string.common_back))
                }
                Text(stringResource(R.string.profile_title), fontWeight = FontWeight.Bold)
                IconButton(onClick = onSettingsClick) {
                    Icon(Icons.Default.Settings, contentDescription = stringResource(R.string.me_settings))
                }
            }

            when {
                state.loading -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
                state.error != null -> {
                    Column(Modifier.fillMaxSize(), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.Center) {
                        Text(state.error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(Modifier.height(8.dp))
                        TextButton(onClick = { viewModel.load(userId) }) { Text(stringResource(R.string.feed_retry)) }
                    }
                }
                else -> LazyColumn(Modifier.fillMaxSize()) {
                    // ── Header ──
                    item {
                        Column(Modifier.fillMaxWidth().padding(16.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                            // Avatar placeholder + Name
                            Surface(Modifier.size(80.dp), shape = MaterialTheme.shapes.extraLarge, color = MaterialTheme.colorScheme.primaryContainer) {
                                Box(contentAlignment = Alignment.Center) {
                                    Text(state.displayName.take(2).uppercase(), style = MaterialTheme.typography.headlineMedium)
                                }
                            }
                            Spacer(Modifier.height(12.dp))
                            Text(state.displayName, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                            if (state.codeId.isNotEmpty()) {
                                Text("@${state.codeId}", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                            }
                            if (state.bio.isNotEmpty()) {
                                Spacer(Modifier.height(8.dp))
                                Text(state.bio, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f))
                            }
                            val tierLabel = when (state.tier) {
                                "verified" -> stringResource(R.string.tier_verified)
                                "elite" -> stringResource(R.string.tier_elite)
                                else -> ""
                            }
                            if (tierLabel.isNotEmpty()) {
                                Spacer(Modifier.height(4.dp))
                                Text(tierLabel, color = MaterialTheme.colorScheme.primary, style = MaterialTheme.typography.labelMedium)
                            }
                            Spacer(Modifier.height(12.dp))

                            // Follow stats + button
                            TraderStatRow(
                                followingLabel = stringResource(R.string.profile_following), followingCount = state.followingCount,
                                followersLabel = stringResource(R.string.profile_followers), followerCount = state.followerCount,
                            )
                            Spacer(Modifier.height(8.dp))
                            Button(
                                onClick = { viewModel.toggleFollow() },
                                enabled = !state.isFollowLoading,
                            ) {
                                if (state.isFollowLoading) {
                                    CircularProgressIndicator(Modifier.size(16.dp), strokeWidth = 2.dp, color = MaterialTheme.colorScheme.onPrimary)
                                } else {
                                    Text(if (state.isFollowing) stringResource(R.string.profile_unfollow) else stringResource(R.string.profile_follow))
                                }
                            }
                        }
                    }

                    // ── Tabs ──
                    item {
                        TabRow(selectedTabIndex = state.currentTab.ordinal) {
                            ProfileTab.entries.forEach { tab ->
                                Tab(
                                    selected = state.currentTab == tab,
                                    onClick = { viewModel.selectTab(tab) },
                                    text = {
                                        Text(
                                            when (tab) {
                                                ProfileTab.POSTS -> stringResource(R.string.profile_tab_posts)
                                                ProfileTab.MEDIA -> stringResource(R.string.profile_tab_media)
                                                ProfileTab.LIKES -> stringResource(R.string.profile_tab_likes)
                                            },
                                            fontWeight = if (state.currentTab == tab) FontWeight.Bold else FontWeight.Normal,
                                        )
                                    },
                                )
                            }
                        }
                    }

                    when (state.currentTab) {
                        ProfileTab.POSTS -> {
                            when {
                                state.postsPhase == AsyncPhase.Loading -> {
                                    item { Box(Modifier.fillMaxWidth().height(200.dp), contentAlignment = Alignment.Center) { CircularProgressIndicator() } }
                                }
                                state.postsError != null -> {
                                    item {
                                        Column(Modifier.fillMaxWidth().padding(32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                                            Text(state.postsError!!, color = MaterialTheme.colorScheme.error)
                                            TextButton(onClick = { viewModel.loadPosts() }) { Text(stringResource(R.string.feed_retry)) }
                                        }
                                    }
                                }
                                state.posts.isEmpty() -> {
                                    item {
                                        Text(
                                            stringResource(R.string.feed_empty_title),
                                            modifier = Modifier.fillMaxWidth().padding(32.dp),
                                            textAlign = TextAlign.Center,
                                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                                        )
                                    }
                                }
                                else -> {
                                    items(state.posts, key = { it.postId }) { post ->
                                        PostCard(
                                            post = post, onPostClick = { onPostClick(post.postId) },
                                            onAuthorClick = {}, onLikeClick = {}, onShareClick = {},
                                        )
                                    }
                                }
                            }
                        }
                        ProfileTab.MEDIA -> {
                            item {
                                Text(
                                    stringResource(R.string.feed_empty_title),
                                    modifier = Modifier.fillMaxWidth().padding(32.dp),
                                    textAlign = TextAlign.Center,
                                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                                )
                            }
                        }
                        ProfileTab.LIKES -> {
                            item {
                                Text(
                                    stringResource(R.string.feed_empty_title),
                                    modifier = Modifier.fillMaxWidth().padding(32.dp),
                                    textAlign = TextAlign.Center,
                                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}
