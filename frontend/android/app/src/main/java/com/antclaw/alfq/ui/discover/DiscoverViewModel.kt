package com.antclaw.alfq.ui.discover

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.AlfqTrader
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

data class TraderItem(
    val userId: String,
    val displayName: String,
    val tier: String,
    val followerCount: Int,
)

data class CircleItem(
    val id: String,
    val name: String,
    val symbol: String,
    val memberCount: Int,
)

data class DiscoverUiState(
    val traders: List<TraderItem> = emptyList(),
    val circles: List<CircleItem> = emptyList(),
    val loading: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class DiscoverViewModel @Inject constructor() : ViewModel() {

    private val _uiState = MutableStateFlow(DiscoverUiState())
    val uiState: StateFlow<DiscoverUiState> = _uiState.asStateFlow()

    private fun client() = ConnectTransportProvider.createProtocolClient()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val traders = loadSuggestedTraders()
                _uiState.value = _uiState.value.copy(loading = false, traders = traders)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    loading = false,
                    error = e.message ?: "\u52a0\u8f7d\u5931\u8d25"
                )
            }
        }
    }

    private suspend fun loadSuggestedTraders(): List<TraderItem> {
        return try {
            val spec = MethodSpec("antclaw.v1.TraderService/GetFollowing",
                AlfqTrader.GetFollowingRequest::class,
                AlfqTrader.UserList::class,
                StreamType.UNARY)
            val req = AlfqTrader.GetFollowingRequest.newBuilder().setPageSize(20).build()
            val resp: ResponseMessage<AlfqTrader.UserList> = client().unary(req, emptyMap(), spec)
            resp.getOrThrow().usersList.map { u ->
                TraderItem(u.userId, u.displayName, u.tier, u.followerCount)
            }
        } catch (_: Exception) { emptyList() }
    }
}
