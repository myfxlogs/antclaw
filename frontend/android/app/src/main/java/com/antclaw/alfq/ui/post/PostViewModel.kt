package com.antclaw.alfq.ui.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.error.AppError
import com.antclaw.alfq.data.error.toAppError
import com.antclaw.alfq.data.repository.SocialRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class PostDraft(
    val content: String = "",
    val signalPair: String = "",
    val signalDirection: String = "",
    val signalConfidence: Int = 0,
    val visibility: String = "public",
)

sealed class PostState {
    object Idle : PostState()
    object Loading : PostState()
    object Success : PostState()
    data class Error(val error: AppError) : PostState()
}

@HiltViewModel
class PostViewModel @Inject constructor(
    private val repository: SocialRepository,
) : ViewModel() {

    private val _postState = MutableStateFlow<PostState>(PostState.Idle)
    val postState: StateFlow<PostState> = _postState.asStateFlow()

    /** 最近一次提交失败的草稿，供重试复用。 */
    private var lastFailedDraft: PostDraft? = null

    fun post(draft: PostDraft) {
        if (_postState.value is PostState.Loading) return
        _postState.value = PostState.Loading
        viewModelScope.launch {
            try {
                repository.createPost(
                    draft.content, draft.signalPair,
                    draft.signalDirection, draft.signalConfidence,
                    draft.visibility,
                )
                lastFailedDraft = null
                _postState.value = PostState.Success
            } catch (e: Exception) {
                lastFailedDraft = draft
                _postState.value = PostState.Error(e.toAppError())
            }
        }
    }

    /** 重试上次失败提交的 payload。 */
    fun retry() {
        val draft = lastFailedDraft ?: return
        post(draft)
    }

    fun reset() {
        _postState.value = PostState.Idle
        lastFailedDraft = null
    }
}
