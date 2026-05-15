package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.ProfileRepository
import com.antclaw.alfq.ui.social.UiEvent
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ProfileUiState(
    val displayName: String = "", val bio: String = "", val tier: String = "normal",
    val winRate: Double = 0.0, val profitFactor: Double = 0.0, val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0, val followerCount: Int = 0, val followingCount: Int = 0,
    val isFollowing: Boolean = false, val loading: Boolean = false,
)

@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val profileRepo: ProfileRepository,
) : ViewModel() {
    private var currentUserId = ""

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()
    private val _uiEvent = MutableSharedFlow<UiEvent>()
    val uiEvent: SharedFlow<UiEvent> = _uiEvent.asSharedFlow()

    fun load(userId: String) {
        currentUserId = userId
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true)
            try {
                val p = profileRepo.getProfile(userId)
                _uiState.value = ProfileUiState(
                    displayName = p.displayName, bio = p.bio, tier = p.tier,
                    winRate = p.winRate, profitFactor = p.profitFactor, sharpeRatio = p.sharpeRatio,
                    totalTrades = p.totalTrades, followerCount = p.followerCount,
                    followingCount = p.followingCount, loading = false,
                )
            } catch (_: Exception) {
                _uiState.value = _uiState.value.copy(loading = false)
            }
        }
    }

    fun toggleFollow() {
        viewModelScope.launch {
            val previous = _uiState.value
            val willFollow = !previous.isFollowing
            _uiState.value = previous.copy(
                isFollowing = willFollow,
                followerCount = if (willFollow) previous.followerCount + 1 else (previous.followerCount - 1).coerceAtLeast(0),
            )
            try {
                val fc = if (previous.isFollowing) profileRepo.unfollow(currentUserId) else profileRepo.follow(currentUserId)
                _uiState.value = _uiState.value.copy(isFollowing = !previous.isFollowing, followerCount = fc)
            } catch (e: Exception) {
                _uiState.value = previous
                _uiEvent.emit(UiEvent.Snackbar(e.message ?: "关注操作失败，已回滚"))
            }
        }
    }
}
