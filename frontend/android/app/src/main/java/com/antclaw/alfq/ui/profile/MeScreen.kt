package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm
import com.antclaw.alfq.ui.theme.SpacingLg

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MeScreen(
    onLogout: () -> Unit,
    onNavigateToMTAccounts: () -> Unit = {},
    onNavigateToAlerts: () -> Unit = {},
) {
    Column(modifier = Modifier.fillMaxSize()) {
        // Top Bar
        TopAppBar(
            title = { },
            actions = {
                IconButton(onClick = { /* TODO: Navigate to settings */ }) {
                    Icon(Icons.Default.Settings, contentDescription = "设置")
                }
            },
            colors = TopAppBarDefaults.topAppBarColors(
                containerColor = MaterialTheme.colorScheme.background
            )
        )

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingMd)
        ) {
            item {
                // Profile Header
                Column(modifier = Modifier.padding(bottom = SpacingMd)) {
                    Text("交易员名称", style = MaterialTheme.typography.displaySmall, fontWeight = FontWeight.Bold)
                    Text("@trader_username", style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(modifier = Modifier.height(SpacingMd))
                    Text("专注外汇与加密货币交易，分享实时信号与交易策略", style = MaterialTheme.typography.bodyMedium)
                    Spacer(modifier = Modifier.height(SpacingMd))
                    Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
                        Column { Text("128", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold); Text("关注", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                        Column { Text("1.2K", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold); Text("粉丝", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                        Column { Text("87%", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold); Text("胜率", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                    }
                }
                HorizontalDivider(color = MaterialTheme.colorScheme.outline)
            }

            item {
                Spacer(modifier = Modifier.height(SpacingMd))
                Text("账号管理", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
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
                            Text("交易账号", style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
                            Text("MT4/MT5 只读连接", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
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
                            Text("我的警报", style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
                            Text("价格/信号/宏观", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                        Text(">", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }

            item {
                Spacer(modifier = Modifier.height(SpacingLg))
                OutlinedButton(onClick = onLogout, modifier = Modifier.fillMaxWidth()) { Text("退出登录") }
            }
        }
    }
}
