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
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.SpacingMd

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostScreen(viewModel: PostViewModel = hiltViewModel()) {
    var content by remember { mutableStateOf("") }
    var signalPair by remember { mutableStateOf("") }
    var visibility by remember { mutableStateOf("public") }
    var showSuccess by remember { mutableStateOf(false) }

    val posted by viewModel.posted.collectAsState()

    LaunchedEffect(posted) { if (posted) { showSuccess = true; content = "" } }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { },
                navigationIcon = {
                    IconButton(onClick = { /* TODO: Navigate back */ }) {
                        Icon(Icons.Default.Close, contentDescription = "关闭")
                    }
                },
                actions = {
                    Button(
                        onClick = { viewModel.post(content, signalPair, "", 0, visibility) },
                        enabled = content.isNotBlank(),
                        modifier = Modifier.padding(end = SpacingMd)
                    ) { Text("发布") }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = SpacingMd),
            verticalArrangement = Arrangement.Top
        ) {
            // Avatar + Input
            Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
                // Avatar placeholder
                Box(
                    modifier = Modifier
                        .size(40.dp)
                        .padding(end = SpacingMd)
                ) {
                    // User avatar would go here
                }

                // Content input
                Column(modifier = Modifier.weight(1f)) {
                    TextField(
                        value = content,
                        onValueChange = { content = it },
                        placeholder = { Text("分享你的交易观点...") },
                        modifier = Modifier.fillMaxWidth(),
                        colors = TextFieldDefaults.colors(
                            focusedContainerColor = Color.Transparent,
                            unfocusedContainerColor = Color.Transparent,
                            focusedIndicatorColor = Color.Transparent,
                            unfocusedIndicatorColor = Color.Transparent
                        ),
                        maxLines = 8
                    )

                    // Signal pair input
                    if (signalPair.isNotBlank()) {
                        OutlinedTextField(
                            value = signalPair,
                            onValueChange = { signalPair = it },
                            placeholder = { Text("引用信号（如 EURUSD）") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(SpacingMd))

            // Visibility options
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf("public" to "公开", "followers" to "关注者", "circle" to "圈子").forEach { (v, label) ->
                    FilterChip(selected = visibility == v, onClick = { visibility = v }, label = { Text(label) })
                }
            }

            if (showSuccess) {
                Spacer(modifier = Modifier.height(SpacingMd))
                Text("发布成功！", color = MaterialTheme.colorScheme.primary)
            }
        }
    }
}
