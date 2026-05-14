package com.antclaw.alfq.data.notification

import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.rpc.RpcHelper
import javax.inject.Inject
import javax.inject.Singleton
import java.time.Instant

@Singleton
class NotificationRepository @Inject constructor() {

    suspend fun unreadCount(): Long = try {
        RpcHelper.unary("antclaw.v1.NotificationService/UnreadCount",
            NotificationOuterClass.UnreadCountRequest.getDefaultInstance(),
            NotificationOuterClass.UnreadCountRequest::class, NotificationOuterClass.UnreadCountResponse::class).count
    } catch (_: Exception) { 0 }

    suspend fun listUnread(limit: Int = 100): List<ClientNotification> = try {
        val req = NotificationOuterClass.ListUnreadRequest.newBuilder().setLimit(limit).build()
        RpcHelper.unary("antclaw.v1.NotificationService/ListUnread", req,
            NotificationOuterClass.ListUnreadRequest::class, NotificationOuterClass.ListUnreadResponse::class)
            .itemsList.map { it.toClient() }
    } catch (_: Exception) { emptyList() }

    suspend fun markRead(id: String) {
        try {
            val req = NotificationOuterClass.MarkReadRequest.newBuilder().setId(id).build()
            RpcHelper.unary("antclaw.v1.NotificationService/MarkRead", req,
                NotificationOuterClass.MarkReadRequest::class, NotificationOuterClass.MarkReadResponse::class)
        } catch (_: Exception) {}
    }

    suspend fun getAlertPrefs(): AlertPrefs {
        val resp = RpcHelper.unary("antclaw.v1.NotificationService/GetAlertPrefs",
            NotificationOuterClass.GetAlertPrefsRequest.getDefaultInstance(),
            NotificationOuterClass.GetAlertPrefsRequest::class,
            NotificationOuterClass.GetAlertPrefsResponse::class)
        val p = resp.prefs
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

    suspend fun updateAlertPrefs(prefs: AlertPrefs) {
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
        RpcHelper.unary("antclaw.v1.NotificationService/UpdateAlertPrefs", req,
            NotificationOuterClass.UpdateAlertPrefsRequest::class,
            NotificationOuterClass.UpdateAlertPrefsResponse::class)
    }

    suspend fun markAllRead(): Long = try {
        RpcHelper.unary("antclaw.v1.NotificationService/MarkAllRead",
            NotificationOuterClass.MarkAllReadRequest.getDefaultInstance(),
            NotificationOuterClass.MarkAllReadRequest::class, NotificationOuterClass.MarkAllReadResponse::class).marked
    } catch (_: Exception) { 0 }
}

private fun NotificationOuterClass.Notification.toClient() = ClientNotification(
    id = id, userId = userId, type = type, category = category,
    severity = severity, title = title, body = body,
    data = dataMap, isRead = isRead,
    createdAt = Instant.ofEpochSecond(createdAt),
    readAt = if (readAt > 0) Instant.ofEpochSecond(readAt) else null,
)