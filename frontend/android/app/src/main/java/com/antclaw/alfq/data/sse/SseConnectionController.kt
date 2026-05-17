package com.antclaw.alfq.data.sse

import android.util.Log
import com.antclaw.alfq.data.rpc.ProtocolClientFactory
import com.antclaw.alfq.data.rpc.TokenManager
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.sse.EventSource
import okhttp3.sse.EventSourceListener
import okhttp3.sse.EventSources
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

/**
 * SSE 连接控制器 — 从 SseManager 分离。
 *
 * 职责：
 * 1. SSE 长连接建立/断开/重连
 * 2. 指数退避重连（backoffDelay）
 * 3. Anti-jitter 保护
 * 4. 连接状态暴露（connectionState）
 *
 * 不涉及：网络监听（委托 NetworkMonitor）、protobuf 解析（委托 NotificationEventParser）。
 */
@Singleton
class SseConnectionController @Inject constructor(
    private val tokenManager: TokenManager,
    private val clientFactory: ProtocolClientFactory,
    private val networkMonitor: NetworkMonitor,
    private val eventParser: NotificationEventParser,
) {
    companion object {
        private const val TAG = "SseConnCtrl"
        private const val BASE_DELAY_MS = 3000L
        private const val MAX_DELAY_MS = 30000L
        private const val CONNECT_TIMEOUT_SEC = 30L

        fun backoffDelay(retry: Int, baseMs: Long = BASE_DELAY_MS, maxMs: Long = MAX_DELAY_MS): Long =
            (baseMs * (1L shl (retry - 1))).coerceAtMost(maxMs)
    }

    enum class ConnectionState { CONNECTING, CONNECTED, DISCONNECTED, ERROR }

    private val sharedHttpClient: OkHttpClient by lazy {
        OkHttpClient.Builder()
            .connectTimeout(CONNECT_TIMEOUT_SEC, TimeUnit.SECONDS)
            .readTimeout(0, TimeUnit.MILLISECONDS)
            .build()
    }

    private var eventSource: EventSource? = null
    private var reconnectJob: Job? = null
    private var retryCount = 0
    private var lastConnectAttempt = 0L
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private val _connectionState = MutableSharedFlow<ConnectionState>(replay = 1, extraBufferCapacity = 16)
    val connectionState: SharedFlow<ConnectionState> = _connectionState

    // ── Public API ──

    fun connect() {
        scope.launch {
            try { connectInternal() }
            catch (e: Exception) {
                Log.e(TAG, "SSE connect failed", e)
                _connectionState.emit(ConnectionState.ERROR)
                scheduleReconnect()
            }
        }
    }

    fun disconnect() {
        reconnectJob?.cancel()
        reconnectJob = null
        eventSource?.cancel()
        eventSource = null
        scope.launch { _connectionState.emit(ConnectionState.DISCONNECTED) }
    }

    fun reconnect() {
        disconnect()
        connect()
    }

    fun destroy() {
        disconnect()
        scope.cancel()
    }

    // ── Internal ──

    private suspend fun connectInternal() {
        val now = System.currentTimeMillis()
        if (now - lastConnectAttempt < 500) {
            Log.d(TAG, "connect: skipping (anti-jitter)")
            return
        }
        if (!networkMonitor.isAvailable.value) {
            Log.d(TAG, "connect: skipping (network unavailable)")
            return
        }
        lastConnectAttempt = now

        val token = tokenManager.getToken()
        if (token.isNullOrEmpty()) {
            Log.w(TAG, "connect: no token available, skipping SSE")
            return
        }
        reconnectJob?.cancel()
        disconnect()

        _connectionState.emit(ConnectionState.CONNECTING)

        val baseUrl = clientFactory.baseUrl.trimEnd('/')
        val sseUrl = "$baseUrl/sse/notifications"

        val request = Request.Builder()
            .url(sseUrl)
            .header("Authorization", "Bearer $token")
            .header("Accept", "text/event-stream")
            .build()

        val factory = EventSources.createFactory(sharedHttpClient)
        eventSource = factory.newEventSource(request, object : EventSourceListener() {
            override fun onOpen(eventSource: EventSource, response: Response) {
                Log.i(TAG, "SSE connected: ${response.code}")
                retryCount = 0
                scope.launch { _connectionState.emit(ConnectionState.CONNECTED) }
            }

            override fun onEvent(eventSource: EventSource, id: String?, type: String?, data: String) {
                when (type) {
                    "notification" -> eventParser.parseAndEmit(data, scope)
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

    private fun scheduleReconnect() {
        if (!networkMonitor.isAvailable.value) return
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            retryCount++
            val delayMs = backoffDelay(retryCount)
            Log.i(TAG, "SSE reconnecting in ${delayMs}ms (retry #$retryCount)")
            delay(delayMs)
            connect()
        }
    }
}
