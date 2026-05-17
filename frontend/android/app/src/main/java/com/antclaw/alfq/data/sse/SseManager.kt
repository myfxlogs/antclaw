package com.antclaw.alfq.data.sse

import android.content.Context
import com.antclaw.alfq.data.notification.ClientNotification
import com.antclaw.alfq.data.rpc.ProtocolClientFactory
import com.antclaw.alfq.data.rpc.TokenManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.GlobalScope
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.launch
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
 * SSE 管理器 facade — 组合 NetworkMonitor + SseConnectionController + NotificationEventParser。
 *
 * 对外提供简洁的 SseClient 接口 + 通知/连接状态流。
 * 内部委托给拆分后的子组件。
 */
@Singleton
class SseManager @Inject constructor(
    @ApplicationContext appContext: Context,
    tokenManager: TokenManager,
    clientFactory: ProtocolClientFactory,
) : SseClient {
    private val networkMonitor: NetworkMonitor
    private val eventParser = NotificationEventParser()
    private val connectionController: SseConnectionController

    init {
        networkMonitor = NetworkMonitor(appContext)
        connectionController = SseConnectionController(
            tokenManager, clientFactory, networkMonitor, eventParser
        )
        networkMonitor.start()
    }

    // ── SseClient ──
    override fun connect() = connectionController.connect()
    override fun disconnect() = connectionController.disconnect()
    override fun reconnect() = connectionController.reconnect()
    override fun destroy() {
        connectionController.destroy()
        networkMonitor.stop()
    }

    // ── Exposed flows ──

    /** 原始通知 JSON 流（向后兼容） */
    val notificationEvents: SharedFlow<String> = eventParser.notificationEvents

    /** 已解析的 ClientNotification 流 */
    val notifications: SharedFlow<ClientNotification> = eventParser.notifications

    /** 连接状态 */
    val connectionState: SharedFlow<SseConnectionController.ConnectionState> =
        connectionController.connectionState

    /** 兼容旧 connected 属性名 */
    val connected: SharedFlow<Boolean> = connectionController.connectionState.let { flow ->
        val out = MutableSharedFlow<Boolean>(replay = 1, extraBufferCapacity = 4)
        GlobalScope.launch {
            flow.collect { state -> out.emit(state == SseConnectionController.ConnectionState.CONNECTED) }
        }
        out
    }
}
