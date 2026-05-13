package com.antclaw.alfq.data.notification

import antclaw.v1.NotificationOuterClass
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import java.time.Instant

class NotificationRepository {

    private fun client() = ConnectTransportProvider.createProtocolClient()

    suspend fun unreadCount(): Long {
        return try {
            val spec = MethodSpec("antclaw.v1.NotificationService/UnreadCount",
                NotificationOuterClass.UnreadCountRequest::class,
                NotificationOuterClass.UnreadCountResponse::class,
                StreamType.UNARY)
            val resp: ResponseMessage<NotificationOuterClass.UnreadCountResponse> =
                client().unary(NotificationOuterClass.UnreadCountRequest.getDefaultInstance(), emptyMap(), spec)
            resp.getOrThrow().count
        } catch (_: Exception) { 0 }
    }

    suspend fun listUnread(limit: Int = 100): List<ClientNotification> {
        return try {
            val req = NotificationOuterClass.ListUnreadRequest.newBuilder().setLimit(limit).build()
            val spec = MethodSpec("antclaw.v1.NotificationService/ListUnread",
                NotificationOuterClass.ListUnreadRequest::class,
                NotificationOuterClass.ListUnreadResponse::class,
                StreamType.UNARY)
            val resp: ResponseMessage<NotificationOuterClass.ListUnreadResponse> = client().unary(req, emptyMap(), spec)
            resp.getOrThrow().itemsList.map { it.toClient() }
        } catch (_: Exception) { emptyList() }
    }

    suspend fun markRead(id: String) {
        try {
            val req = NotificationOuterClass.MarkReadRequest.newBuilder().setId(id).build()
            val spec = MethodSpec("antclaw.v1.NotificationService/MarkRead",
                NotificationOuterClass.MarkReadRequest::class,
                NotificationOuterClass.MarkReadResponse::class,
                StreamType.UNARY)
            client().unary(req, emptyMap(), spec)
        } catch (_: Exception) {}
    }

    suspend fun markAllRead(): Long {
        return try {
            val spec = MethodSpec("antclaw.v1.NotificationService/MarkAllRead",
                NotificationOuterClass.MarkAllReadRequest::class,
                NotificationOuterClass.MarkAllReadResponse::class,
                StreamType.UNARY)
            val resp: ResponseMessage<NotificationOuterClass.MarkAllReadResponse> =
                client().unary(NotificationOuterClass.MarkAllReadRequest.getDefaultInstance(), emptyMap(), spec)
            resp.getOrThrow().marked
        } catch (_: Exception) { 0 }
    }
}

private fun NotificationOuterClass.Notification.toClient() = ClientNotification(
    id = id, userId = userId, type = type, category = category,
    severity = severity, title = title, body = body,
    data = dataMap, isRead = isRead,
    createdAt = Instant.ofEpochSecond(createdAt),
    readAt = if (readAt > 0) Instant.ofEpochSecond(readAt) else null,
)
