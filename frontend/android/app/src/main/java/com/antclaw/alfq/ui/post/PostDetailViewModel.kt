package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.CommentUi
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
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

    fun loadPost(postId: String) {
        viewModelScope.launch {
            _state.update { it.copy(isLoading = true, error = null) }
            try {
                val post = repository.getPost(postId)
                _state.update { it.copy(post = post, isLoading = false) }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message ?: "Failed to load post", isLoading = false) }
            }
        }
    }

    fun toggleLike() {
        val post = _state.value.post ?: return
        viewModelScope.launch {
            try {
                val updated = if (post.isLiked) {
                    repository.unlikePost(post.postId)
                } else {
                    repository.likePost(post.postId)
                }
                _state.update { it.copy(post = it.post?.copy(isLiked = !post.isLiked, likeCount = updated.likeCount)) }
            } catch (_: Exception) { }
        }
    }

    fun sharePost() {
        val post = _state.value.post ?: return
        viewModelScope.launch {
            try {
                repository.sharePost(post.postId)
                _state.update { it.copy(post = it.post?.copy(shareCount = post.shareCount + 1)) }
            } catch (_: Exception) { }
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
            } catch (_: Exception) { }
        }
    }
}
