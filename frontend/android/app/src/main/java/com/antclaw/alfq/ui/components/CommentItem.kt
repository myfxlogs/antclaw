package com.antclaw.alfq.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import com.antclaw.alfq.R
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.ui.social.CommentUi
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

/**
 * 评论条目组件 — 支持多级嵌套缩进。
 */
@Composable
fun CommentItem(
    comment: CommentUi,
    modifier: Modifier = Modifier,
    depth: Int = 0,
    onReplyClick: (() -> Unit)? = null,
) {
    Row(
        modifier = modifier.fillMaxWidth().padding(start = (depth * 28).dp),
    ) {
        // Avatar circle
        Surface(
            modifier = Modifier.size(28.dp),
            shape = CircleShape,
            color = MaterialTheme.colorScheme.primary.copy(alpha = 0.15f),
        ) {
            Box(contentAlignment = Alignment.Center) {
                Text(comment.authorName.take(1).uppercase(), style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.Bold, color = MaterialTheme.colorScheme.primary)
            }
        }
        Spacer(modifier = Modifier.width(SpacingSm))
        Column(modifier = Modifier.weight(1f)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(comment.authorName, style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.width(SpacingSm))
                Text(comment.createdAt.timeAgo(), style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
            }
            Spacer(modifier = Modifier.height(2.dp))
            Text(comment.content, style = MaterialTheme.typography.bodyMedium)
            if (onReplyClick != null && depth < 2) {
                Spacer(modifier = Modifier.height(SpacingSm))
                TextButton(onClick = onReplyClick, contentPadding = PaddingValues(0.dp),
                    modifier = Modifier.height(28.dp)) {
                    Text(stringResource(R.string.comment_reply), style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                }
            }
        }
    }
}
