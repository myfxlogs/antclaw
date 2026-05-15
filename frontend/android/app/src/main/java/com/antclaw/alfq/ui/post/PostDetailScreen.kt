package com.antclaw.alfq.ui.post

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.antclaw.alfq.ui.components.PostCard
import com.antclaw.alfq.ui.social.CommentSection
import com.antclaw.alfq.ui.social.CommentUi
import com.antclaw.alfq.ui.social.UiEvent
import com.antclaw.alfq.ui.theme.SpacingMd

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostDetailScreen(
    postId: String,
    onBack: () -> Unit = {},
    viewModel: PostDetailViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(postId) { viewModel.loadPost(postId) }
    LaunchedEffect(Unit) {
        viewModel.uiEvent.collect { event ->
            if (event is UiEvent.Snackbar) snackbarHostState.showSnackbar(event.message)
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = { Text("Post", fontWeight = FontWeight.Bold) },
                navigationIcon = { TextButton(onClick = onBack) { Text("Back") } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
            )
        },
    ) { padding ->
        when {
            state.isLoading -> LoadingView(Modifier.fillMaxSize().padding(padding))
            state.error != null -> ErrorView(state.error!!, Modifier.fillMaxSize().padding(padding)) {
                viewModel.loadPost(postId)
            }
            state.post != null -> {
                Column(modifier = Modifier.fillMaxSize().padding(padding)) {
                    PostCard(
                        post = state.post!!,
                        onLikeClick = { viewModel.toggleLike() },
                        onShareClick = { viewModel.sharePost() },
                        modifier = Modifier.padding(horizontal = SpacingMd, vertical = SpacingMd),
                    )
                    HorizontalDivider(color = MaterialTheme.colorScheme.outline.copy(alpha = 0.3f))
                    CommentSection(
                        comments = state.comments,
                        onSendComment = { viewModel.sendComment(it) },
                    )
                }
            }
        }
    }
}

// ── State ──

data class PostDetailState(
    val post: com.antclaw.alfq.ui.social.PostUi? = null,
    val comments: List<CommentUi> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val isLoadingComments: Boolean = false,
    val isAppendingComments: Boolean = false,
    val commentError: String? = null,
    val commentCursor: String? = null,
)

// ── Shared views ──

@Composable
private fun LoadingView(modifier: Modifier = Modifier) {
    Box(modifier, contentAlignment = Alignment.Center) { CircularProgressIndicator() }
}

@Composable
private fun ErrorView(error: String, modifier: Modifier = Modifier, onRetry: () -> Unit) {
    Box(modifier, contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(error, color = MaterialTheme.colorScheme.error)
            Spacer(modifier = Modifier.height(SpacingMd))
            TextButton(onClick = onRetry) { Text("Retry") }
        }
    }
}
