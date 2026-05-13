package com.antclaw.alfq.data.notification

import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.rpc.RpcHelper
import java.time.Instant

class NotificationRepository {

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
