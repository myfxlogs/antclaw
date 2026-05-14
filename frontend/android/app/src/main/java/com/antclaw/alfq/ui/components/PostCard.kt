package com.antclaw.alfq.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Share
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.ui.social.PostType
import com.antclaw.alfq.ui.social.PostUi
import com.antclaw.alfq.ui.social.PostVisibility
import com.antclaw.alfq.ui.social.SignalCardUi
import com.antclaw.alfq.ui.social.ChartShareUi
import com.antclaw.alfq.ui.theme.BearRed
import com.antclaw.alfq.ui.theme.BullGreen
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm
import com.antclaw.alfq.ui.theme.SpacingXs

/**
 * 通用帖子卡片 — 支持 TEXT / SIGNAL_CARD / CHART_SHARE / SHARE 四种类型。
 */
@Composable
fun PostCard(
    post: PostUi,
    onLikeClick: () -> Unit = {},
    onCommentClick: () -> Unit = {},
    onShareClick: () -> Unit = {},
    onCardClick: () -> Unit = {},
    modifier: Modifier = Modifier,
) {
    Card(
        onClick = onCardClick,
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant),
    ) {
        Column(modifier = Modifier.padding(SpacingMd)) {
            PostHeader(post.authorName, post.createdAt.timeAgo(), post.visibility)
            Spacer(modifier = Modifier.height(SpacingSm))
            PostBody(post)
            Spacer(modifier = Modifier.height(SpacingMd))
            PostActions(post, onLikeClick, onCommentClick, onShareClick)
        }
    }
}

// ── Sub-composables ──

@Composable
private fun PostHeader(authorName: String, timeAgo: String, visibility: PostVisibility) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Surface(
            modifier = Modifier.size(36.dp),
            shape = CircleShape,
            color = MaterialTheme.colorScheme.primary.copy(alpha = 0.2f),
        ) {
            Box(contentAlignment = Alignment.Center) {
                Text(authorName.take(1).uppercase(), style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.primary)
            }
        }
        Spacer(modifier = Modifier.width(SpacingSm))
        Column {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(authorName, style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                if (visibility != PostVisibility.PUBLIC) {
                    Spacer(modifier = Modifier.width(SpacingXs))
                    Text(visibilityLabel(visibility), style = MaterialTheme.typography.labelSmall)
                }
            }
            Text(timeAgo, style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
        }
    }
}

@Composable
private fun PostBody(post: PostUi) {
    when (post.postType) {
        PostType.SIGNAL_CARD -> SignalPostBody(post)
        PostType.CHART_SHARE -> ChartShareBody(post)
        PostType.SHARE -> SharePostBody(post.content)
        else -> TextPostBody(post.content)
    }
}

@Composable
private fun TextPostBody(content: String) {
    Text(content, style = MaterialTheme.typography.bodyMedium, maxLines = 8, overflow = TextOverflow.Ellipsis)
}

@Composable
private fun SignalPostBody(post: PostUi) {
    Column {
        if (post.content.isNotBlank()) {
            Text(post.content, style = MaterialTheme.typography.bodyMedium, maxLines = 4, overflow = TextOverflow.Ellipsis)
            Spacer(modifier = Modifier.height(SpacingSm))
        }
        post.signalCard?.let { signal ->
            Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
                Row(Modifier.padding(SpacingSm), verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(SpacingSm)) {
                    Text(signal.pair, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                    Text(signal.direction.uppercase(), style = MaterialTheme.typography.labelSmall,
                        color = signalDirectionColor(signal.direction))
                    Text("${signal.confidence}%", style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.primary)
                }
            }
        }
    }
}

@Composable
private fun ChartShareBody(post: PostUi) {
    Column {
        if (post.content.isNotBlank()) {
            Text(post.content, style = MaterialTheme.typography.bodyMedium, maxLines = 3, overflow = TextOverflow.Ellipsis)
            Spacer(modifier = Modifier.height(SpacingSm))
        }
        post.chartShare?.let { chart ->
            Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
                Column(Modifier.padding(SpacingSm)) {
                    Text("📈 ${chart.pair}", style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                    if (chart.chartUrl == null) {
                        Spacer(modifier = Modifier.height(SpacingSm))
                        Text("[Chart]", style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f))
                    }
                }
            }
        }
    }
}

@Composable
private fun SharePostBody(content: String) {
    Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
        Text(content.ifBlank { "Reposted" }, style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(SpacingSm),
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f))
    }
}

@Composable
private fun PostActions(post: PostUi, onLike: () -> Unit, onComment: () -> Unit, onShare: () -> Unit) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
        ActionButton(
            icon = if (post.isLiked) Icons.Default.Favorite else Icons.Default.FavoriteBorder,
            tint = if (post.isLiked) Color(0xFFE91E63) else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            count = post.likeCount,
            onClick = onLike,
        )
        ActionButton(
            icon = Icons.Default.Email,
            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            count = post.commentCount,
            onClick = onComment,
        )
        ActionButton(
            icon = Icons.Default.Share,
            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f),
            count = post.shareCount,
            onClick = onShare,
        )
    }
}

@Composable
private fun ActionButton(icon: androidx.compose.ui.graphics.vector.ImageVector, tint: Color, count: Int, onClick: () -> Unit) {
    IconButton(onClick = onClick) {
        Icon(icon, contentDescription = null, tint = tint, modifier = Modifier.size(20.dp))
    }
    if (count > 0) {
        Text(formatCount(count), style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
    }
}

// ── Helpers ──

private fun visibilityLabel(v: PostVisibility) = when (v) {
    PostVisibility.FOLLOWERS_ONLY -> "👥"
    PostVisibility.CIRCLE_ONLY -> "🔵"
    else -> ""
}

@Composable
private fun signalDirectionColor(dir: String) = when (dir) {
    "bullish" -> BullGreen
    "bearish" -> BearRed
    else -> MaterialTheme.colorScheme.onSurface
}
