package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.TraderProfile
import com.antclaw.alfq.data.rpc.TraderRpcClient
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ProfileUiState(
    val displayName: String = "",
    val bio: String = "",
    val tier: String = "normal",
    val winRate: Double = 0.0,
    val profitFactor: Double = 0.0,
    val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0,
    val followerCount: Int = 0,
    val followingCount: Int = 0,
    val isFollowing: Boolean = false,
    val loading: Boolean = true,
)

@HiltViewModel
class ProfileViewModel @Inject constructor() : ViewModel() {
    private val client = TraderRpcClient()
    private var currentUserId = ""

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()

    fun load(userId: String) {
        currentUserId = userId
        viewModelScope.launch { updateState(client.getProfile(userId)) }
    }

    fun toggleFollow() {
        viewModelScope.launch {
            val resp = if (_uiState.value.isFollowing) client.unfollow(currentUserId)
            else { client.follow(currentUserId); return@launch client.follow(currentUserId) }
            _uiState.value = _uiState.value.copy(isFollowing = !_uiState.value.isFollowing, followerCount = resp.follower_count)
        }
    }

    private fun updateState(p: TraderProfile) {
        _uiState.value = ProfileUiState(
            displayName = p.display_name, bio = p.bio, tier = p.tier,
            winRate = p.win_rate, profitFactor = p.profit_factor,
            sharpeRatio = p.sharpe_ratio, totalTrades = p.total_trades,
            followerCount = p.follower_count, followingCount = p.following_count,
            loading = false
        )
    }
}
