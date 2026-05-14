package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqFeed
import com.antclaw.alfq.data.rpc.SocialRpc
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
    private val socialRpc: SocialRpc,
) : ViewModel() {

    private val _postState = MutableStateFlow<PostState>(PostState.Idle)
    val postState: StateFlow<PostState> = _postState.asStateFlow()

    fun post(content: String, signalPair: String, signalDirection: String, signalConfidence: Int, visibility: String) {
        if (_postState.value is PostState.Loading) return
        viewModelScope.launch {
            _postState.value = PostState.Loading
            try {
                socialRpc.createPost(AlfqFeed.CreatePostRequest.newBuilder()
                    .setContent(content)
                    .setPostType(if (signalPair.isBlank()) "text" else "signal_card")
                    .setSignalPair(signalPair)
                    .setSignalDirection(signalDirection)
                    .setSignalConfidence(signalConfidence)
                    .setVisibility(visibility)
                    .build())
                _postState.value = PostState.Success
            } catch (e: Exception) {
                _postState.value = PostState.Error(e.message ?: "发布失败，请重试")
            }
        }
    }

    fun reset() { _postState.value = PostState.Idle }
}