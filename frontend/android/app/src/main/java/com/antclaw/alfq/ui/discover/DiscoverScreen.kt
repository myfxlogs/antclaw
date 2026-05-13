package com.antclaw.alfq.ui.discover

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DiscoverScreen(
    viewModel: DiscoverViewModel = hiltViewModel(),
    onTraderClick: (userId: String) -> Unit = {},
    onCircleClick: (circleId: String) -> Unit = {},
) {
    val state by viewModel.uiState.collectAsState()
    var searchQuery by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize()) {
        OutlinedTextField(
            value = searchQuery,
            onValueChange = { searchQuery = it },
            placeholder = { Text("\u641c\u7d22\u4ea4\u6613\u5458\u3001\u5708\u5b50\u3001\u4fe1\u53f7...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = "\u641c\u7d22") },
            modifier = Modifier
                .fillMaxWidth()
                .padding(SpacingMd),
            singleLine = true
        )

        when {
            state.loading -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }
            else -> {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingMd),
                    verticalArrangement = Arrangement.spacedBy(SpacingMd)
                ) {
                    // 热门交易员
                    if (state.traders.isNotEmpty()) {
                        item {
                            Text("\u70ed\u95e8\u4ea4\u6613\u5458", style = MaterialTheme.typography.titleMedium,
                                fontWeight = FontWeight.Bold)
                            Spacer(modifier = Modifier.height(SpacingSm))
                        }
                        items(state.traders.take(5), key = { it.userId }) { trader ->
                            Surface(
                                onClick = { onTraderClick(trader.userId) },
                                shape = MaterialTheme.shapes.small,
                                color = MaterialTheme.colorScheme.surfaceVariant,
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Row(
                                    Modifier.padding(SpacingMd),
                                    horizontalArrangement = Arrangement.SpaceBetween,
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Column {
                                        Text(trader.displayName.ifEmpty { trader.userId },
                                            style = MaterialTheme.typography.bodyLarge,
                                            fontWeight = FontWeight.Bold)
                                        if (trader.tier.isNotEmpty() && trader.tier != "normal") {
                                            Text(trader.tier, style = MaterialTheme.typography.bodySmall,
                                                color = MaterialTheme.colorScheme.primary)
                                        }
                                    }
                                    OutlinedButton(onClick = { onTraderClick(trader.userId) }) {
                                        Text("\u67e5\u770b")
                                    }
                                }
                            }
                        }
                    }

                    // 提示空状态
                    if (state.traders.isEmpty() && state.error == null) {
                        item {
                            Box(
                                modifier = Modifier.fillMaxWidth().padding(vertical = 32.dp),
                                contentAlignment = Alignment.Center
                            ) {
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    Text("\u6682\u65e0\u63a8\u8350\u4ea4\u6613\u5458",
                                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                                    Spacer(modifier = Modifier.height(8.dp))
                                    Text("\u5173\u6ce8\u4e00\u4e9b\u4ea4\u6613\u5458\u540e\u5c06\u5728\u6b64\u663e\u793a",
                                        style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f))
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
