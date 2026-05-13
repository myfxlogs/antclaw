package com.antclaw.alfq.ui.notification

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.antclaw.alfq.data.notification.AlertPrefs

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationPrefsScreen(
    onBack: () -> Unit,
    vm: NotificationViewModel = viewModel(),
) {
    val prefs by vm.prefs.collectAsState()
    val loading by vm.prefsLoading.collectAsState()

    LaunchedEffect(Unit) { vm.loadPrefs() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("\u63a8\u9001\u504f\u597d", color = MaterialTheme.colorScheme.onSurface) },
                navigationIcon = { TextButton(onClick = onBack) { Text("\u8fd4\u56de") } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
            )
        },
        containerColor = MaterialTheme.colorScheme.background,
    ) { padding ->
        if (loading) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                item { SectionHeader("\u544a\u8b66\u9891\u9053") }

                item {
                    ToggleItem("\u5e02\u573a\u544a\u8b66", prefs.optionsAlertsEnabled) {
                        vm.updatePrefs(prefs.copy(optionsAlertsEnabled = it))
                    }
                }
                item {
                    ToggleItem("\u5b8f\u89c2\u544a\u8b66", prefs.macroAlertsEnabled) {
                        vm.updatePrefs(prefs.copy(macroAlertsEnabled = it))
                    }
                }
                item {
                    ToggleItem("COT \u504f\u5411", prefs.cotAlertsEnabled) {
                        vm.updatePrefs(prefs.copy(cotAlertsEnabled = it))
                    }
                }
                item {
                    ToggleItem("\u94fe\u4e0a\u6570\u636e", prefs.onchainAlertsEnabled) {
                        vm.updatePrefs(prefs.copy(onchainAlertsEnabled = it))
                    }
                }

                item { SectionHeader("\u6458\u8981") }

                item {
                    ToggleItem("\u6bcf\u65e5\u6458\u8981", prefs.dailyDigestEnabled) {
                        vm.updatePrefs(prefs.copy(dailyDigestEnabled = it))
                    }
                }
                item {
                    ToggleItem("\u6bcf\u5468\u6458\u8981", prefs.weeklyDigestEnabled) {
                        vm.updatePrefs(prefs.copy(weeklyDigestEnabled = it))
                    }
                }

                item { SectionHeader("\u8fc7\u6ee4") }

                item {
                    ToggleItem("\u4ec5\u9ad8\u5f71\u54cd\u4e8b\u4ef6", prefs.highImpactOnly) {
                        vm.updatePrefs(prefs.copy(highImpactOnly = it))
                    }
                }

                item {
                    Spacer(modifier = Modifier.height(32.dp))
                }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String) {
    Text(
        title,
        style = MaterialTheme.typography.titleSmall,
        fontWeight = FontWeight.Bold,
        color = MaterialTheme.colorScheme.primary,
        modifier = Modifier.padding(top = 8.dp),
    )
}

@Composable
private fun ToggleItem(
    label: String,
    checked: Boolean,
    onToggle: (Boolean) -> Unit,
) {
    Surface(
        shape = MaterialTheme.shapes.small,
        color = MaterialTheme.colorScheme.surfaceVariant,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(label, style = MaterialTheme.typography.bodyMedium)
            Switch(checked = checked, onCheckedChange = onToggle)
        }
    }
}
