package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.BullGreen
import com.antclaw.alfq.ui.theme.BearRed

@Composable
fun MTAccountsScreen(
    viewModel: MTAccountsViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.uiState.collectAsState()

    Column(modifier = Modifier.fillMaxSize()) {
        // Header
        Row(
            modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            TextButton(onClick = onBack) { Text("← 返回") }
            Text("交易账号", style = MaterialTheme.typography.titleLarge)
            TextButton(onClick = { viewModel.showAddDialog = true }) {
                Text("+ 添加")
            }
        }

        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (state.accounts.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("暂无交易账号", style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                    Spacer(modifier = Modifier.height(8.dp))
                    TextButton(onClick = { viewModel.showAddDialog = true }) {
                        Text("添加 MT5 账号")
                    }
                }
            }
        } else {
            LazyColumn(modifier = Modifier.padding(horizontal = 16.dp)) {
                items(state.accounts) { account ->
                    MTAccountCard(account, onRemove = { viewModel.removeAccount(account.id) })
                    Spacer(modifier = Modifier.height(8.dp))
                }
            }
        }
    }

    // Add Account Dialog
    if (viewModel.showAddDialog) {
        AddAccountDialog(
            onDismiss = { viewModel.showAddDialog = false },
            onAdd = { server, account, password, label, isDemo ->
                viewModel.addAccount(server, account, password, label, isDemo)
            }
        )
    }
}

@Composable
fun MTAccountCard(account: MTAccountUi, onRemove: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text(
                        "${account.label.ifEmpty { account.server }} #${account.account}",
                        style = MaterialTheme.typography.titleSmall
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        if (account.isDemo) "模拟盘" else "实盘",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)
                    )
                }
                TextButton(onClick = onRemove) {
                    Text("断开", color = MaterialTheme.colorScheme.error)
                }
            }

            if (account.info != null) {
                Spacer(modifier = Modifier.height(12.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                    AccountStat("余额", "$${String.format("%.0f", account.info!!.balance)}")
                    AccountStat("净值", "$${String.format("%.0f", account.info!!.equity)}")
                    AccountStat(
                        "今日",
                        "${if (account.info!!.today_pnl >= 0) "+" else ""}${String.format("%.1f", account.info!!.today_pnl * 100)}%",
                        if (account.info!!.today_pnl >= 0) BullGreen else BearRed
                    )
                }
            }

            if (account.positions.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    "持仓 ${account.positions.size} 笔",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

@Composable
fun AccountStat(label: String, value: String, color: androidx.compose.ui.graphics.Color = MaterialTheme.colorScheme.onSurface) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, style = MaterialTheme.typography.titleMedium, color = color)
        Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
    }
}


