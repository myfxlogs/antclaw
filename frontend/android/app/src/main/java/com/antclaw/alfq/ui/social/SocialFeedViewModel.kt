package com.antclaw.alfq.ui.social

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
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

@HiltViewModel
class SocialFeedViewModel @Inject constructor(
    private val repository: SocialRepository,
) : ViewModel() {

    private val _state = MutableStateFlow(SocialFeedState())
    val state: StateFlow<SocialFeedState> = _state.asStateFlow()

    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    init { loadFeed(FeedTab.FOLLOWING) }

    // ── Feed Loading ──

    fun loadFeed(tab: FeedTab) {
        _state.update { it.copy(isLoading = true, error = null, currentTab = tab, nextCursor = null) }
        fetchPage("") { posts, cursor ->
            _state.update { it.copy(posts = posts, nextCursor = cursor, hasMore = cursor != null,
                isLoading = false, isRefreshing = false) }
        }
    }

    fun refresh() {
        _state.update { it.copy(isRefreshing = true, error = null) }
        fetchPage("") { posts, cursor ->
            _state.update { it.copy(posts = posts, nextCursor = cursor, hasMore = cursor != null,
                isLoading = false, isRefreshing = false) }
        }
    }

    fun loadMore() {
        val current = _state.value
        if (current.isLoading || !current.hasMore || current.nextCursor == null) return
        _state.update { it.copy(isLoading = true) }
        fetchPage(current.nextCursor) { posts, cursor ->
            _state.update { it.copy(posts = it.posts + posts, nextCursor = cursor,
                hasMore = cursor != null, isLoading = false) }
        }
    }

    private fun fetchPage(cursor: String, onSuccess: (List<PostUi>, String?) -> Unit) {
        viewModelScope.launch {
            try {
                val (posts, next) = repository.getFeed(cursor, 20)
                onSuccess(posts, next)
            } catch (e: Exception) {
                android.util.Log.e("SocialFeed", "Fetch feed failed: ${e.message}", e)
                _state.update { it.copy(error = e.message ?: "加载失败", isLoading = false, isRefreshing = false) }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载失败"))
            }
        }
    }

    // ── Actions ──

    fun toggleLike(postId: String) {
        viewModelScope.launch {
            val post = _state.value.posts.find { it.postId == postId } ?: return@launch
            val willLike = !post.isLiked
            updatePost(postId) { it.copy(isLiked = willLike, likeCount = if (willLike) it.likeCount + 1 else (it.likeCount - 1).coerceAtLeast(0)) }
            try {
                val updated = if (willLike) repository.likePost(postId) else repository.unlikePost(postId)
                updatePost(postId) { it.copy(likeCount = updated.likeCount) }
            } catch (e: Exception) {
                android.util.Log.e("SocialFeed", "Like/unlike failed: ${e.message}", e)
                _uiEvent.emit(UiEvent.Snackbar("操作失败，已回滚"))
                updatePost(postId) { it.copy(isLiked = !willLike, likeCount = if (!willLike) it.likeCount + 1 else (it.likeCount - 1).coerceAtLeast(0)) }
            }
        }
    }

    fun sharePost(postId: String) {
        viewModelScope.launch {
            val post = _state.value.posts.find { it.postId == postId } ?: return@launch
            updatePost(postId) { it.copy(shareCount = it.shareCount + 1) }
            try {
                repository.sharePost(postId)
            } catch (e: Exception) {
                android.util.Log.e("SocialFeed", "Share failed: ${e.message}", e)
                _uiEvent.emit(UiEvent.Snackbar("分享失败"))
                updatePost(postId) { it.copy(shareCount = (it.shareCount - 1).coerceAtLeast(0)) }
            }
        }
    }

    private fun updatePost(postId: String, transform: (PostUi) -> PostUi) {
        _state.update { s -> s.copy(posts = s.posts.map { p -> if (p.postId == postId) transform(p) else p }) }
    }
}
