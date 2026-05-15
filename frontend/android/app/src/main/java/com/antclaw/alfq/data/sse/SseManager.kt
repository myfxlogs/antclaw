package com.antclaw.alfq.data.sse

import android.util.Log
import com.antclaw.alfq.data.notification.ClientNotification
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources
import org.json.JSONObject
import java.time.Instant
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

/** SSE 客户端接口 — 允许测试注入 fake 实现。 */
interface SseClient {
    fun connect()
    fun disconnect()
    fun reconnect()
    fun destroy()
}

/**
 * 统一 SSE 管理器
 *
 * 职责：
 * 1. 维持 /sse/notifications 长连接（在线状态上报）
 * 2. 解析 event: notification → 推送 ClientNotification
 * 3. 暴露连接状态 + 通知流
 *
 * 合并前身：SseManager + NotificationSseClient（消除双 SSE 连接）
 */
@Singleton
class SseManager @Inject constructor() : SseClient {

    companion object {
        private const val TAG = "SseManager"
        private const val BASE_DELAY_MS = 3000L
        private const val MAX_DELAY_MS = 30000L
        private const val CONNECT_TIMEOUT_SEC = 30L

        fun backoffDelay(retry: Int, baseMs: Long = BASE_DELAY_MS, maxMs: Long = MAX_DELAY_MS): Long =
            (baseMs * (1L shl (retry - 1))).coerceAtMost(maxMs)
    }

    private var eventSource: EventSource? = null
    private var reconnectJob: Job? = null
    private var retryCount = 0
    private var lastConnectAttempt = 0L // anti-jitter
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    // 原始通知 JSON 流（向后兼容）
    private val _notificationEvents = MutableSharedFlow<String>(extraBufferCapacity = 64)
    val notificationEvents: SharedFlow<String> = _notificationEvents

    // 已解析的 ClientNotification 流（供 NotificationViewModel 使用）
    private val _notifications = MutableSharedFlow<ClientNotification>(extraBufferCapacity = 32)
    val notifications: SharedFlow<ClientNotification> = _notifications

    // 连接状态
    private val _connectionState = MutableSharedFlow<ConnectionState>(replay = 1, extraBufferCapacity = 16)
    val connectionState: SharedFlow<ConnectionState> = _connectionState

    /** 兼容 NotificationSseClient 的 connected 属性名 */
    val connected: SharedFlow<Boolean> = _connectionState.let { flow ->
        val out = MutableSharedFlow<Boolean>(replay = 1, extraBufferCapacity = 4)
        scope.launch {
            flow.collect { state -> out.emit(state == ConnectionState.CONNECTED) }
        }
        out
    }

    enum class ConnectionState {
        CONNECTING, CONNECTED, DISCONNECTED, ERROR
    }

    /** 登录成功后调用，建立 SSE 长连接。线程安全，可在主线程调用。 */
    override fun connect() {
        scope.launch {
            try {
                connectInternal()
            } catch (e: Exception) {
                Log.e(TAG, "SSE connect failed", e)
                _connectionState.emit(ConnectionState.ERROR)
                scheduleReconnect()
            }
        }
    }

    private suspend fun connectInternal() {
        // Anti-jitter: skip if called within 500ms of last attempt
        val now = System.currentTimeMillis()
        if (now - lastConnectAttempt < 500) {
            Log.d(TAG, "connect: skipping (anti-jitter, ${now - lastConnectAttempt}ms since last)")
            return
        }
        lastConnectAttempt = now

        val token = ConnectTransportProvider.getToken()
        if (token.isNullOrEmpty()) {
            Log.w(TAG, "connect: no token available, skipping SSE")
            return
        }
        reconnectJob?.cancel()
        disconnect()

        _connectionState.emit(ConnectionState.CONNECTING)

        val baseUrl = ConnectTransportProvider.baseUrl.trimEnd('/')
        val sseUrl = "$baseUrl/sse/notifications"

        val client = OkHttpClient.Builder()
            .connectTimeout(CONNECT_TIMEOUT_SEC, TimeUnit.SECONDS)
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()

        val request = Request.Builder()
            .url(sseUrl)
            .header("Authorization", "Bearer $token")
            .header("Accept", "text/event-stream")
            .build()

        val factory = EventSources.createFactory(client)
        eventSource = factory.newEventSource(request, object : EventSourceListener() {
            override fun onOpen(eventSource: EventSource, response: Response) {
                Log.i(TAG, "SSE connected: ${response.code}")
                retryCount = 0
                scope.launch { _connectionState.emit(ConnectionState.CONNECTED) }
            }

            override fun onEvent(eventSource: EventSource, id: String?, type: String?, data: String) {
                when (type) {
                    "notification" -> {
                        scope.launch {
                            _notificationEvents.emit(data)
                            parseAndEmit(data)
                        }
                    }
                    else -> Log.d(TAG, "SSE event type=$type data=$data")
                }
            }

            override fun onClosed(eventSource: EventSource) {
                Log.i(TAG, "SSE closed, scheduling reconnect")
                scope.launch { _connectionState.emit(ConnectionState.DISCONNECTED) }
                scheduleReconnect()
            }

            override fun onFailure(eventSource: EventSource, t: Throwable?, response: Response?) {
                Log.e(TAG, "SSE failure: ${t?.message} code=${response?.code}", t)
                scope.launch { _connectionState.emit(ConnectionState.ERROR) }
                scheduleReconnect()
            }
        })
    }

    /** 登出/前台→后台切换时调用。 */
    override fun disconnect() {
        reconnectJob?.cancel()
        reconnectJob = null
        eventSource?.cancel()
        eventSource = null
        scope.launch { _connectionState.emit(ConnectionState.DISCONNECTED) }
    }

    /** 断开后立即重连（前后台切换用）。 */
    override fun reconnect() {
        disconnect()
        connect()
    }

    private fun scheduleReconnect() {
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            retryCount++
            val delayMs = backoffDelay(retryCount)
            Log.i(TAG, "SSE reconnecting in ${delayMs}ms (retry #$retryCount)")
            delay(delayMs)
            connect()
        }
    }

    /** 生命周期清理。 */
    override fun destroy() {
        disconnect()
        scope.cancel()
    }

    // ── JSON 解析 ──

    private suspend fun parseAndEmit(data: String) {
        try {
            val json = JSONObject(data)
            val notif = ClientNotification(
                id = json.optString("id", ""),
                category = json.optString("category", "system"),
                type = json.optString("type", "in_app"),
                title = json.optString("title", ""),
                body = json.optString("body", ""),
                severity = json.optString("severity", "normal"),
                data = emptyMap(),
                createdAt = Instant.parse(
                    json.optString("created_at", Instant.now().toString())
                ),
            )
            _notifications.emit(notif)
        } catch (e: Exception) {
            Log.e(TAG, "Parse SSE notification failed", e)
        }
    }
}
