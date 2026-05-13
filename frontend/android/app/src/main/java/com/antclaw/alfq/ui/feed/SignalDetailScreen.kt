package com.antclaw.alfq.ui.feed

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
import com.antclaw.alfq.ui.theme.BearRed

@Composable
fun SignalDetailScreen(
    pair: String,
    viewModel: SignalDetailViewModel = hiltViewModel(),
    onBack: () -> Unit,
) {
    val state by viewModel.uiState.collectAsState()
    LaunchedEffect(pair) { viewModel.load(pair) }

    Column(modifier = Modifier.fillMaxSize()) {
        Row(modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) }
            Text(pair, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        }
        if (state.loading) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator() }
        } else if (state.error != null) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(state.error!!, color = MaterialTheme.colorScheme.error)
                    Spacer(modifier = Modifier.height(8.dp))
                    TextButton(onClick = { viewModel.load(pair) }) { Text(stringResource(R.string.common_retry)) }
                }
            }
        } else {
            Column(Modifier.fillMaxSize().padding(16.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                val color = when (state.direction) {
                    "bullish" -> BullGreen; "bearish" -> BearRed
                    else -> MaterialTheme.colorScheme.onSurface
                }
                Text(state.direction, style = MaterialTheme.typography.headlineSmall, color = color)
                Text(stringResource(R.string.signal_detail_confidence, state.confidence), style = MaterialTheme.typography.bodyLarge)
                Text(stringResource(R.string.signal_detail_price, state.price), style = MaterialTheme.typography.bodyMedium)
            }
        }
    }
}
