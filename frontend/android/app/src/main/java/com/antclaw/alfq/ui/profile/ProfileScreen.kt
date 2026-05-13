package com.antclaw.alfq.ui.profile

import androidx.compose.foundation.layout.*
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
fun ProfileScreen(
    userId: String = "me",
    viewModel: ProfileViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state by viewModel.uiState.collectAsState()
    LaunchedEffect(userId) { viewModel.load(userId) }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) }
            Text(stringResource(R.string.profile_title), style = MaterialTheme.typography.titleLarge)
        }
        if (state.loading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else {
            Column(modifier = Modifier.fillMaxSize().padding(16.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Text(state.displayName, style = MaterialTheme.typography.headlineMedium)
                val tierLabel = when (state.tier) { "verified" -> stringResource(R.string.tier_verified); "elite" -> stringResource(R.string.tier_elite); else -> "" }
                if (tierLabel.isNotEmpty()) Text(tierLabel, color = MaterialTheme.colorScheme.primary)
                if (state.bio.isNotEmpty()) Text(state.bio, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
                Spacer(modifier = Modifier.height(16.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(32.dp)) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("${state.followerCount}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                        Text(stringResource(R.string.profile_followers), style = MaterialTheme.typography.bodySmall)
                    }
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text("${state.followingCount}", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                        Text(stringResource(R.string.profile_following), style = MaterialTheme.typography.bodySmall)
                    }
                }
                Spacer(modifier = Modifier.height(16.dp))
                Button(onClick = { viewModel.toggleFollow() }) {
                    Text(if (state.isFollowing) stringResource(R.string.profile_unfollow) else stringResource(R.string.profile_follow))
                }
                Spacer(modifier = Modifier.height(24.dp))
                Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Column(modifier = Modifier.padding(16.dp)) {
                        Text(stringResource(R.string.profile_stats_title), style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                        Spacer(modifier = Modifier.height(12.dp))
                        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                            StatCell(stringResource(R.string.profile_win_rate), "${(state.winRate * 100).toInt()}%")
                            StatCell(stringResource(R.string.profile_profit_factor), String.format("%.2f", state.profitFactor))
                            StatCell(stringResource(R.string.profile_sharpe), String.format("%.2f", state.sharpeRatio))
                            StatCell(stringResource(R.string.profile_total_trades), "${state.totalTrades}")
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun StatCell(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BullGreen)
        Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
    }
}
