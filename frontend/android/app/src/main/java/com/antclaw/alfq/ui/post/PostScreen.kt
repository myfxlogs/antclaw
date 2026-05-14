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
fun PostScreen(viewModel: PostViewModel = hiltViewModel(), onClose: () -> Unit = {}) {
    var content by remember { mutableStateOf("") }
    var signalPair by remember { mutableStateOf("") }
    var visibility by remember { mutableStateOf("public") }
    
    val postState by viewModel.postState.collectAsState()

    // 成功后清空内容
    LaunchedEffect(postState) {
        if (postState is PostState.Success) {
            content = ""
            signalPair = ""
            // 延迟后重置状态，允许用户再次发布
            kotlinx.coroutines.delay(2000)
            viewModel.reset()
        }
    }

    // 加载状态
    val isLoading = postState is PostState.Loading
    
    // 错误信息
    val errorMessage = (postState as? PostState.Error)?.message

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
                            viewModel.post(content, signalPair, "", 0, visibility) 
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
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf(
                    "public" to stringResource(R.string.post_visibility_public),
                    "followers" to stringResource(R.string.post_visibility_followers),
                    "circle" to stringResource(R.string.post_visibility_circle)
                ).forEach { (v, label) ->
                    FilterChip(
                        selected = visibility == v, 
                        onClick = { visibility = v }, 
                        label = { Text(label) }
                    )
                }
            }
            
            // 状态提示区域
            Spacer(modifier = Modifier.height(SpacingMd))
            
            // 成功提示
            if (postState is PostState.Success) {
                Text(stringResource(R.string.post_success), color = MaterialTheme.colorScheme.primary)
            }
            
            // 错误提示
            if (errorMessage != null) {
                Text(errorMessage, color = MaterialTheme.colorScheme.error)
                // 重试按钮
                TextButton(onClick = { viewModel.reset() }) {
                    Text(stringResource(R.string.common_retry))
                }
            }
        }
    }
}