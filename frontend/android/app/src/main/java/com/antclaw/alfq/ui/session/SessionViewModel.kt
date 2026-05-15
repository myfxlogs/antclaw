package com.antclaw.alfq.ui.session

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.antclaw.alfq.data.sse.SseClient
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class SessionState { UNKNOWN, AUTHENTICATED, UNAUTHENTICATED, EXPIRED }

data class SessionInfo(
    val state: SessionState = SessionState.UNKNOWN,
    val userId: String = "",
    val displayName: String = "",
)

sealed class SessionEvent {
    data object RequireLogin : SessionEvent()
    data object LoggedOut : SessionEvent()
}

/** 集中管理登录态、Token、用户身份、SSE 连接。 */
@HiltViewModel
class SessionViewModel @Inject constructor(
    private val tokenStore: TokenStore,
    private val sseClient: SseClient,
) : ViewModel() {

    private val _session = MutableStateFlow(SessionInfo())
    val session: StateFlow<SessionInfo> = _session.asStateFlow()

    private val _events = MutableSharedFlow<SessionEvent>()
    val events: SharedFlow<SessionEvent> = _events.asSharedFlow()

    init {
        viewModelScope.launch {
            val (token, userId) = tokenStore.getAccessToken() to tokenStore.getUserId().orEmpty()
            if (token != null && userId.isNotBlank()) {
                setSession(SessionState.AUTHENTICATED, userId)
                sseClient.connect()
            } else {
                setSession(SessionState.UNAUTHENTICATED)
            }
        }
        ConnectTransportProvider.init(tokenStore)
    }

    fun onLoginSuccess(userId: String, accessToken: String, refreshToken: String, displayName: String = "") {
        viewModelScope.launch {
            tokenStore.saveAccessToken(accessToken)
            tokenStore.saveRefreshToken(refreshToken)
            tokenStore.saveUserId(userId)
            ConnectTransportProvider.setToken(accessToken)
            setSession(SessionState.AUTHENTICATED, userId, displayName)
            sseClient.connect()
        }
    }

    fun onSessionExpired() {
        viewModelScope.launch {
            clearSession()
            setSession(SessionState.EXPIRED)
            _events.emit(SessionEvent.RequireLogin)
        }
    }

    fun logout() {
        viewModelScope.launch {
            clearSession()
            tokenStore.clearUserId()
            setSession(SessionState.UNAUTHENTICATED)
            _events.emit(SessionEvent.LoggedOut)
        }
    }

    fun onForeground() {
        if (isAuthenticated()) sseClient.reconnect()
    }

    fun onBackground() {
        sseClient.disconnect()
    }

    fun isAuthenticated(): Boolean = _session.value.state == SessionState.AUTHENTICATED

    // ── internal ──

    private fun setSession(state: SessionState, userId: String = "", displayName: String = "") {
        _session.value = SessionInfo(state = state, userId = userId, displayName = displayName)
    }

    private suspend fun clearSession() {
        tokenStore.clearTokens()
        ConnectTransportProvider.clearToken()
        sseClient.disconnect()
    }
}
