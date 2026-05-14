package com.antclaw.alfq.ui.social

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.ui.components.CommentItem
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

/**
 * 评论区域 — 含评论列表 + 底部输入框。
 *
 * 注意：服务端 ListComments RPC 尚未实现（P0 依赖）。
 * 当前仅展示本地累积的评论（评论发送后即时追加到列表）。
 */
@Composable
fun CommentSection(
    comments: List<CommentUi>,
    onSendComment: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    var inputText by remember { mutableStateOf("") }

    Column(modifier = modifier.fillMaxSize()) {
        // ── Comment List ──
        if (comments.isEmpty()) {
            Box(
                modifier = Modifier.weight(1f).fillMaxWidth(),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = "No comments yet",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        } else {
            LazyColumn(
                modifier = Modifier.weight(1f).fillMaxWidth(),
                contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingSm),
                verticalArrangement = Arrangement.spacedBy(SpacingMd),
            ) {
                items(comments, key = { it.commentId }) { comment ->
                    Column {
                        CommentItem(comment = comment, depth = 0)
                        // Render nested replies
                        comment.replies.forEach { reply ->
                            Spacer(modifier = Modifier.height(SpacingSm))
                            CommentItem(comment = reply, depth = 1)
                        }
                    }
                }
            }
        }

        HorizontalDivider(color = MaterialTheme.colorScheme.outline.copy(alpha = 0.3f))

        // ── Comment Input ──
        Surface(
            color = MaterialTheme.colorScheme.surface,
            tonalElevation = 2.dp,
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = SpacingMd, vertical = SpacingSm),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                OutlinedTextField(
                    value = inputText,
                    onValueChange = { inputText = it },
                    placeholder = { Text("Write a comment...") },
                    modifier = Modifier.weight(1f),
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Send),
                    keyboardActions = KeyboardActions(
                        onSend = {
                            if (inputText.isNotBlank()) {
                                onSendComment(inputText.trim())
                                inputText = ""
                            }
                        }
                    ),
                )
                Spacer(modifier = Modifier.width(SpacingSm))
                IconButton(
                    onClick = {
                        if (inputText.isNotBlank()) {
                            onSendComment(inputText.trim())
                            inputText = ""
                        }
                    },
                    enabled = inputText.isNotBlank(),
                ) {
                    Icon(
                        Icons.Default.Send,
                        contentDescription = "Send",
                        tint = if (inputText.isNotBlank())
                            MaterialTheme.colorScheme.primary
                        else
                            MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f),
                    )
                }
            }
        }
    }
}
