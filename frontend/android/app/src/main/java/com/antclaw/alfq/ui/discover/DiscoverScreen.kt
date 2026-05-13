package com.antclaw.alfq.ui.discover

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.ui.theme.SpacingMd
import com.antclaw.alfq.ui.theme.SpacingSm

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DiscoverScreen(viewModel: DiscoverViewModel = hiltViewModel()) {
    val state by viewModel.uiState.collectAsState()
    var searchQuery by remember { mutableStateOf("") }

    Column(modifier = Modifier.fillMaxSize()) {
        // Search Bar
        OutlinedTextField(
            value = searchQuery,
            onValueChange = { searchQuery = it },
            placeholder = { Text("搜索交易员、圈子、信号...") },
            leadingIcon = { Icon(Icons.Default.Search, contentDescription = "搜索") },
            modifier = Modifier
                .fillMaxWidth()
                .padding(SpacingMd),
            singleLine = true
        )

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            contentPadding = PaddingValues(horizontal = SpacingMd, vertical = SpacingMd),
            verticalArrangement = Arrangement.spacedBy(SpacingMd)
        ) {
            item {
                Text("热门交易员", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(SpacingSm))
            }
            items(state.traders.take(3)) { name ->
                Surface(
                    onClick = { /* TODO: Navigate to trader profile */ },
                    shape = MaterialTheme.shapes.small,
                    color = MaterialTheme.colorScheme.surfaceVariant,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Row(
                        Modifier.padding(SpacingMd),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(name, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
                        OutlinedButton(onClick = { }) { Text("关注") }
                    }
                }
            }

            item {
                Spacer(modifier = Modifier.height(SpacingMd))
                Text("热门圈子", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                Spacer(modifier = Modifier.height(SpacingSm))
            }
            items(state.circles) { c ->
                Surface(
                    onClick = { /* TODO: Navigate to circle */ },
                    shape = MaterialTheme.shapes.small,
                    color = MaterialTheme.colorScheme.surfaceVariant,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(Modifier.padding(SpacingMd)) {
                        Text(c.name, style = MaterialTheme.typography.bodyLarge, fontWeight = FontWeight.Bold)
                        Text("${c.symbol} · ${c.memberCount} 成员", style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }
        }
    }
}
