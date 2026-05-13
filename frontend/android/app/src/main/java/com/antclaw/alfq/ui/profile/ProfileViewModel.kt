package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqTrader
import com.antclaw.alfq.data.rpc.RpcHelper
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
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
class ProfileViewModel @Inject constructor() : ViewModel() {
    private var currentUserId = ""

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()

    fun load(userId: String) {
        currentUserId = userId
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true)
            try {
                val req = AlfqTrader.GetTraderProfileRequest.newBuilder().setUserId(userId).build()
                val p = RpcHelper.unary(
                    "antclaw.v1.TraderService/GetProfile", req,
                    AlfqTrader.GetTraderProfileRequest::class, AlfqTrader.TraderProfile::class)
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
            val isFollowing = _uiState.value.isFollowing
            try {
                if (isFollowing) {
                    val req = AlfqTrader.UnfollowRequest.newBuilder().setTargetUserId(currentUserId).build()
                    val resp = RpcHelper.unary(
                        "antclaw.v1.TraderService/Unfollow", req,
                        AlfqTrader.UnfollowRequest::class, AlfqTrader.FollowResponse::class)
                    _uiState.value = _uiState.value.copy(isFollowing = false, followerCount = resp.followerCount)
                } else {
                    val req = AlfqTrader.FollowRequest.newBuilder().setTargetUserId(currentUserId).build()
                    val resp = RpcHelper.unary(
                        "antclaw.v1.TraderService/Follow", req,
                        AlfqTrader.FollowRequest::class, AlfqTrader.FollowResponse::class)
                    _uiState.value = _uiState.value.copy(isFollowing = true, followerCount = resp.followerCount)
                }
            } catch (_: Exception) {}
        }
    }
}
