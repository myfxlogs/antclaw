package com.antclaw.alfq.ui.components

import java.time.Duration
import java.time.Instant

/**
 * 共享格式化工具 — timeAgo / count 缩写。
 */
fun Instant.timeAgo(): String {
    val duration = Duration.between(this, Instant.now())
    return when {
        duration.toMinutes() < 1 -> "just now"
        duration.toMinutes() < 60 -> "${duration.toMinutes()}m"
        duration.toHours() < 24 -> "${duration.toHours()}h"
        duration.toDays() < 7 -> "${duration.toDays()}d"
        else -> "${duration.toDays() / 7}w"
    }
}

fun formatCount(count: Int): String = when {
    count >= 1_000_000 -> "${count / 1_000_000}M"
    count >= 1_000 -> "${count / 1000}K"
    else -> count.toString()
}
