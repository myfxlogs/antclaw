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
import androidx.navigation.NavController
import com.antclaw.alfq.ui.theme.SpacingMd

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostScreen(
    viewModel: PostViewModel = hiltViewModel(),
    onClose: () -> Unit = {},
) {
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
                    IconButton(onClick = onClose) {
                        Icon(Icons.Default.Close, contentDescription = "\u5173\u95ed")
                    }
                },
                actions = {
                    Button(
                        onClick = { viewModel.post(content, signalPair, "", 0, visibility) },
                        enabled = content.isNotBlank(),
                        modifier = Modifier.padding(end = SpacingMd)
                    ) { Text("\u53d1\u5e03") }
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
            Row(modifier = Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
                Box(
                    modifier = Modifier
                        .size(40.dp)
                        .padding(end = SpacingMd)
                )

                Column(modifier = Modifier.weight(1f)) {
                    TextField(
                        value = content,
                        onValueChange = { content = it },
                        placeholder = { Text("\u5206\u4eab\u4f60\u7684\u4ea4\u6613\u89c2\u70b9...") },
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
                            placeholder = { Text("\u5f15\u7528\u4fe1\u53f7\uff08\u5982 EURUSD\uff09") },
                            modifier = Modifier.fillMaxWidth(),
                            singleLine = true,
                            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(SpacingMd))

            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf("public" to "\u516c\u5f00", "followers" to "\u5173\u6ce8\u8005", "circle" to "\u5708\u5b50").forEach { (v, label) ->
                    FilterChip(selected = visibility == v, onClick = { visibility = v }, label = { Text(label) })
                }
            }

            if (showSuccess) {
                Spacer(modifier = Modifier.height(SpacingMd))
                Text("\u53d1\u5e03\u6210\u529f\uff01", color = MaterialTheme.colorScheme.primary)
            }
        }
    }
}
