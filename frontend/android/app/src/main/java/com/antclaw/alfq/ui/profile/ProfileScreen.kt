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
import com.antclaw.alfq.ui.components.StatCell
import com.antclaw.alfq.ui.components.TraderStatRow

@Composable
fun ProfileScreen(userId: String = "me", viewModel: ProfileViewModel = hiltViewModel(), onBack: () -> Unit) {
    val state by viewModel.uiState.collectAsState()
    LaunchedEffect(userId) { viewModel.load(userId) }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(Modifier.fillMaxWidth().padding(16.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) }
            Text(stringResource(R.string.profile_title), style = MaterialTheme.typography.titleLarge)
        }
        if (state.loading) Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        else Column(Modifier.fillMaxSize().padding(16.dp), horizontalAlignment = Alignment.CenterHorizontally) {
            Text(state.displayName, style = MaterialTheme.typography.headlineMedium)
            val tierLabel = when (state.tier) { "verified" -> stringResource(R.string.tier_verified); "elite" -> stringResource(R.string.tier_elite); else -> "" }
            if (tierLabel.isNotEmpty()) Text(tierLabel, color = MaterialTheme.colorScheme.primary)
            if (state.bio.isNotEmpty()) Text(state.bio, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f))
            Spacer(Modifier.height(16.dp))
            TraderStatRow(
                followingLabel = stringResource(R.string.profile_following), followingCount = state.followingCount,
                followersLabel = stringResource(R.string.profile_followers), followerCount = state.followerCount)
            Spacer(Modifier.height(16.dp))
            Button(onClick = { viewModel.toggleFollow() }) { Text(if (state.isFollowing) stringResource(R.string.profile_unfollow) else stringResource(R.string.profile_follow)) }
            Spacer(Modifier.height(24.dp))
            Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                Column(Modifier.padding(16.dp)) {
                    Text(stringResource(R.string.profile_stats_title), style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                    Spacer(Modifier.height(12.dp))
                    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
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
