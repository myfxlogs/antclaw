package com.antclaw.alfq.data.notification

import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.rpc.NotificationRpc
import javax.inject.Inject
import javax.inject.Singleton
import java.time.Instant

@Singleton
class NotificationRepository @Inject constructor(
    private val rpc: NotificationRpc,
) {
    suspend fun unreadCount(): Long = rpc.unreadCount().count
    suspend fun listUnread(limit: Int = 100): List<ClientNotification> =
        rpc.listUnread(limit).itemsList.map { it.toClient() }
    suspend fun markRead(id: String) { rpc.markRead(id) }
    suspend fun markAllRead(): Long = rpc.markAllRead().marked

    suspend fun getAlertPrefs(): AlertPrefs {
        val p = rpc.getAlertPrefs().prefs
        return AlertPrefs(
            currencies = p.currenciesList, symbols = p.symbolsList, impacts = p.impactsList,
            reminderMinutes = p.reminderMinutesList, highImpactOnly = p.highImpactOnly,
            dailyDigestEnabled = p.dailyDigestEnabled, weeklyDigestEnabled = p.weeklyDigestEnabled,
            cotAlertsEnabled = p.cotAlertsEnabled, macroAlertsEnabled = p.macroAlertsEnabled,
            optionsAlertsEnabled = p.optionsAlertsEnabled, onchainAlertsEnabled = p.onchainAlertsEnabled,
        )
    }

    suspend fun updateAlertPrefs(prefs: AlertPrefs) {
        val proto = NotificationOuterClass.AlertPrefs.newBuilder()
            .addAllCurrencies(prefs.currencies).addAllSymbols(prefs.symbols).addAllImpacts(prefs.impacts)
            .addAllReminderMinutes(prefs.reminderMinutes).setHighImpactOnly(prefs.highImpactOnly)
            .setDailyDigestEnabled(prefs.dailyDigestEnabled).setWeeklyDigestEnabled(prefs.weeklyDigestEnabled)
            .setCotAlertsEnabled(prefs.cotAlertsEnabled).setMacroAlertsEnabled(prefs.macroAlertsEnabled)
            .setOptionsAlertsEnabled(prefs.optionsAlertsEnabled).setOnchainAlertsEnabled(prefs.onchainAlertsEnabled)
            .build()
        rpc.updateAlertPrefs(NotificationOuterClass.UpdateAlertPrefsRequest.newBuilder().setPrefs(proto).build())
    }
}

private fun NotificationOuterClass.Notification.toClient() = ClientNotification(
    id = id, userId = userId, type = type, category = category,
    severity = severity, title = title, body = body, data = dataMap, isRead = isRead,
    createdAt = Instant.ofEpochSecond(createdAt),
    readAt = if (readAt > 0) Instant.ofEpochSecond(readAt) else null,
)
