package com.antclaw.alfq.data.notification

import android.util.Log
import com.antclaw.alfq.BuildConfig
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import kotlinx.coroutines.*
import kotlin.coroutines.coroutineContext
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.time.Instant
import java.util.concurrent.TimeUnit

// SSE 长连接客户端，订阅 /sse/notifications 端点，解析 event: notification → 推送 ClientNotification。
class NotificationSseClient {

    companion object {
        private const val TAG = "NotificationSseClient"
        private const val SSE_URL_PATH = "/sse/notifications"
        // 重连退避：1s → 2s → 5s → 15s → 30s，上限 30s
        private val BACKOFF_MS = longArrayOf(1000, 2000, 5000, 15000, 30000)
    }

    private var job: Job? = null
    private var scope: CoroutineScope? = null
    private var backoffIndex = 0

    // 收到的实时通知流
    private val _notifications = MutableSharedFlow<ClientNotification>(extraBufferCapacity = 32)
    val notifications: SharedFlow<ClientNotification> = _notifications

    // 连接状态
    private val _connected = MutableSharedFlow<Boolean>(extraBufferCapacity = 1)
    val connected: SharedFlow<Boolean> = _connected

    fun connect() {
        if (job?.isActive == true) return
        val token = ConnectTransportProvider.getToken()
        if (token == null) {
            Log.w(TAG, "No token, skip SSE connect")
            return
        }
        scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
        backoffIndex = 0
        job = scope?.launch {
            while (coroutineContext.isActive) {
                try {
                    connectBlocking(token)
                } catch (e: CancellationException) {
                    throw e
                } catch (e: Exception) {
                    Log.e(TAG, "SSE error", e)
                }
                // 指数退避
                val delay = if (backoffIndex < BACKOFF_MS.size) BACKOFF_MS[backoffIndex] else 30000L
                backoffIndex++
                Log.d(TAG, "Reconnect in ${delay}ms (attempt ${backoffIndex})")
                delay(delay)
            }
        }
    }

    fun disconnect() {
        job?.cancel()
        job = null
        scope?.cancel()
        scope = null
        _connected.tryEmit(false)
    }

    /** 断开后重新连接（用于前后台切换） */
    fun reconnect() {
        disconnect()
        connect()
    }

    private suspend fun connectBlocking(token: String) {
        val httpClient = OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(0, TimeUnit.MILLISECONDS) // 无限读超时
            .build()

        val url = "${BuildConfig.BASE_URL}$SSE_URL_PATH"
        val request = Request.Builder()
            .url(url)
            .header("Authorization", "Bearer $token")
            .header("Accept", "text/event-stream")
            .build()

        val response: Response = withContext(Dispatchers.IO) {
            httpClient.newCall(request).execute()
        }

        if (response.code == 401) {
            Log.w(TAG, "SSE 401, stop reconnecting")
            ConnectTransportProvider.clearToken()
            disconnect()
            return
        }

        if (!response.isSuccessful) {
            Log.w(TAG, "SSE connect failed: ${response.code}")
            return
        }

        _connected.tryEmit(true)
        backoffIndex = 0 // 重置退避

        try {
            val body = response.body ?: return
            val reader = BufferedReader(InputStreamReader(body.byteStream()))

            var eventType = ""
            val dataBuilder = StringBuilder()

            while (coroutineContext.isActive) {
                val line = reader.readLine()
                if (line == null) break
                when {
                    line.startsWith("event:") -> {
                        eventType = line.removePrefix("event:").trim()
                    }
                    line.startsWith("data:") -> {
                        dataBuilder.append(line.removePrefix("data:").trim())
                    }
                    line.isEmpty() -> {
                        // 空行表示消息结束
                        if (eventType == "notification" && dataBuilder.isNotEmpty()) {
                            try {
                                val json = JSONObject(dataBuilder.toString())
                                val notif = ClientNotification(
                                    id = json.optString("id", ""),
                                    category = json.optString("category", "system"),
                                    type = json.optString("type", "in_app"),
                                    title = json.optString("title", ""),
                                    body = json.optString("body", ""),
                                    severity = json.optString("severity", "normal"),
                                    data = emptyMap(), // SSE 消息不完整，需后续拉取
                                    createdAt = Instant.parse(json.optString("created_at", Instant.now().toString())),
                                )
                                _notifications.emit(notif)
                            } catch (e: Exception) {
                                Log.e(TAG, "Parse SSE notification failed", e)
                            }
                        }
                        eventType = ""
                        dataBuilder.clear()
                    }
                    // 忽略以 ":" 开头的注释行（心跳）
                }
            }
            reader.close()
        } finally {
            response.close()
            _connected.tryEmit(false)
        }
    }
}
