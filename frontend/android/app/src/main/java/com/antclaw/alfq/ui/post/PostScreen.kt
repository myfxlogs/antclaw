package com.antclaw.alfq.ui.post

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.theme.SpacingMd

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostScreen(
    viewModel: PostViewModel = hiltViewModel(),
    onClose: () -> Unit = {},
    onPostCreated: () -> Unit = {},
) {
    var content by remember { mutableStateOf("") }
    var signalPair by remember { mutableStateOf("") }
    var visibility by remember { mutableStateOf("public") }

    val postState by viewModel.postState.collectAsState()
    val isLoading = postState is PostState.Loading
    val appError = (postState as? PostState.Error)?.error

    // 发布成功后：清空输入并通过回调通知父级（导航回 Feed）
    LaunchedEffect(postState) {
        if (postState is PostState.Success) {
            content = ""
            signalPair = ""
            viewModel.reset()
            onPostCreated()
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { },
                navigationIcon = {
                    IconButton(onClick = onClose) {
                        Icon(Icons.Default.Close, contentDescription = stringResource(R.string.common_close))
                    }
                },
                actions = {
                    Button(
                        onClick = {
                            viewModel.post(
                                PostDraft(content = content, signalPair = signalPair, visibility = visibility)
                            )
                        },
                        enabled = content.isNotBlank() && !isLoading,
                        modifier = Modifier.padding(end = SpacingMd)
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text(stringResource(R.string.post_publish))
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background)
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = SpacingMd),
            verticalArrangement = Arrangement.Top
        ) {
            Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
                Box(modifier = Modifier.size(40.dp).padding(end = SpacingMd))
                Column(modifier = Modifier.weight(1f)) {
                    TextField(
                        value = content,
                        onValueChange = { content = it },
                        placeholder = { Text(stringResource(R.string.post_hint)) },
                        modifier = Modifier.fillMaxWidth(),
                        colors = TextFieldDefaults.colors(
                            focusedContainerColor = Color.Transparent,
                            unfocusedContainerColor = Color.Transparent,
                            focusedIndicatorColor = Color.Transparent,
                            unfocusedIndicatorColor = Color.Transparent
                        ),
                        maxLines = 8
                    )
                    if (signalPair.isNotBlank()) {
                        OutlinedTextField(
                            value = signalPair,
                            onValueChange = { signalPair = it },
                            placeholder = { Text(stringResource(R.string.post_signal_hint)) },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done)
                        )
                    }
                }
            }
            Spacer(modifier = Modifier.height(SpacingMd))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                // 公开 / 关注者 — 可用
                listOf(
                    "public" to R.string.post_visibility_public,
                    "followers" to R.string.post_visibility_followers,
                ).forEach { (v, resId) ->
                    FilterChip(
                        selected = visibility == v,
                        onClick = { visibility = v },
                        label = { Text(stringResource(resId)) },
                    )
                }
                // 圈子 — 后端未实现，前端禁用入口
                FilterChip(
                    selected = false,
                    onClick = {},
                    enabled = false,
                    label = { Text(stringResource(R.string.post_visibility_circle)) },
                )
                Text(
                    text = stringResource(R.string.post_visibility_unavailable),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                )
            }

            Spacer(modifier = Modifier.height(SpacingMd))

            if (appError != null) {
                Text(stringResource(appError.userMessageRes), color = MaterialTheme.colorScheme.error)
                TextButton(onClick = { viewModel.retry() }) {
                    Text(stringResource(R.string.common_retry))
                }
            }
        }
    }
}
