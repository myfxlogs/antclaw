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

/**
 * 会话状态 — 集中管理登录态、Token、用户身份、SSE 连接。
 *
 * 文档要求：
 * - access token 注入所有 Connect-RPC 请求
 * - refresh token 加密存储
 * - 401 refresh 串行化，失败后通知 UI 跳登录
 * - 当前用户 ID 在登录/注册成功后保存
 * - SSE 连接由会话统一管理，支持幂等 connect 和指数退避
 */
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

@HiltViewModel
class SessionViewModel @Inject constructor(
    private val tokenStore: TokenStore,
    private val sseManager: SseClient,
) : ViewModel() {

    private val _session = MutableStateFlow(SessionInfo())
    val session: StateFlow<SessionInfo> = _session.asStateFlow()

    private val _events = MutableSharedFlow<SessionEvent>()
    val events: SharedFlow<SessionEvent> = _events.asSharedFlow()

    init {
        viewModelScope.launch {
            val token = tokenStore.getAccessToken()
            val userId = tokenStore.getUserId().orEmpty()
            if (token != null && userId.isNotBlank()) {
                _session.value = SessionInfo(
                    state = SessionState.AUTHENTICATED,
                    userId = userId,
                )
                sseManager.connect()
            } else {
                _session.value = SessionInfo(state = SessionState.UNAUTHENTICATED)
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
            _session.value = SessionInfo(
                state = SessionState.AUTHENTICATED,
                userId = userId,
                displayName = displayName,
            )
            sseManager.connect()
        }
    }

    fun onSessionExpired() {
        viewModelScope.launch {
            tokenStore.clearTokens()
            ConnectTransportProvider.clearToken()
            sseManager.disconnect()
            _session.value = SessionInfo(state = SessionState.EXPIRED)
            _events.emit(SessionEvent.RequireLogin)
        }
    }

    fun logout() {
        viewModelScope.launch {
            tokenStore.clearTokens()
            tokenStore.clearUserId()
            ConnectTransportProvider.clearToken()
            sseManager.disconnect()
            _session.value = SessionInfo(state = SessionState.UNAUTHENTICATED)
            _events.emit(SessionEvent.LoggedOut)
        }
    }

    /** App 回前台时调用：重连 SSE 并补拉未读数 */
    fun onForeground() {
        if (_session.value.state == SessionState.AUTHENTICATED) {
            sseManager.reconnect()
        }
    }

    /** App 进后台时调用：断开 SSE 以节省资源 */
    fun onBackground() {
        sseManager.disconnect()
    }

    fun isAuthenticated(): Boolean = _session.value.state == SessionState.AUTHENTICATED
}
