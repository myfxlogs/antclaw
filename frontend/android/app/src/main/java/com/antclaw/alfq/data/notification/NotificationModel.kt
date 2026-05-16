package com.antclaw.alfq.data.notification

import antclaw.v1.NotificationOuterClass
import java.time.Instant

// 对应 proto antclaw.v1.Notification，客户端本地模型。
data class ClientNotification(
    val id: String = "",
    val userId: String = "",
    val type: String = "in_app",
    val category: String = "system",
    val severity: String = "normal",
    val title: String = "",
    val body: String = "",
    val data: Map<String, String> = emptyMap(),
    val isRead: Boolean = false,
    val createdAt: Instant = Instant.now(),
    val readAt: Instant? = null,
) {
    companion object {
        fun fromProto(proto: NotificationOuterClass.Notification): ClientNotification = ClientNotification(
            id = proto.id,
            userId = proto.userId,
            type = proto.type,
            category = proto.category,
            severity = proto.severity,
            title = proto.title,
            body = proto.body,
            data = proto.dataMap,
            isRead = proto.isRead,
            createdAt = Instant.ofEpochSecond(proto.createdAt),
            readAt = if (proto.readAt > 0) Instant.ofEpochSecond(proto.readAt) else null,
        )
    }
}

// 通知 UI 状态。
data class NotificationUiState(
    val unreadCount: Int = 0,
    val items: List<ClientNotification> = emptyList(),
    val connected: Boolean = false,
    val error: String? = null,
)

// 业务告警偏好（对应 proto antclaw.v1.AlertPrefs）。
data class AlertPrefs(
    val currencies: List<String> = listOf("USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "NZD"),
    val symbols: List<String> = emptyList(),
    val impacts: List<String> = listOf("high", "medium"),
    val reminderMinutes: List<Int> = listOf(60, 15),
    val highImpactOnly: Boolean = false,
    val dailyDigestEnabled: Boolean = true,
    val weeklyDigestEnabled: Boolean = true,
    val cotAlertsEnabled: Boolean = true,
    val macroAlertsEnabled: Boolean = true,
    val optionsAlertsEnabled: Boolean = true,
    val onchainAlertsEnabled: Boolean = true,
)
