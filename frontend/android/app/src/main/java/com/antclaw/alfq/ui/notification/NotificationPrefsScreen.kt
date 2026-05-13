package com.antclaw.alfq.ui.notification

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.antclaw.alfq.data.notification.AlertPrefs

// 业务推送偏好设置页。
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationPrefsScreen(
    onBack: () -> Unit,
    vm: NotificationViewModel = viewModel(),
) {
    val prefs by vm.alertPrefs.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("推送偏好", color = MaterialTheme.colorScheme.onSurface) },
                navigationIcon = {
                    TextButton(onClick = onBack) { Text("返回") }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.background,
                ),
            )
        },
        containerColor = MaterialTheme.colorScheme.background,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 16.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text("频道开关", style = MaterialTheme.typography.titleSmall)

            SwitchRow("每日晨报", prefs.dailyDigestEnabled) {
                vm.updateAlertPrefs(prefs.copy(dailyDigestEnabled = it))
            }
            SwitchRow("周度展望", prefs.weeklyDigestEnabled) {
                vm.updateAlertPrefs(prefs.copy(weeklyDigestEnabled = it))
            }
            SwitchRow("COT 持仓提醒", prefs.cotAlertsEnabled) {
                vm.updateAlertPrefs(prefs.copy(cotAlertsEnabled = it))
            }
            SwitchRow("宏观 regime 提醒", prefs.macroAlertsEnabled) {
                vm.updateAlertPrefs(prefs.copy(macroAlertsEnabled = it))
            }
            SwitchRow("期权风险提醒", prefs.optionsAlertsEnabled) {
                vm.updateAlertPrefs(prefs.copy(optionsAlertsEnabled = it))
            }
            SwitchRow("链上数据提醒", prefs.onchainAlertsEnabled) {
                vm.updateAlertPrefs(prefs.copy(onchainAlertsEnabled = it))
            }

            Divider(modifier = Modifier.padding(vertical = 8.dp))

            Text("事件过滤", style = MaterialTheme.typography.titleSmall)

            SwitchRow("仅高影响事件", prefs.highImpactOnly) {
                vm.updateAlertPrefs(prefs.copy(highImpactOnly = it))
            }

            Text(
                "关注货币：${prefs.currencies.joinToString(", ")}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun SwitchRow(label: String, checked: Boolean, onToggle: (Boolean) -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
    ) {
        Text(label, style = MaterialTheme.typography.bodyMedium)
        Switch(checked = checked, onCheckedChange = onToggle)
    }
}
