package com.antclaw.alfq.ui.feed

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

/** 信息流加载阶段 */
enum class AsyncPhase { Idle, Loading, Refreshing, Appending }

/** 信息流统一状态 */
data class TimelineState(
    val posts: List<PostUi> = emptyList(),
    val phase: AsyncPhase = AsyncPhase.Idle,
    val error: String? = null,
    val appendError: String? = null,
    val nextCursor: String? = null,
    val hasMore: Boolean = true,
)

/**
 * 信息流通用控制器 — 封装分页、刷新、点赞、分享逻辑。
 * 供 FeedViewModel / SocialFeedViewModel 复用。
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
        _state.update { it.copy(phase = AsyncPhase.Loading, error = null, appendError = null, nextCursor = null) }
        fetchFirstPage(filter)
    }

    fun refresh(filter: String = "") {
        _state.update { it.copy(phase = AsyncPhase.Refreshing, error = null, appendError = null, nextCursor = null) }
        fetchFirstPage(filter)
    }

    fun loadMore(filter: String = "") {
        val s = _state.value
        val cursor = s.nextCursor ?: return
        if (s.phase == AsyncPhase.Loading || s.phase == AsyncPhase.Appending || !s.hasMore) return
        _state.update { it.copy(phase = AsyncPhase.Appending, appendError = null) }
        scope.launch {
            try {
                val (posts, next) = repository.getFeed(cursor, 20, filter)
                _state.update {
                    it.copy(
                        posts = it.posts + posts,
                        nextCursor = next,
                        hasMore = next != null,
                        phase = AsyncPhase.Idle,
                    )
                }
            } catch (e: Exception) {
                _state.update { it.copy(phase = AsyncPhase.Idle, appendError = e.message ?: "加载更多失败") }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载更多失败"))
            }
        }
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
                _uiEvent.emit(UiEvent.Snackbar("操作失败，已回滚"))
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
                _uiEvent.emit(UiEvent.Snackbar("分享失败"))
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
                        phase = AsyncPhase.Idle,
                    )
                }
            } catch (e: Exception) {
                _state.update { it.copy(phase = AsyncPhase.Idle, error = e.message ?: "加载失败") }
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "加载失败"))
            }
        }
    }

    private fun updatePost(postId: String, transform: (PostUi) -> PostUi) {
        _state.update { s -> s.copy(posts = s.posts.map { if (it.postId == postId) transform(it) else it }) }
    }
}
