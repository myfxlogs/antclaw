package com.antclaw.alfq.ui.notification

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.notification.*
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch

class NotificationViewModel(application: Application) : AndroidViewModel(application) {

    private val repo = NotificationRepository()
    private val sseClient = NotificationSseClient()
    private val systemNotifier = AndroidNotificationHelper(application)

    private val _state = MutableStateFlow(NotificationUiState())
    val state: StateFlow<NotificationUiState> = _state.asStateFlow()

    private val _isForeground = MutableStateFlow(true)
    val isForeground: StateFlow<Boolean> = _isForeground.asStateFlow()

    // 偏好设置状态
    private val _prefs = MutableStateFlow(AlertPrefs())
    val prefs: StateFlow<AlertPrefs> = _prefs.asStateFlow()

    private val _prefsLoading = MutableStateFlow(false)
    val prefsLoading: StateFlow<Boolean> = _prefsLoading.asStateFlow()

    init {
        loadInitial()
        connectSse()
    }

    fun loadInitial() {
        viewModelScope.launch {
            try {
                val count = repo.unreadCount()
                val items = repo.listUnread()
                _state.update { it.copy(unreadCount = count.toInt(), items = items) }
            } catch (e: Exception) {
                _state.update { it.copy(error = e.message) }
            }
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

    fun onForeground() {
        _isForeground.value = true
        loadInitial()
        sseClient.reconnect()
    }

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

    fun refresh() { loadInitial() }

    // ===== 偏好设置 =====

    fun loadPrefs() {
        viewModelScope.launch {
            _prefsLoading.value = true
            try {
                val client = ConnectTransportProvider.createProtocolClient()
                val spec = MethodSpec("antclaw.v1.NotificationService/GetAlertPrefs",
                    NotificationOuterClass.GetAlertPrefsRequest::class,
                    NotificationOuterClass.GetAlertPrefsResponse::class,
                    StreamType.UNARY)
                val resp = client.unary(
                    NotificationOuterClass.GetAlertPrefsRequest.getDefaultInstance(),
                    emptyMap(), spec
                ).getOrThrow()
                val p = resp.prefs
                _prefs.value = AlertPrefs(
                    currencies = p.currenciesList,
                    symbols = p.symbolsList,
                    impacts = p.impactsList,
                    reminderMinutes = p.reminderMinutesList,
                    highImpactOnly = p.highImpactOnly,
                    dailyDigestEnabled = p.dailyDigestEnabled,
                    weeklyDigestEnabled = p.weeklyDigestEnabled,
                    cotAlertsEnabled = p.cotAlertsEnabled,
                    macroAlertsEnabled = p.macroAlertsEnabled,
                    optionsAlertsEnabled = p.optionsAlertsEnabled,
                    onchainAlertsEnabled = p.onchainAlertsEnabled,
                )
            } catch (_: Exception) {
                // 保留默认值
            } finally {
                _prefsLoading.value = false
            }
        }
    }

    fun updatePrefs(prefs: AlertPrefs) {
        viewModelScope.launch {
            try {
                val client = ConnectTransportProvider.createProtocolClient()
                val proto = NotificationOuterClass.AlertPrefs.newBuilder()
                    .addAllCurrencies(prefs.currencies)
                    .addAllSymbols(prefs.symbols)
                    .addAllImpacts(prefs.impacts)
                    .addAllReminderMinutes(prefs.reminderMinutes)
                    .setHighImpactOnly(prefs.highImpactOnly)
                    .setDailyDigestEnabled(prefs.dailyDigestEnabled)
                    .setWeeklyDigestEnabled(prefs.weeklyDigestEnabled)
                    .setCotAlertsEnabled(prefs.cotAlertsEnabled)
                    .setMacroAlertsEnabled(prefs.macroAlertsEnabled)
                    .setOptionsAlertsEnabled(prefs.optionsAlertsEnabled)
                    .setOnchainAlertsEnabled(prefs.onchainAlertsEnabled)
                    .build()
                val req = NotificationOuterClass.UpdateAlertPrefsRequest.newBuilder().setPrefs(proto).build()
                val spec = MethodSpec("antclaw.v1.NotificationService/UpdateAlertPrefs",
                    NotificationOuterClass.UpdateAlertPrefsRequest::class,
                    NotificationOuterClass.UpdateAlertPrefsResponse::class,
                    StreamType.UNARY)
                client.unary(req, emptyMap(), spec).getOrThrow()
                _prefs.value = prefs
            } catch (_: Exception) {}
        }
    }

    override fun onCleared() {
        super.onCleared()
        sseClient.disconnect()
        systemNotifier.cancelAll()
    }
}
