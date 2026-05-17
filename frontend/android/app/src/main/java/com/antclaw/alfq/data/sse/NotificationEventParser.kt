package com.antclaw.alfq.data.sse

import android.util.Log
import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.notification.ClientNotification
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

/**
 * SSE 通知事件解析器 — 从 SseManager 分离。
 *
 * 职责：
 * 1. 接收原始 SSE event data（base64）
 * 2. 解析 Protobuf Notification
 * 3. 转换为 ClientNotification 并 emit 到 SharedFlow
 *
 * 纯数据转换，无网络/连接逻辑。
 */
@Singleton
class NotificationEventParser @Inject constructor() {
    companion object {
        private const val TAG = "NotificationParser"
    }

    private val _notificationEvents = MutableSharedFlow<String>(extraBufferCapacity = 64)
    val notificationEvents: SharedFlow<String> = _notificationEvents

    private val _notifications = MutableSharedFlow<ClientNotification>(extraBufferCapacity = 32)
    val notifications: SharedFlow<ClientNotification> = _notifications

    /**
     * 解析并 emit 一条 SSE 通知。
     * 调用方需传入自己的 CoroutineScope。
     */
    fun parseAndEmit(data: String, scope: CoroutineScope) {
        scope.launch {
            _notificationEvents.emit(data)
            try {
                val bytes = android.util.Base64.decode(data, android.util.Base64.DEFAULT)
                val proto = NotificationOuterClass.Notification.parseFrom(bytes)
                _notifications.emit(ClientNotification.fromProto(proto))
            } catch (e: Exception) {
                Log.e(TAG, "Parse SSE notification failed", e)
            }
        }
    }
}
