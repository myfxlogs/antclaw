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
    var showSuccess by remember { mutableStateOf(false) }
    val posted by viewModel.posted.collectAsState()
    LaunchedEffect(posted) { if (posted) { showSuccess = true; content = "" } }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { },
                navigationIcon = { IconButton(onClick = onClose) { Icon(Icons.Default.Close, contentDescription = stringResource(R.string.common_close)) } },
                actions = {
                    Button(onClick = { viewModel.post(content, signalPair, "", 0, visibility) },
                        enabled = content.isNotBlank(), modifier = Modifier.padding(end = SpacingMd)) { Text(stringResource(R.string.post_publish)) }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background)
            )
        }
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = SpacingMd), verticalArrangement = Arrangement.Top) {
            Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
                Box(modifier = Modifier.size(40.dp).padding(end = SpacingMd))
                Column(modifier = Modifier.weight(1f)) {
                    TextField(value = content, onValueChange = { content = it },
                        placeholder = { Text(stringResource(R.string.post_hint)) },
                        modifier = Modifier.fillMaxWidth(),
                        colors = TextFieldDefaults.colors(
                            focusedContainerColor = Color.Transparent, unfocusedContainerColor = Color.Transparent,
                            focusedIndicatorColor = Color.Transparent, unfocusedIndicatorColor = Color.Transparent
                        ), maxLines = 8)
                    if (signalPair.isNotBlank()) {
                        OutlinedTextField(value = signalPair, onValueChange = { signalPair = it },
                            placeholder = { Text(stringResource(R.string.post_signal_hint)) },
                            modifier = Modifier.fillMaxWidth(), singleLine = true,
                            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done))
                    }
                }
            }
            Spacer(modifier = Modifier.height(SpacingMd))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf("public" to stringResource(R.string.post_visibility_public),
                    "followers" to stringResource(R.string.post_visibility_followers),
                    "circle" to stringResource(R.string.post_visibility_circle)).forEach { (v, label) ->
                    FilterChip(selected = visibility == v, onClick = { visibility = v }, label = { Text(label) })
                }
            }
            if (showSuccess) { Spacer(modifier = Modifier.height(SpacingMd)); Text(stringResource(R.string.post_success), color = MaterialTheme.colorScheme.primary) }
        }
    }
}
