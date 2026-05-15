package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.AuthRepository
import com.antclaw.alfq.data.repository.ProfileRepository
import com.antclaw.alfq.data.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MeUiState(
    val userId: String = "", val displayName: String = "", val username: String = "",
    val bio: String = "", val tier: String = "normal",
    val followerCount: Int = 0, val followingCount: Int = 0,
    val winRate: Double = 0.0, val profitFactor: Double = 0.0, val sharpeRatio: Double = 0.0,
    val totalTrades: Int = 0, val loggingOut: Boolean = false, val loading: Boolean = true,
)

@HiltViewModel
class MeViewModel @Inject constructor(
    private val authRepo: AuthRepository,
    private val userRepo: UserRepository,
    private val profileRepo: ProfileRepository,
) : ViewModel() {
    private val _state = MutableStateFlow(MeUiState())
    val state: StateFlow<MeUiState> = _state.asStateFlow()

    init { loadProfile() }

    fun loadProfile() {
        viewModelScope.launch {
            _state.value = _state.value.copy(loading = true)
            try {
                val me = userRepo.getMe()
                var dp = me.displayName; var bio = ""; var tier = "normal"
                var fc = 0; var ing = 0; var wr = 0.0; var pf = 0.0; var sr = 0.0; var tt = 0

                try {
                    val p = profileRepo.getProfile(me.userId)
                    dp = p.displayName.ifEmpty { dp }; bio = p.bio; tier = p.tier
                    fc = p.followerCount; ing = p.followingCount
                    wr = p.winRate; pf = p.profitFactor; sr = p.sharpeRatio; tt = p.totalTrades
                } catch (_: Exception) {}

                _state.value = MeUiState(userId = me.userId, displayName = dp, username = "@${me.username}",
                    bio = bio, tier = tier, followerCount = fc, followingCount = ing,
                    winRate = wr, profitFactor = pf, sharpeRatio = sr, totalTrades = tt, loading = false)
            } catch (_: Exception) { _state.value = _state.value.copy(loading = false) }
        }
    }

    fun logout(onDone: () -> Unit) {
        viewModelScope.launch { _state.value = _state.value.copy(loggingOut = true); authRepo.logout(); _state.value = _state.value.copy(loggingOut = false); onDone() }
    }
}
