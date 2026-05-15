package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.UiEvent
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class HomeFeedTab(val filter: String) {
    RECOMMENDED("all"),
    SIGNALS("signals_only"),
    LATEST("all"),
}

data class FeedCard(
    val id: String,
    val author: String = "",
    val pair: String? = null,
    val direction: String = "neutral",
    val confidence: Int = 0,
    val content: String = "",
    val timeAgo: String = "",
)

data class FeedUiState(
    val posts: List<PostUi> = emptyList(),
    val currentTab: HomeFeedTab = HomeFeedTab.RECOMMENDED,
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val isAppending: Boolean = false,
    val error: String? = null,
    val appendError: String? = null,
    val nextCursor: String? = null,
    val hasMore: Boolean = true,
)

@HiltViewModel
class FeedViewModel @Inject constructor(
    private val repository: SocialRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(FeedUiState())
    val uiState: StateFlow<FeedUiState> = _uiState.asStateFlow()

    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    init { load(HomeFeedTab.RECOMMENDED) }

    fun load(tab: HomeFeedTab = _uiState.value.currentTab) {
        _uiState.update {
            it.copy(isLoading = true, error = null, appendError = null, currentTab = tab, nextCursor = null)
        }
        fetchFirstPage(tab)
    }

    fun refresh() {
        val tab = _uiState.value.currentTab
        _uiState.update { it.copy(isRefreshing = true, error = null, appendError = null, nextCursor = null) }
        fetchFirstPage(tab)
    }

    fun loadMore() {
        val state = _uiState.value
        val cursor = state.nextCursor ?: return
        if (state.isLoading || state.isAppending || !state.hasMore) return
        _uiState.update { it.copy(isAppending = true, appendError = null) }
        viewModelScope.launch {
            try {
                val (posts, next) = repository.getFeed(cursor, 20, state.currentTab.filter)
                _uiState.update {
                    it.copy(
                        posts = it.posts + posts,
                        nextCursor = next,
                        hasMore = next != null,
                        isAppending = false,
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(isAppending = false, appendError = e.message ?: "加载更多失败") }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载更多失败"))
            }
        }
    }

    fun toggleLike(postId: String) {
        viewModelScope.launch {
            val post = _uiState.value.posts.find { it.postId == postId } ?: return@launch
            val willLike = !post.isLiked
            updatePost(postId) {
                it.copy(
                    isLiked = willLike,
                    likeCount = if (willLike) it.likeCount + 1 else (it.likeCount - 1).coerceAtLeast(0),
                )
            }
            try {
                val updated = if (willLike) repository.likePost(postId) else repository.unlikePost(postId)
                updatePost(postId) { it.copy(likeCount = updated.likeCount) }
            } catch (e: Exception) {
                updatePost(postId) { post }
                _uiEvent.emit(UiEvent.Snackbar("操作失败，已回滚"))
            }
        }
    }

    fun sharePost(postId: String) {
        viewModelScope.launch {
            val post = _uiState.value.posts.find { it.postId == postId } ?: return@launch
            updatePost(postId) { it.copy(shareCount = it.shareCount + 1) }
            try {
                repository.sharePost(postId)
            } catch (e: Exception) {
                updatePost(postId) { post }
                _uiEvent.emit(UiEvent.Snackbar("分享失败"))
            }
        }
    }

    private fun fetchFirstPage(tab: HomeFeedTab) {
        viewModelScope.launch {
            try {
                val (posts, next) = repository.getFeed("", 20, tab.filter)
                _uiState.update {
                    it.copy(
                        posts = posts,
                        nextCursor = next,
                        hasMore = next != null,
                        isLoading = false,
                        isRefreshing = false,
                    )
                }
            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        error = e.message ?: "加载失败",
                        isLoading = false,
                        isRefreshing = false,
                    )
                }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载失败"))
            }
        }
    }

    private fun updatePost(postId: String, transform: (PostUi) -> PostUi) {
        _uiState.update { state ->
            state.copy(posts = state.posts.map { if (it.postId == postId) transform(it) else it })
        }
    }
}
