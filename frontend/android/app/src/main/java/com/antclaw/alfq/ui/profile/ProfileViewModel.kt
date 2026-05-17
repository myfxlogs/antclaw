package com.antclaw.alfq.ui.profile

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.R
import com.antclaw.alfq.data.repository.ProfileRepository
import com.antclaw.alfq.data.repository.SocialRepository
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.UiEvent
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class ProfileTab { POSTS, MEDIA, LIKES }

data class ProfileUiState(
    val displayName: String = "", val username: String = "", val codeId: String = "",
    val bio: String = "", val tier: String = "normal",
    val winRate: Double = 0.0, val profitFactor: Double = 0.0, val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0, val followerCount: Int = 0, val followingCount: Int = 0,
    val isFollowing: Boolean = false,
    val loading: Boolean = false, val error: String? = null,
    val isFollowLoading: Boolean = false,
    val currentTab: ProfileTab = ProfileTab.POSTS,
    // Posts tab
    val posts: List<PostUi> = emptyList(),
    val postsLoading: Boolean = false,
    val postsError: String? = null,
)

@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val profileRepo: ProfileRepository,
    private val socialRepo: SocialRepository,
    @ApplicationContext private val appContext: Context,
) : ViewModel() {
    private var currentUserId = ""

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()
    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    fun load(userId: String) {
        currentUserId = userId
        _uiState.update { it.copy(loading = true, error = null) }
        viewModelScope.launch {
            try {
                val p = profileRepo.getProfile(userId)
                _uiState.update {
                    it.copy(
                        displayName = p.displayName, bio = p.bio, tier = p.tier,
                        winRate = p.winRate, profitFactor = p.profitFactor, sharpeRatio = p.sharpeRatio,
                        totalTrades = p.totalTrades, followerCount = p.followerCount,
                        followingCount = p.followingCount, loading = false,
                    )
                }
            } catch (e: Exception) {
                _uiState.update { it.copy(loading = false, error = e.message ?: appContext.getString(R.string.common_error)) }
            }
        }
        loadPosts()
    }

    fun loadPosts() {
        if (_uiState.value.postsLoading) return
        _uiState.update { it.copy(postsLoading = true, postsError = null) }
        viewModelScope.launch {
            try {
                val (posts, _) = socialRepo.listUserPosts(currentUserId)
                _uiState.update { it.copy(posts = posts, postsLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(postsLoading = false, postsError = e.message ?: appContext.getString(R.string.snackbar_posts_load_failed)) }
            }
        }
    }

    fun selectTab(tab: ProfileTab) {
        _uiState.update { it.copy(currentTab = tab) }
    }

    fun toggleFollow() {
        viewModelScope.launch {
            val previous = _uiState.value
            val willFollow = !previous.isFollowing
            _uiState.update {
                it.copy(
                    isFollowing = willFollow, isFollowLoading = true,
                    followerCount = if (willFollow) it.followerCount + 1 else (it.followerCount - 1).coerceAtLeast(0),
                )
            }
            try {
                val fc = if (previous.isFollowing) profileRepo.unfollow(currentUserId)
                else profileRepo.follow(currentUserId)
                _uiState.update { it.copy(isFollowing = !previous.isFollowing, followerCount = fc, isFollowLoading = false) }
            } catch (e: Exception) {
                _uiState.update { it.copy(isFollowing = previous.isFollowing, followerCount = previous.followerCount, isFollowLoading = false) }
                _uiEvent.emit(if (e.message != null) UiEvent.Snackbar(e.message!!) else UiEvent.SnackbarRes(R.string.snackbar_follow_failed))
            }
        }
    }
}
