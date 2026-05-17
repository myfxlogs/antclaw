package com.antclaw.alfq.ui.feed

import com.antclaw.alfq.R
import com.antclaw.alfq.data.error.toAppError
import com.antclaw.alfq.data.error.AppError
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.UiEvent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** 首屏加载阶段 */
enum class InitialPhase { Idle, Loading, Success, Empty, Error }

/** 追加加载阶段 */
enum class AppendPhase { Idle, Loading, Error }

/** 信息流统一状态 — 首屏/追加拆分为独立阶段，避免错误态污染已有数据。 */
data class TimelineState(
    val posts: List<PostUi> = emptyList(),
    val initialPhase: InitialPhase = InitialPhase.Idle,
    val initialError: AppError? = null,
    val appendPhase: AppendPhase = AppendPhase.Idle,
    val appendError: AppError? = null,
    val nextCursor: String? = null,
    val hasMore: Boolean = true,
)

/**
 * 信息流通用控制器 — 封装分页、刷新、点赞、分享逻辑。
 * 供 FeedViewModel 复用。
 */
class TimelineController(
    private val scope: CoroutineScope,
    private val repository: SocialRepository,
) {
    private val _state = MutableStateFlow(TimelineState())
    val state: StateFlow<TimelineState> = _state.asStateFlow()

    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    /** filter: Feed 传 tab.filter，Social 传空字符串。 */
    fun load(filter: String = "") {
        val current = _state.value
        if (current.initialPhase == InitialPhase.Loading) return
        _state.update {
            it.copy(
                initialPhase = InitialPhase.Loading,
                initialError = null,
                appendPhase = AppendPhase.Idle,
                appendError = null,
                nextCursor = null,
            )
        }
        fetchFirstPage(filter)
    }

    fun refresh(filter: String = "") {
        if (_state.value.initialPhase == InitialPhase.Loading) return
        _state.update {
            it.copy(
                initialPhase = InitialPhase.Loading,
                initialError = null,
                appendPhase = AppendPhase.Idle,
                appendError = null,
                nextCursor = null,
            )
        }
        fetchFirstPage(filter)
    }

    fun loadMore(filter: String = "") {
        val s = _state.value
        val cursor = s.nextCursor ?: return
        // 防御：首屏未完成 / 无更多数据 / 已在追加中 均不发起
        if (s.initialPhase != InitialPhase.Success && s.initialPhase != InitialPhase.Empty) return
        if (s.appendPhase == AppendPhase.Loading || !s.hasMore) return
        _state.update { it.copy(appendPhase = AppendPhase.Loading, appendError = null) }
        scope.launch {
            try {
                val (posts, next) = repository.getFeed(cursor, 20, filter)
                _state.update {
                    it.copy(
                        posts = it.posts + posts,
                        nextCursor = next,
                        hasMore = next != null,
                        appendPhase = AppendPhase.Idle,
                    )
                }
            } catch (e: Exception) {
                _state.update {
                    it.copy(
                        appendPhase = AppendPhase.Error,
                        appendError = e.toAppError(),
                    )
                }
            }
        }
    }

    fun retryLoadMore(filter: String = "") {
        loadMore(filter)
    }

    fun toggleLike(postId: String) {
        scope.launch {
            val post = _state.value.posts.find { it.postId == postId } ?: return@launch
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
                _uiEvent.emit(UiEvent.SnackbarRes(R.string.snackbar_action_rollback))
            }
        }
    }

    fun sharePost(postId: String) {
        scope.launch {
            val post = _state.value.posts.find { it.postId == postId } ?: return@launch
            updatePost(postId) { it.copy(shareCount = it.shareCount + 1) }
            try {
                repository.sharePost(postId)
            } catch (e: Exception) {
                updatePost(postId) { post }
                _uiEvent.emit(UiEvent.SnackbarRes(R.string.snackbar_share_failed))
            }
        }
    }

    private fun fetchFirstPage(filter: String) {
        scope.launch {
            try {
                val (posts, next) = repository.getFeed("", 20, filter)
                _state.update {
                    it.copy(
                        posts = posts,
                        nextCursor = next,
                        hasMore = next != null,
                        initialPhase = if (posts.isEmpty()) InitialPhase.Empty else InitialPhase.Success,
                    )
                }
            } catch (e: Exception) {
                _state.update {
                    it.copy(
                        initialPhase = InitialPhase.Error,
                        initialError = e.toAppError(),
                    )
                }
            }
        }
    }

    private fun updatePost(postId: String, transform: (PostUi) -> PostUi) {
        _state.update { s -> s.copy(posts = s.posts.map { if (it.postId == postId) transform(it) else it }) }
    }
}
