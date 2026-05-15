package com.antclaw.alfq.ui.notification

import android.app.Application
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.notification.*
import com.antclaw.alfq.data.sse.SseManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class NotificationViewModel @Inject constructor(
    private val sseManager: SseManager,
    private val notifRepo: NotificationRepository,
    application: Application,
) : androidx.lifecycle.AndroidViewModel(application) {

    private val systemNotifier = AndroidNotificationHelper(application)

    private val _state = MutableStateFlow(NotificationUiState())
    val state: StateFlow<NotificationUiState> = _state.asStateFlow()

    private val _isForeground = MutableStateFlow(true)
    val isForeground: StateFlow<Boolean> = _isForeground.asStateFlow()

    private val _prefs = MutableStateFlow(AlertPrefs())
    val prefs: StateFlow<AlertPrefs> = _prefs.asStateFlow()

    private val _prefsLoading = MutableStateFlow(false)
    val prefsLoading: StateFlow<Boolean> = _prefsLoading.asStateFlow()

    init {
        loadInitial()
        // SSE 连接由 SessionViewModel 统一管理，此处仅被动订阅事件
        viewModelScope.launch {
            sseManager.notifications.collect { notif ->
                _state.update {
                    it.copy(
                        unreadCount = it.unreadCount + 1,
                        items = listOf(notif) + it.items,
                    )
                }
                if (!_isForeground.value) {
                    systemNotifier.showNotificationForAppInBackground(notif)
                }
            }
        }
        viewModelScope.launch {
            sseManager.connected.collect { connected ->
                _state.update { it.copy(connected = connected, error = if (!connected) null else it.error) }
            }
        }
    }

    fun loadInitial() {
        viewModelScope.launch {
            try {
                val count = notifRepo.unreadCount()
                val items = notifRepo.listUnread()
                _state.update { it.copy(unreadCount = count.toInt(), items = items) }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
    }

    /** 前后台切换通知，由 Activity/Fragment 转发给 SessionViewModel。 */
    fun setForeground(fg: Boolean) {
        _isForeground.value = fg
        if (fg) loadInitial()
    }

    fun markRead(id: String) {
        viewModelScope.launch {
            try {
                notifRepo.markRead(id)
                _state.update {
                    val newItems = it.items.map { notif ->
                        if (notif.id == id) notif.copy(isRead = true, readAt = java.time.Instant.now()) else notif
                    }
                    it.copy(items = newItems, unreadCount = (it.unreadCount - 1).coerceAtLeast(0))
                }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
        }
    }

    fun markAllRead() {
        viewModelScope.launch {
            try {
                notifRepo.markAllRead()
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

    fun refresh() { loadInitial() }

    // ===== 偏好设置 =====

    fun loadPrefs() {
        viewModelScope.launch {
            _prefsLoading.value = true
            try {
                _prefs.value = notifRepo.getAlertPrefs()
            } catch (_: Exception) {
            } finally {
                _prefsLoading.value = false
            }
        }
    }

    fun updatePrefs(prefs: AlertPrefs) {
        viewModelScope.launch {
            try {
                notifRepo.updateAlertPrefs(prefs)
                _prefs.value = prefs
            } catch (_: Exception) {}
        }
    }

    override fun onCleared() {
        super.onCleared()
        systemNotifier.cancelAll()
        // SSE disconnect 由 SessionViewModel 管理，此处不再调用
    }
}
