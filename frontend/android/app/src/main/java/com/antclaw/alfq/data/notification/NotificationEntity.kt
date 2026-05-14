package com.antclaw.alfq.data.notification

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "notifications")
data class NotificationEntity(
    @PrimaryKey val id: String,
    val userId: String = "",
    val type: String = "in_app",
    val category: String = "system",
    val severity: String = "normal",
    val title: String = "",
    val body: String = "",
    val dataJson: String = "{}",
    val isRead: Boolean = false,
    val createdAt: Long = 0,
    val readAt: Long? = null,
)
