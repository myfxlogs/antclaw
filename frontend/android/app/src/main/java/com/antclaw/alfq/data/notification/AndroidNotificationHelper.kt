package com.antclaw.alfq.data.notification

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.antclaw.alfq.MainActivity
import com.antclaw.alfq.R
import java.util.concurrent.atomic.AtomicInteger

/**
 * Android 系统通知管理：创建渠道、后台推送、点击跳转。
 *
 * 文档 §8.7 要求四个渠道：
 *   - market_alerts（高优先级，可震动）
 *   - signals（默认优先级）
 *   - digests（低优先级，不震动）
 *   - system（默认优先级）
 */
class AndroidNotificationHelper(private val context: Context) {

    companion object {
        const val CHANNEL_ALERTS = "market_alerts"
        const val CHANNEL_SIGNALS = "signals"
        const val CHANNEL_DIGESTS = "digests"
        const val CHANNEL_SYSTEM = "system"

        private const val GROUP_KEY = "com.antclaw.alfq.NOTIFICATIONS"
    }

    private val notifId = AtomicInteger(1000)

    init {
        createChannels()
    }

    private fun createChannels() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val nm = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        nm.createNotificationChannel(
            NotificationChannel(
                CHANNEL_ALERTS,
                "市场警报",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "经济日历、宏观 regime、options/onchain 风险"
                enableVibration(true)
            }
        )

        nm.createNotificationChannel(
            NotificationChannel(
                CHANNEL_SIGNALS,
                "交易信号",
                NotificationManager.IMPORTANCE_DEFAULT
            ).apply {
                description = "COT 信号、surprise、多资产共振"
            }
        )

        nm.createNotificationChannel(
            NotificationChannel(
                CHANNEL_DIGESTS,
                "摘要与报告",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "晨报、周度展望、校准更新"
                enableVibration(false)
            }
        )

        nm.createNotificationChannel(
            NotificationChannel(
                CHANNEL_SYSTEM,
                "系统通知",
                NotificationManager.IMPORTANCE_DEFAULT
            ).apply {
                description = "账户、系统维护通知"
            }
        )
    }

    /**
     * App 在后台时，对 high/critical 通知发送 Android 系统通知。
     * digest 类型不发送系统通知（低打扰）。
     */
    fun showNotificationForAppInBackground(notif: ClientNotification) {
        val severity = notif.severity.lowercase()
        // 只对 high/critical 发送系统通知；digest 不弹
        if (severity != "high" && severity != "critical") return

        val channel = when (notif.category.lowercase()) {
            "alert" -> CHANNEL_ALERTS
            "signal" -> CHANNEL_SIGNALS
            "digest" -> CHANNEL_DIGESTS
            else -> CHANNEL_SYSTEM
        }

        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            // 通过 extra 传递跳转目标
            putExtra("ALFQ_NOTIF_KIND", notif.data["kind"] ?: "")
            putExtra("ALFQ_NOTIF_DATA", HashMap(notif.data))
        }
        val requestCode = notifId.incrementAndGet()
        val pending = PendingIntent.getActivity(
            context,
            requestCode,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )

        val builder = NotificationCompat.Builder(context, channel)
            .setSmallIcon(com.antclaw.alfq.R.drawable.ic_notification_bell)
            .setContentTitle(notif.title)
            .setContentText(notif.body)
            .setPriority(
                if (severity == "critical") NotificationCompat.PRIORITY_HIGH
                else NotificationCompat.PRIORITY_DEFAULT
            )
            .setAutoCancel(true)
            .setContentIntent(pending)
            .setGroup(GROUP_KEY)

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            builder.setChannelId(channel)
        }

        try {
            NotificationManagerCompat.from(context).notify(notifId.incrementAndGet(), builder.build())
        } catch (e: SecurityException) {
            // 未授权通知权限时静默跳过
        }
    }

    /** 清除所有通知 */
    fun cancelAll() {
        try {
            NotificationManagerCompat.from(context).cancelAll()
        } catch (_: Exception) { }
    }
}
