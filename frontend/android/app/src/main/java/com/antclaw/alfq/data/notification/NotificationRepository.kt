package com.antclaw.alfq.data.notification

import antclaw.v1.Notification
import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import java.time.Instant

// 封装 NotificationService RPC 调用。
class NotificationRepository {

    fun createClient() = ConnectTransportProvider.createProtocolClient()

    // ListUnread
    suspend fun listUnread(limit: Int = 100): List<ClientNotification> {
        val client = createClient()
        val req = NotificationOuterClass.ListUnreadRequest.newBuilder()
            .setLimit(limit)
            .build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/ListUnread",
            requestClass = NotificationOuterClass.ListUnreadRequest::class,
            responseClass = NotificationOuterClass.ListUnreadResponse::class,
            streamType = StreamType.UNARY,
        )
        val res: ResponseMessage<NotificationOuterClass.ListUnreadResponse> =
            client.unary(req, emptyMap(), spec)
        return res.getOrThrow().itemsList.map { it.toClient() }
    }

    // UnreadCount
    suspend fun unreadCount(): Long {
        val client = createClient()
        val req = NotificationOuterClass.UnreadCountRequest.newBuilder().build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/UnreadCount",
            requestClass = NotificationOuterClass.UnreadCountRequest::class,
            responseClass = NotificationOuterClass.UnreadCountResponse::class,
            streamType = StreamType.UNARY,
        )
        val res: ResponseMessage<NotificationOuterClass.UnreadCountResponse> =
            client.unary(req, emptyMap(), spec)
        return res.getOrThrow().count
    }

    // MarkAsRead
    suspend fun markRead(id: String) {
        val client = createClient()
        val req = NotificationOuterClass.MarkReadRequest.newBuilder()
            .setId(id)
            .build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/MarkRead",
            requestClass = NotificationOuterClass.MarkReadRequest::class,
            responseClass = NotificationOuterClass.MarkReadResponse::class,
            streamType = StreamType.UNARY,
        )
        client.unary(req, emptyMap(), spec)
    }

    // MarkAllRead
    suspend fun markAllRead(): Long {
        val client = createClient()
        val req = NotificationOuterClass.MarkAllReadRequest.newBuilder().build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/MarkAllRead",
            requestClass = NotificationOuterClass.MarkAllReadRequest::class,
            responseClass = NotificationOuterClass.MarkAllReadResponse::class,
            streamType = StreamType.UNARY,
        )
        val res: ResponseMessage<NotificationOuterClass.MarkAllReadResponse> =
            client.unary(req, emptyMap(), spec)
        return res.getOrThrow().marked
    }

    // GetAlertPrefs
    suspend fun getAlertPrefs(): AlertPrefs {
        val client = createClient()
        val req = NotificationOuterClass.GetAlertPrefsRequest.newBuilder().build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/GetAlertPrefs",
            requestClass = NotificationOuterClass.GetAlertPrefsRequest::class,
            responseClass = NotificationOuterClass.GetAlertPrefsResponse::class,
            streamType = StreamType.UNARY,
        )
        val res: ResponseMessage<NotificationOuterClass.GetAlertPrefsResponse> =
            client.unary(req, emptyMap(), spec)
        val p = res.getOrThrow().prefs
        return AlertPrefs(
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
    }

    // UpdateAlertPrefs
    suspend fun updateAlertPrefs(prefs: AlertPrefs) {
        val client = createClient()
        val p = NotificationOuterClass.AlertPrefs.newBuilder()
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
        val req = NotificationOuterClass.UpdateAlertPrefsRequest.newBuilder()
            .setPrefs(p)
            .build()
        val spec = MethodSpec(
            path = "antclaw.v1.NotificationService/UpdateAlertPrefs",
            requestClass = NotificationOuterClass.UpdateAlertPrefsRequest::class,
            responseClass = NotificationOuterClass.UpdateAlertPrefsResponse::class,
            streamType = StreamType.UNARY,
        )
        client.unary(req, emptyMap(), spec)
    }
}

// Proto → Client 模型转换
private fun Notification.toClient(): ClientNotification = ClientNotification(
    id = id,
    userId = userId,
    type = type,
    category = category,
    severity = severity,
    title = title,
    body = body,
    data = dataMap,
    isRead = isRead,
    createdAt = Instant.ofEpochSecond(createdAt),
    readAt = if (readAt > 0) Instant.ofEpochSecond(readAt) else null,
)
