package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.theme.BullGreen

@Composable
fun MTAccountsScreen(
    viewModel: MTAccountsViewModel = hiltViewModel(),
    onBack: () -> Unit,
    onBindClick: () -> Unit = {},
) {
    val state by viewModel.uiState.collectAsState()
    Column(modifier = Modifier.fillMaxSize()) {
        Row(modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) }
            Text(stringResource(R.string.mt_title), style = MaterialTheme.typography.titleLarge)
            TextButton(onClick = onBindClick) {
                Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(18.dp))
                Spacer(modifier = Modifier.width(4.dp))
                Text(stringResource(R.string.mt_bind_new))
            }
        }
        when {
            state.loading -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
            state.error != null -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(stringResource(R.string.mt_load_failed), style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.error)
                    Text(state.error ?: "", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                    Spacer(modifier = Modifier.height(16.dp))
                    OutlinedButton(onClick = { viewModel.loadAccounts() }) { Text(stringResource(R.string.common_retry)) }
                }
            }
            state.accounts.isEmpty() -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(stringResource(R.string.mt_empty_title), style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                    Spacer(modifier = Modifier.height(8.dp))
                    Button(onClick = onBindClick) {
                        Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(18.dp))
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(stringResource(R.string.mt_empty_action))
                    }
                }
            }
            else -> {
                LazyColumn(modifier = Modifier.padding(horizontal = 16.dp)) {
                    items(state.accounts, key = { "${it.type}_${it.id}" }) { account ->
                        MtAccountCard(account = account, onRemove = {
                            if (account.type == "MT4") viewModel.removeMt4Account(account.id)
                            else viewModel.removeMt5Account(account.id)
                        })
                        Spacer(modifier = Modifier.height(8.dp))
                    }
                }
            }
        }
    }
}

@Composable
fun MtAccountCard(account: MtAccountItem, onRemove: () -> Unit) {
    Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Column {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("[${account.type}] ", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Bold)
                        Text("${account.label.ifEmpty { account.server }} #${account.account}", style = MaterialTheme.typography.titleSmall)
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text(if (account.isDemo) stringResource(R.string.mt_demo) else stringResource(R.string.mt_live),
                            style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                        if (account.connected) { Text(stringResource(R.string.mt_connected), style = MaterialTheme.typography.bodySmall, color = BullGreen) }
                    }
                }
                TextButton(onClick = onRemove) { Text(stringResource(R.string.mt_disconnect), color = MaterialTheme.colorScheme.error) }
            }
            if (account.balance > 0 || account.equity > 0) {
                Spacer(modifier = Modifier.height(12.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                    AccountStat(stringResource(R.string.mt_balance), "$${String.format("%.0f", account.balance)}")
                    AccountStat(stringResource(R.string.mt_equity), "$${String.format("%.0f", account.equity)}")
                    AccountStat(stringResource(R.string.mt_server), account.server)
                }
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
