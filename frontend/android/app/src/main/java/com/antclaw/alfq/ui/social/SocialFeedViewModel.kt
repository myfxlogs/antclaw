package com.antclaw.alfq.ui.social

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
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
                _state.update { it.copy(error = e.message ?: "Load failed", isLoading = false, isRefreshing = false) }
            }
        }
    }

    // ── Actions ──

    fun toggleLike(postId: String) {
        viewModelScope.launch {
            val post = _state.value.posts.find { it.postId == postId } ?: return@launch
            try {
                val updated = if (post.isLiked) repository.unlikePost(postId) else repository.likePost(postId)
                updatePost(postId) { it.copy(isLiked = !it.isLiked, likeCount = updated.likeCount) }
            } catch (_: Exception) { }
        }
    }

    fun sharePost(postId: String) {
        viewModelScope.launch {
            try {
                repository.sharePost(postId)
                updatePost(postId) { it.copy(shareCount = it.shareCount + 1) }
            } catch (_: Exception) { }
        }
    }

    private fun updatePost(postId: String, transform: (PostUi) -> PostUi) {
        _state.update { s -> s.copy(posts = s.posts.map { p -> if (p.postId == postId) transform(p) else p }) }
    }
}
