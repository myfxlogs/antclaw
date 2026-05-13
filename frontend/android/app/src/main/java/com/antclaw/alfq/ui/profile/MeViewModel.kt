package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqTrader
import antclaw.v1.UserOuterClass
import com.antclaw.alfq.data.repository.AuthRepository
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MeUiState(
    val userId: String = "",
    val displayName: String = "",
    val username: String = "",
    val bio: String = "",
    val tier: String = "normal",
    val followerCount: Int = 0,
    val followingCount: Int = 0,
    val winRate: Double = 0.0,
    val profitFactor: Double = 0.0,
    val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0,
    val loggingOut: Boolean = false,
    val loading: Boolean = true,
)

@HiltViewModel
class MeViewModel @Inject constructor(
    private val authRepo: AuthRepository
) : ViewModel() {

    private val _state = MutableStateFlow(MeUiState())
    val state: StateFlow<MeUiState> = _state.asStateFlow()

    private fun client() = ConnectTransportProvider.createProtocolClient()

    init { loadProfile() }

    fun loadProfile() {
        viewModelScope.launch {
            _state.value = _state.value.copy(loading = true)
            try {
                // 1. UserService/GetMe for basic identity
                val meSpec = MethodSpec("antclaw.v1.UserService/GetMe",
                    UserOuterClass.GetMeRequest::class, UserOuterClass.GetMeResponse::class, StreamType.UNARY)
                val meResp: ResponseMessage<UserOuterClass.GetMeResponse> =
                    client().unary(UserOuterClass.GetMeRequest.getDefaultInstance(), emptyMap(), meSpec)
                val u = meResp.getOrThrow().user
                val userId = u.userId

                // 2. TraderService/GetProfile for trading stats
                var displayName = u.displayName.ifEmpty { u.username }
                var followerCount = 0
                var followingCount = 0
                var winRate = 0.0
                var profitFactor = 0.0
                var sharpeRatio = 0.0
                var totalTrades = 0
                var tier = "normal"
                var bio = ""

                try {
                    val tpReq = AlfqTrader.GetTraderProfileRequest.newBuilder().setUserId(userId).build()
                    val tpSpec = MethodSpec("antclaw.v1.TraderService/GetProfile",
                        AlfqTrader.GetTraderProfileRequest::class, AlfqTrader.TraderProfile::class, StreamType.UNARY)
                    val tpResp: ResponseMessage<AlfqTrader.TraderProfile> = client().unary(tpReq, emptyMap(), tpSpec)
                    val p = tpResp.getOrThrow()
                    displayName = p.displayName.ifEmpty { displayName }
                    bio = p.bio
                    tier = p.tier
                    followerCount = p.followerCount
                    followingCount = p.followingCount
                    winRate = p.winRate
                    profitFactor = p.profitFactor
                    sharpeRatio = p.sharpeRatio
                    totalTrades = p.totalTrades
                } catch (_: Exception) {
                    // GetProfile 失败时用 GetMe 回退数据
                }

                _state.value = MeUiState(
                    userId = userId,
                    displayName = displayName,
                    username = "@${u.username}",
                    bio = bio,
                    tier = tier,
                    followerCount = followerCount,
                    followingCount = followingCount,
                    winRate = winRate,
                    profitFactor = profitFactor,
                    sharpeRatio = sharpeRatio,
                    totalTrades = totalTrades,
                    loading = false,
                )
            } catch (_: Exception) {
                _state.value = _state.value.copy(loading = false)
            }
        }
    }

    fun logout(onDone: () -> Unit) {
        viewModelScope.launch {
            _state.value = _state.value.copy(loggingOut = true)
            authRepo.logout()
            _state.value = _state.value.copy(loggingOut = false)
            onDone()
        }
    }
}
