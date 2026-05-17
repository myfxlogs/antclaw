package com.antclaw.alfq.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.clickable
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Share
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.SignalCardUi
import com.antclaw.alfq.ui.social.ChartShareUi
import com.antclaw.alfq.ui.theme.*
import java.time.Duration
import java.time.Instant

@Composable
fun PostCard(
    post: PostUi,
    modifier: Modifier = Modifier,
    onPostClick: (String) -> Unit = {},
    onAuthorClick: (String) -> Unit = {},
    onLikeClick: () -> Unit = {},
    onShareClick: () -> Unit = {},
    onReportClick: () -> Unit = {},
) {
    var showMenu by remember { mutableStateOf(false) }

    Card(
        modifier = modifier.fillMaxWidth().clickable { onPostClick(post.postId) },
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 0.dp),
    ) {
        Row(Modifier.padding(12.dp)) {
            // 头像
            Surface(
                modifier = Modifier.size(40.dp).clickable { onAuthorClick(post.authorId) },
                shape = CircleShape,
                color = MaterialTheme.colorScheme.primaryContainer,
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(post.authorName.take(1).uppercase(), fontWeight = FontWeight.Bold)
                }
            }

            Spacer(Modifier.width(12.dp))

            Column(Modifier.weight(1f)) {
                // 头部行：用户名 + @codeId + 时间 + 菜单
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(post.authorName, fontWeight = FontWeight.Bold, style = MaterialTheme.typography.bodyMedium)
                    Spacer(Modifier.width(4.dp))
                    Text("@${post.authorCodeId.ifEmpty { post.authorId.take(8) }}",
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                        style = MaterialTheme.typography.bodySmall)
                    Text(" · ${timeAgo(post.createdAt)}",
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                        style = MaterialTheme.typography.bodySmall)
                    Spacer(Modifier.weight(1f))
                    Box {
                        IconButton(onClick = { showMenu = true }, modifier = Modifier.size(24.dp)) {
                            Icon(Icons.Default.MoreVert, stringResource(R.string.post_more_options), modifier = Modifier.size(16.dp))
                        }
                        DropdownMenu(expanded = showMenu, onDismissRequest = { showMenu = false }) {
                            DropdownMenuItem(text = { Text(stringResource(R.string.post_report)) }, onClick = { onReportClick(); showMenu = false })
                            DropdownMenuItem(text = { Text(stringResource(R.string.post_not_interested)) }, onClick = { showMenu = false })
                        }
                    }
                }

                Spacer(Modifier.height(4.dp))

                // 正文
                Text(post.content, style = MaterialTheme.typography.bodyLarge, maxLines = 10, overflow = TextOverflow.Ellipsis)

                // 信号卡片
                post.signalCard?.let { signal ->
                    Spacer(Modifier.height(8.dp))
                    SignalCardEmbed(signal)
                }
                // 图表分享
                post.chartShare?.let { chart ->
                    Spacer(Modifier.height(8.dp))
                    ChartShareEmbed(chart)
                }

                Spacer(Modifier.height(8.dp))

                // 互动按钮行
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    ActionButton(Icons.Default.Email, "${post.commentCount}", stringResource(R.string.a11y_comment), onClick = { onPostClick(post.postId) })
                    ActionButton(Icons.Default.Share, "${post.shareCount}", stringResource(R.string.a11y_share), onClick = { onShareClick() })
                    ActionButton(
                        if (post.isLiked) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
                        "${post.likeCount}",
                        stringResource(R.string.a11y_like),
                        onClick = { onLikeClick() },
                        tint = if (post.isLiked) Color.Red else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                    )
                }
            }
        }
    }
}

@Composable
private fun ActionButton(icon: ImageVector, text: String, contentDesc: String, onClick: () -> Unit = {}, tint: Color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)) {
    Row(Modifier.clickable(onClick = onClick).padding(horizontal = 4.dp, vertical = 2.dp), verticalAlignment = Alignment.CenterVertically) {
        Icon(icon, contentDescription = contentDesc, modifier = Modifier.size(18.dp), tint = tint)
        if (text != "0") {
            Spacer(Modifier.width(2.dp))
            Text(text, style = MaterialTheme.typography.labelSmall, color = tint)
        }
    }
}

@Composable
private fun SignalCardEmbed(signal: SignalCardUi) {
    Card(
        Modifier.fillMaxWidth().padding(0.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
    ) {
        Row(Modifier.padding(SpacingSm), verticalAlignment = Alignment.CenterVertically) {
            Text(signal.pair, fontWeight = FontWeight.Bold, style = MaterialTheme.typography.labelLarge)
            Spacer(Modifier.width(SpacingSm))
            Text(signal.direction, fontSize = MaterialTheme.typography.bodySmall.fontSize,
                color = if (signal.direction.lowercase() == "buy") BullGreen else BearRed)
            Spacer(Modifier.width(4.dp))
            Text("${signal.confidence}%", style = MaterialTheme.typography.bodySmall)
            Spacer(Modifier.weight(1f))
            Text(stringResource(R.string.signal_confidence_label), style = MaterialTheme.typography.labelSmall)
            Spacer(Modifier.width(SpacingSm))
            Text("${signal.confidence}%", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary)
        }
    }
}

@Composable
private fun ChartShareEmbed(chart: ChartShareUi) {
    // 图表分享缩略图占位
    Surface(Modifier.fillMaxWidth().height(160.dp), color = MaterialTheme.colorScheme.surfaceVariant) {
        Box(contentAlignment = Alignment.Center) {
            Text("📈 ${chart.pair}", style = MaterialTheme.typography.titleMedium)
        }
    }
}

@Composable
private fun timeAgo(instant: Instant): String {
    val seconds = Duration.between(instant, Instant.now()).seconds
    return when {
        seconds < 60 -> stringResource(R.string.time_just_now)
        seconds < 3600 -> stringResource(R.string.time_minutes_short, seconds / 60)
        seconds < 86400 -> stringResource(R.string.time_hours_short, seconds / 3600)
        seconds < 259200 -> stringResource(R.string.time_days_short, seconds / 86400)
        else -> {
            val dt = instant.atZone(java.time.ZoneId.systemDefault())
            "${dt.monthValue}/${dt.dayOfMonth}"
        }
    }
}
