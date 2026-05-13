package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm
import com.antclaw.alfq.ui.theme.SpacingLg

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MeScreen(
    onLogout: () -> Unit,
    onNavigateToMTAccounts: () -> Unit = {},
    onNavigateToAlerts: () -> Unit = {},
    onNavigateToSettings: () -> Unit = {},
    vm: MeViewModel = hiltViewModel(),
) {
    val state by vm.state.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(
            title = { },
            actions = {
                IconButton(onClick = onNavigateToSettings) {
                    Icon(Icons.Default.Settings, contentDescription = "\u8bbe\u7f6e")
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(
                containerColor = MaterialTheme.colorScheme.background
            )
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
                    contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingMd)
                ) {
                    item {
                        Column(modifier = Modifier.padding(bottom = SpacingMd)) {
                            val tierLabel = when (state.tier) {
                                "verified" -> "\ud83d\udfe2 \u8ba4\u8bc1\u4ea4\u6613\u5458"
                                "elite" -> "\ud83d\udd35 \u7cbe\u82f1\u4ea4\u6613\u5458"
                                else -> ""
                            }
                            if (tierLabel.isNotEmpty()) {
                                Text(tierLabel, style = MaterialTheme.typography.labelSmall,
                                    color = MaterialTheme.colorScheme.primary)
                            }
                            Text(state.displayName.ifEmpty { "\u4ea4\u6613\u5458" },
                                style = MaterialTheme.typography.displaySmall, fontWeight = FontWeight.Bold)
                            Text(state.username, style = MaterialTheme.typography.bodyLarge,
                                color = MaterialTheme.colorScheme.onSurfaceVariant)
                            Spacer(modifier = Modifier.height(SpacingMd))
                            if (state.bio.isNotEmpty()) {
                                Text(state.bio, style = MaterialTheme.typography.bodyMedium)
                                Spacer(modifier = Modifier.height(SpacingMd))
                            }
                            Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                                Column {
                                    Text("${state.followingCount}", style = MaterialTheme.typography.titleMedium,
                                        fontWeight = FontWeight.Bold)
                                    Text("\u5173\u6ce8", style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                                Column {
                                    Text("${state.followerCount}", style = MaterialTheme.typography.titleMedium,
                                        fontWeight = FontWeight.Bold)
                                    Text("\u7c89\u4e1d", style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                                if (state.totalTrades > 0) {
                                    Column {
                                        val wr = (state.winRate * 100).toInt()
                                        Text("${wr}%", style = MaterialTheme.typography.titleMedium,
                                            fontWeight = FontWeight.Bold)
                                        Text("\u80dc\u7387", style = MaterialTheme.typography.bodySmall,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant)
                                    }
                                }
                            }
                        }
                        HorizontalDivider(color = MaterialTheme.colorScheme.outline)
                    }

                    item {
                        Spacer(modifier = Modifier.height(SpacingMd))
                        Text("\u8d26\u53f7\u7ba1\u7406", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        Spacer(modifier = Modifier.height(SpacingSm))
                    }

                    item {
                        Surface(
                            onClick = onNavigateToMTAccounts,
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
                                    Text("\u4ea4\u6613\u8d26\u53f7", style = MaterialTheme.typography.bodyLarge,
                                        fontWeight = FontWeight.Bold)
                                    Text("MT4/MT5 \u53ea\u8bfb\u8fde\u63a5", style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                                Text(">", color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                        }
                    }

                    item {
                        Surface(
                            onClick = onNavigateToAlerts,
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
                                    Text("\u6211\u7684\u8b66\u62a5", style = MaterialTheme.typography.bodyLarge,
                                        fontWeight = FontWeight.Bold)
                                    Text("\u4ef7\u683c/\u4fe1\u53f7/\u5b8f\u89c2", style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                                }
                                Text(">", color = MaterialTheme.colorScheme.onSurfaceVariant)
                            }
                        }
                    }

                    item {
                        Spacer(modifier = Modifier.height(SpacingLg))
                        OutlinedButton(
                            onClick = { vm.logout { onLogout() } },
                            enabled = !state.loggingOut,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            if (state.loggingOut) {
                                CircularProgressIndicator(modifier = Modifier.size(20.dp))
                            } else {
                                Text("\u9000\u51fa\u767b\u5f55")
                            }
                        }
                    }
                }
            }
        }
    }
}
