package com.antclaw.alfq.ui.notification

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.data.notification.ClientNotification
import com.antclaw.alfq.ui.theme.*
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NotificationCenterScreen(
    onBack: () -> Unit,
    onNotificationClick: (ClientNotification) -> Unit = {},
    vm: NotificationViewModel = viewModel(),
) {
    val state by vm.state.collectAsState()
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.notif_title), color = MaterialTheme.colorScheme.onSurface) },
                navigationIcon = { TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) } },
                actions = { TextButton(onClick = { vm.markAllRead() }) { Text(stringResource(R.string.notif_mark_all_read)) } },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
            )
        },
        containerColor = MaterialTheme.colorScheme.background,
    ) { padding ->
        if (state.error != null) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(stringResource(R.string.notif_load_failed), style = MaterialTheme.typography.bodyLarge)
                    TextButton(onClick = { vm.refresh() }) { Text(stringResource(R.string.common_retry)) }
                }
            }
        } else if (state.items.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Text(stringResource(R.string.notif_empty), color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        } else {
            LazyColumn(modifier = Modifier.fillMaxSize().padding(padding).padding(horizontal = 16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)) {
                items(state.items, key = { it.id }) { notif ->
                    NotificationCard(notif = notif, onClick = { if (!notif.isRead) vm.markRead(notif.id); onNotificationClick(notif) })
                }
            }
        }
    }
}

@Composable
private fun NotificationCard(notif: ClientNotification, onClick: () -> Unit) {
    val severityColor = when (notif.severity) { "critical" -> ErrorRed; "high" -> WarnOrange; "normal" -> InfoBlue; else -> TextMuted }
    Card(modifier = Modifier.fillMaxWidth().clickable(onClick = onClick),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = if (notif.isRead) BgSecondary else BgMain),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)) {
        Row(modifier = Modifier.fillMaxWidth()) {
            Box(modifier = Modifier.width(4.dp).fillMaxHeight().defaultMinSize(minHeight = 64.dp).background(severityColor))
            Column(modifier = Modifier.weight(1f).padding(12.dp)) {
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                    Text(text = notif.title, style = MaterialTheme.typography.bodyMedium.copy(fontWeight = FontWeight.SemiBold),
                        maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.weight(1f))
                    Spacer(Modifier.width(8.dp))
                    Text(text = fmtTime(notif.createdAt), style = MaterialTheme.typography.labelSmall, color = TextMuted)
                }
                Spacer(Modifier.height(4.dp))
                Text(text = notif.body, style = MaterialTheme.typography.bodySmall, color = TextSecondary, maxLines = 2, overflow = TextOverflow.Ellipsis)
                Spacer(Modifier.height(4.dp))
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    CategoryChip(notif.category)
                    if (notif.category == "signal") { Text(text = notif.severity, fontSize = 11.sp, color = severityColor) }
                }
            }
            if (!notif.isRead) { Box(modifier = Modifier.padding(12.dp).size(8.dp).background(color = MaterialTheme.colorScheme.primary, shape = RoundedCornerShape(4.dp))) }
        }
    }
}

@Composable
private fun CategoryChip(category: String) {
    val (bg, text) = when (category) {
        "alert" -> Pair(PrimaryLight.copy(alpha = 0.20f), PrimaryDark)
        "signal" -> Pair(SuccessGreen.copy(alpha = 0.20f), SuccessGreen)
        "digest" -> Pair(InfoBlue.copy(alpha = 0.20f), InfoBlue)
        else -> Pair(TextSecondary.copy(alpha = 0.20f), TextSecondary)
    }
    Box(modifier = Modifier.background(bg, RoundedCornerShape(4.dp)).padding(horizontal = 6.dp, vertical = 2.dp)) {
        Text(category, fontSize = 10.sp, color = text)
    }
}

private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm")
private fun fmtTime(instant: java.time.Instant): String {
    return try { instant.atZone(java.time.ZoneId.systemDefault()).format(timeFormatter) } catch (_: Exception) { "" }
}
