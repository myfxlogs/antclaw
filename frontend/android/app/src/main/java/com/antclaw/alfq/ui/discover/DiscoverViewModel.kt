package com.antclaw.alfq.ui.discover

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

data class TraderItem(val userId: String, val displayName: String, val tier: String, val followerCount: Int)
data class DiscoverUiState(val traders: List<TraderItem> = emptyList(), val loading: Boolean = false, val error: String? = null)

@HiltViewModel
class DiscoverViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(DiscoverUiState())
    val uiState: StateFlow<DiscoverUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val resp = RpcHelper.unary(
                    "antclaw.v1.TraderService/GetFollowing",
                    AlfqTrader.GetFollowingRequest.newBuilder().setPageSize(20).build(),
                    AlfqTrader.GetFollowingRequest::class, AlfqTrader.UserList::class)
                _uiState.value = DiscoverUiState(traders = resp.usersList.map {
                    TraderItem(it.userId, it.displayName, it.tier, it.followerCount)
                })
            } catch (e: Exception) { _uiState.value = _uiState.value.copy(loading = false, error = e.message) }
        }
    }
}
