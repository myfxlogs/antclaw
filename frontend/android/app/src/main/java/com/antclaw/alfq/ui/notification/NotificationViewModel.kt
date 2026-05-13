package com.antclaw.alfq.ui.notification

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.notification.*
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

// ViewModel: 管理通知列表、未读数、SSE 连接、偏好。
// 支持前后台生命周期：后台断开 SSE，前台先拉未读再重连。
class NotificationViewModel(application: Application) : AndroidViewModel(application) {

    private val repo = NotificationRepository()
    private val sseClient = NotificationSseClient()
    private val systemNotifier = AndroidNotificationHelper(application)

    private val _state = MutableStateFlow(NotificationUiState())
    val state: StateFlow<NotificationUiState> = _state.asStateFlow()

    // 业务偏好
    private val _alertPrefs = MutableStateFlow(AlertPrefs())
    val alertPrefs: StateFlow<AlertPrefs> = _alertPrefs.asStateFlow()

    // 前后台状态
    private val _isForeground = MutableStateFlow(true)
    val isForeground: StateFlow<Boolean> = _isForeground.asStateFlow()

    init {
        loadInitial()
        connectSse()
    }

    fun loadInitial() {
        viewModelScope.launch {
            try {
                val count = repo.unreadCount()
                val items = repo.listUnread()
                _state.update {
                    it.copy(unreadCount = count.toInt(), items = items)
                }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
        viewModelScope.launch {
            try {
                _alertPrefs.value = repo.getAlertPrefs()
            } catch (_: Exception) { }
        }
    }

    private fun connectSse() {
        viewModelScope.launch {
            sseClient.notifications.collect { notif ->
                _state.update {
                    it.copy(
                        unreadCount = it.unreadCount + 1,
                        items = listOf(notif) + it.items,
                    )
                }
                // 后台时发送系统通知
                if (!_isForeground.value) {
                    systemNotifier.showNotificationForAppInBackground(notif)
                }
            }
        }
        viewModelScope.launch {
            sseClient.connected.collect { connected ->
                _state.update { it.copy(connected = connected, error = if (!connected) null else it.error) }
            }
        }
        sseClient.connect()
    }

    /** 前台恢复时调用：先拉未读列表，再重连 SSE */
    fun onForeground() {
        _isForeground.value = true
        loadInitial()
        sseClient.reconnect()
    }

    /** 后台进入时调用：断开 SSE，节省电量和连接 */
    fun onBackground() {
        _isForeground.value = false
        sseClient.disconnect()
    }

    fun markRead(id: String) {
        viewModelScope.launch {
            try {
                repo.markRead(id)
                _state.update {
                    val newItems = it.items.map { notif ->
                        if (notif.id == id) notif.copy(isRead = true, readAt = java.time.Instant.now()) else notif
                    }
                    it.copy(
                        items = newItems,
                        unreadCount = (it.unreadCount - 1).coerceAtLeast(0),
                    )
                }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
    }

    fun markAllRead() {
        viewModelScope.launch {
            try {
                repo.markAllRead()
                _state.update {
                    val now = java.time.Instant.now()
                    it.copy(
                        items = it.items.map { n -> n.copy(isRead = true, readAt = now) },
                        unreadCount = 0,
                    )
                }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
    }

    fun updateAlertPrefs(prefs: AlertPrefs) {
        viewModelScope.launch {
            try {
                repo.updateAlertPrefs(prefs)
                _alertPrefs.value = prefs
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
    }

    fun refresh() {
        loadInitial()
    }

    override fun onCleared() {
        super.onCleared()
        sseClient.disconnect()
        systemNotifier.cancelAll()
    }
}
