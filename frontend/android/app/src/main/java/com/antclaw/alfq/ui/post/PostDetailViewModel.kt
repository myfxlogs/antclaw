package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.CommentUi
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

@HiltViewModel
class PostDetailViewModel @Inject constructor(
    private val repository: SocialRepository,
) : ViewModel() {

    private val _state = MutableStateFlow(PostDetailState())
    val state: StateFlow<PostDetailState> = _state.asStateFlow()
    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    fun loadPost(postId: String) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            try {
                val post = repository.getPost(postId)
                _state.update { it.copy(post = post, isLoading = false) }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message ?: "Failed to load post", isLoading = false) }
            }
            loadComments(postId)
        }
    }

    fun loadComments(postId: String = _state.value.post?.postId ?: return) {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingComments = true, commentError = null) }
            try {
                val (comments, next) = repository.listComments(postId)
                _state.update {
                    it.copy(comments = comments, commentCursor = next, isLoadingComments = false)
                }
            } catch (e: Exception) {
                android.util.Log.e("PostDetail", "Load comments failed: ${e.message}", e)
                _state.update { it.copy(commentError = e.message, isLoadingComments = false) }
            }
        }
    }

    fun loadMoreComments() {
        val state = _state.value
        val cursor = state.commentCursor ?: return
        val postId = state.post?.postId ?: return
        if (state.isAppendingComments) return
        viewModelScope.launch {
            _state.update { it.copy(isAppendingComments = true) }
            try {
                val (comments, next) = repository.listComments(postId, cursor)
                _state.update {
                    it.copy(
                        comments = it.comments + comments,
                        commentCursor = next,
                        isAppendingComments = false,
                    )
                }
            } catch (e: Exception) {
                android.util.Log.e("PostDetail", "Append comments failed: ${e.message}", e)
                _state.update { it.copy(isAppendingComments = false) }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载更多评论失败"))
            }
        }
    }

    fun toggleLike() {
        val post = _state.value.post ?: return
        viewModelScope.launch {
            val willLike = !post.isLiked
            _state.update {
                it.copy(post = it.post?.copy(
                    isLiked = willLike,
                    likeCount = if (willLike) post.likeCount + 1 else (post.likeCount - 1).coerceAtLeast(0),
                ))
            }
            try {
                val updated = if (post.isLiked) {
                    repository.unlikePost(post.postId)
                } else {
                    repository.likePost(post.postId)
                }
                _state.update { it.copy(post = it.post?.copy(likeCount = updated.likeCount)) }
            } catch (e: Exception) {
                android.util.Log.e("PostDetail", "Like/unlike failed: ${e.message}", e)
                _state.update { it.copy(post = post) }
                _uiEvent.emit(UiEvent.Snackbar("操作失败，已回滚"))
            }
        }
    }

    fun sharePost() {
        val post = _state.value.post ?: return
        viewModelScope.launch {
            _state.update { it.copy(post = it.post?.copy(shareCount = post.shareCount + 1)) }
            try {
                repository.sharePost(post.postId)
            } catch (e: Exception) {
                android.util.Log.e("PostDetail", "Share failed: ${e.message}", e)
                _state.update { it.copy(post = post) }
                _uiEvent.emit(UiEvent.Snackbar("分享失败"))
            }
        }
    }

    fun sendComment(content: String) {
        val post = _state.value.post ?: return
        viewModelScope.launch {
            try {
                val comment = repository.commentOnPost(post.postId, content)
                _state.update {
                    it.copy(
                        comments = it.comments + comment,
                        post = it.post?.copy(commentCount = it.post!!.commentCount + 1),
                    )
                }
            } catch (e: Exception) {
                android.util.Log.e("PostDetail", "Send comment failed: ${e.message}", e)
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "评论失败"))
            }
        }
    }
}
