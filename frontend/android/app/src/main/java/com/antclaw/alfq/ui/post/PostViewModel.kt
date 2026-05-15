package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SocialRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * 帖子发布状态
 */
sealed class PostState {
    object Idle : PostState()
    object Loading : PostState()
    object Success : PostState()
    data class Error(val message: String) : PostState()
}

/**
 * 帖子发布 ViewModel
 */
@HiltViewModel
class PostViewModel @Inject constructor(
    private val repository: SocialRepository,
) : ViewModel() {

    private val _postState = MutableStateFlow<PostState>(PostState.Idle)
    val postState: StateFlow<PostState> = _postState.asStateFlow()

    fun post(content: String, signalPair: String, signalDirection: String, signalConfidence: Int, visibility: String) {
        if (_postState.value is PostState.Loading) return
        viewModelScope.launch {
            _postState.value = PostState.Loading
            try {
                repository.createPost(content, signalPair, signalDirection, signalConfidence, visibility)
                _postState.value = PostState.Success
            } catch (e: Exception) {
                _postState.value = PostState.Error(e.message ?: "发布失败，请重试")
            }
        }
    }

    fun reset() { _postState.value = PostState.Idle }
}