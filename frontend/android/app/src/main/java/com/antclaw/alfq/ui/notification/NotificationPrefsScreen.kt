package com.antclaw.alfq.ui.notification

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.data.notification.AlertPrefs

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationPrefsScreen(
    onBack: () -> Unit,
    vm: NotificationViewModel = hiltViewModel(),
) {
    val prefs by vm.prefs.collectAsState()
    val loading by vm.prefsLoading.collectAsState()
    LaunchedEffect(Unit) { vm.loadPrefs() }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.notif_prefs_title), color = MaterialTheme.colorScheme.onSurface) },
                navigationIcon = { TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
            )
        },
        containerColor = MaterialTheme.colorScheme.background,
    ) { padding ->
        if (loading) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)) {
                item { SectionHeader(stringResource(R.string.notif_prefs_alerts)) }
                item { ToggleItem(stringResource(R.string.notif_prefs_market), prefs.optionsAlertsEnabled) { vm.updatePrefs(prefs.copy(optionsAlertsEnabled = it)) } }
                item { ToggleItem(stringResource(R.string.notif_prefs_macro), prefs.macroAlertsEnabled) { vm.updatePrefs(prefs.copy(macroAlertsEnabled = it)) } }
                item { ToggleItem(stringResource(R.string.notif_prefs_cot), prefs.cotAlertsEnabled) { vm.updatePrefs(prefs.copy(cotAlertsEnabled = it)) } }
                item { ToggleItem(stringResource(R.string.notif_prefs_onchain), prefs.onchainAlertsEnabled) { vm.updatePrefs(prefs.copy(onchainAlertsEnabled = it)) } }
                item { SectionHeader(stringResource(R.string.notif_prefs_digest)) }
                item { ToggleItem(stringResource(R.string.notif_prefs_daily_digest), prefs.dailyDigestEnabled) { vm.updatePrefs(prefs.copy(dailyDigestEnabled = it)) } }
                item { ToggleItem(stringResource(R.string.notif_prefs_weekly_digest), prefs.weeklyDigestEnabled) { vm.updatePrefs(prefs.copy(weeklyDigestEnabled = it)) } }
                item { SectionHeader(stringResource(R.string.notif_prefs_filter)) }
                item { ToggleItem(stringResource(R.string.notif_prefs_high_impact_only), prefs.highImpactOnly) { vm.updatePrefs(prefs.copy(highImpactOnly = it)) } }
                item { Spacer(modifier = Modifier.height(32.dp)) }
            }
        }
    }
}

@Composable
private fun SectionHeader(title: String) {
    Text(title, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold,
        color = MaterialTheme.colorScheme.primary, modifier = Modifier.padding(top = 8.dp))
}

@Composable
private fun ToggleItem(label: String, checked: Boolean, onToggle: (Boolean) -> Unit) {
    Surface(shape = MaterialTheme.shapes.small, color = MaterialTheme.colorScheme.surfaceVariant, modifier = Modifier.fillMaxWidth()) {
        Row(modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 4.dp),
            horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            Text(label, style = MaterialTheme.typography.bodyMedium)
            Switch(checked = checked, onCheckedChange = onToggle)
        }
    }
}
