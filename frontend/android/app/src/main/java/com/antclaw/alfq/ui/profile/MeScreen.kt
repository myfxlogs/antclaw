package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.components.TraderStatRow
import com.antclaw.alfq.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MeScreen(
    onLogout: () -> Unit, onNavigateToMTAccounts: () -> Unit = {}, onNavigateToAlerts: () -> Unit = {},
    onNavigateToSettings: () -> Unit = {}, vm: MeViewModel = hiltViewModel(),
) {
    val state by vm.state.collectAsState()
    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(
            title = { }, actions = {
                IconButton(onClick = onNavigateToSettings) { Icon(Icons.Default.Settings, contentDescription = stringResource(R.string.me_settings)) }
            },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background)
        )
        if (state.loading) Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        else LazyColumn(Modifier.fillMaxSize(), contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingMd)) {
            item {
                Column(modifier = Modifier.padding(bottom = SpacingMd)) {
                    val tierLabel = when (state.tier) { "verified" -> stringResource(R.string.tier_verified); "elite" -> stringResource(R.string.tier_elite); else -> "" }
                    if (tierLabel.isNotEmpty()) Text(tierLabel, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.primary)
                    Text(state.displayName.ifEmpty { stringResource(R.string.me_default_name) }, style = MaterialTheme.typography.displaySmall, fontWeight = FontWeight.Bold)
                    Text(state.username, style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    if (state.bio.isNotEmpty()) { Spacer(Modifier.height(SpacingMd)); Text(state.bio, style = MaterialTheme.typography.bodyMedium) }
                    Spacer(Modifier.height(SpacingMd))
                    TraderStatRow(
                        followingLabel = stringResource(R.string.me_following), followingCount = state.followingCount,
                        followersLabel = stringResource(R.string.me_followers), followerCount = state.followerCount,
                        winRateLabel = stringResource(R.string.me_win_rate), winRateValue = if (state.totalTrades > 0) "${(state.winRate * 100).toInt()}%" else "")
                }
                HorizontalDivider(color = MaterialTheme.colorScheme.outline)
            }
            item { Spacer(Modifier.height(SpacingMd)); Text(stringResource(R.string.me_account_management), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold); Spacer(Modifier.height(SpacingSm)) }
            item {
                Surface(onClick = onNavigateToMTAccounts, shape = MaterialTheme.shapes.small, color = MaterialTheme.colorScheme.surfaceVariant, modifier = Modifier.fillMaxWidth()) {
                    Row(Modifier.padding(SpacingMd), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                        Column { Text(stringResource(R.string.me_trading_accounts), style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold); Text(stringResource(R.string.me_mt_desc), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                        Text(">", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
            item {
                Surface(onClick = onNavigateToAlerts, shape = MaterialTheme.shapes.small, color = MaterialTheme.colorScheme.surfaceVariant, modifier = Modifier.fillMaxWidth()) {
                    Row(Modifier.padding(SpacingMd), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                        Column { Text(stringResource(R.string.me_my_alerts), style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold); Text(stringResource(R.string.me_alerts_desc), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant) }
                        Text(">", color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
            item { Spacer(Modifier.height(SpacingLg)); OutlinedButton(onClick = { vm.logout { onLogout() } }, enabled = !state.loggingOut, modifier = Modifier.fillMaxWidth()) { if (state.loggingOut) CircularProgressIndicator(Modifier.size(20.dp)) else Text(stringResource(R.string.me_logout)) } }
        }
    }
}
