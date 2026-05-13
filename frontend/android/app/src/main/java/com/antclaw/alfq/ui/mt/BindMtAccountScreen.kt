package com.antclaw.alfq.ui.mt

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.ui.profile.MTAccountsViewModel

private val serverList = listOf("ICMarkets-Demo","ICMarkets-Live","RoboForex-Demo","RoboForex-Live","FBS-Demo","FBS-Live","Exness-Demo","Exness-Live")

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BindMtAccountScreen(onBack: () -> Unit, vm: MTAccountsViewModel = hiltViewModel()) {
    val state by vm.uiState.collectAsState()
    var step by remember { mutableIntStateOf(1) }
    var mtType by remember { mutableStateOf("MT4") }; var server by remember { mutableStateOf("") }
    var selectedServer by remember { mutableStateOf("") }; var account by remember { mutableStateOf("") }
    var investorPassword by remember { mutableStateOf("") }; var label by remember { mutableStateOf("") }
    var isDemo by remember { mutableStateOf(true) }

    LaunchedEffect(state.bindSuccess) { if (state.bindSuccess) { vm.clearBindResult(); onBack() } }
    if (state.bindError != null) AlertDialog(
        onDismissRequest = { vm.clearBindResult() }, title = { Text(stringResource(R.string.mt_bind_failed)) },
        text = { Text(state.bindError ?: "") }, confirmButton = { TextButton(onClick = { vm.clearBindResult() }) { Text(stringResource(R.string.common_ok)) } })

    Column(Modifier.fillMaxSize()) {
        TopAppBar(title = { Text(stringResource(R.string.bind_mt_title)) },
            navigationIcon = { TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) } },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background))

        // Step indicator (inline)
        Row(Modifier.fillMaxWidth().padding(horizontal = 32.dp, vertical = 16.dp), horizontalArrangement = Arrangement.Center, verticalAlignment = Alignment.CenterVertically) {
            listOf(R.string.bind_step_platform, R.string.bind_step_credentials, R.string.bind_step_confirm).forEachIndexed { i, resId ->
                if (i > 0) HorizontalDivider(Modifier.weight(1f).padding(horizontal = 8.dp))
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Surface(shape = MaterialTheme.shapes.extraLarge, color = if (i < step) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant, modifier = Modifier.size(32.dp)) {
                        Box(contentAlignment = Alignment.Center) {
                            if (i < step) Icon(Icons.Default.Check, null, tint = MaterialTheme.colorScheme.onPrimary, modifier = Modifier.size(16.dp))
                            else Text("${i+1}", style = MaterialTheme.typography.labelMedium, color = if (i < step) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                    Spacer(Modifier.height(4.dp))
                    Text(stringResource(resId), style = MaterialTheme.typography.labelSmall, color = if (i < step) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }
        }

        Card(Modifier.fillMaxWidth().padding(horizontal = 16.dp), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
            Column(Modifier.padding(16.dp)) {
                when (step) {
                    1 -> {
                        SectionTitle(R.string.bind_step1_title)
                        Spacer(Modifier.height(16.dp))
                        Text(stringResource(R.string.bind_platform_type), style = MaterialTheme.typography.bodySmall)
                        Spacer(Modifier.height(8.dp))
                        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                            FilterChip(selected = mtType == "MT4", onClick = { mtType = "MT4" }, label = { Text("MT4") })
                            FilterChip(selected = mtType == "MT5", onClick = { mtType = "MT5" }, label = { Text("MT5") })
                        }
                        Spacer(Modifier.height(16.dp))
                        Text(stringResource(R.string.bind_server), style = MaterialTheme.typography.bodySmall)
                        Spacer(Modifier.height(8.dp))
                        OutlinedTextField(server, { server = it; selectedServer = "" }, label = { Text(stringResource(R.string.bind_search_server)) }, singleLine = true, modifier = Modifier.fillMaxWidth())
                        Spacer(Modifier.height(8.dp))
                        val filtered = if (server.isBlank()) serverList else serverList.filter { it.contains(server, true) }
                        if (filtered.isNotEmpty()) LazyColumn(Modifier.heightIn(max = 160.dp)) {
                            items(filtered) { srv -> Surface(onClick = { selectedServer = srv; server = srv }, color = if (selectedServer == srv) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surface, modifier = Modifier.fillMaxWidth()) {
                                Row(Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                    if (selectedServer == srv) { Icon(Icons.Default.Check, null, tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(18.dp)); Spacer(Modifier.width(8.dp)) }
                                    Text(srv, style = MaterialTheme.typography.bodyMedium)
                                }
                            }}
                        }
                        Spacer(Modifier.height(16.dp))
                        Button({ step = 2 }, enabled = selectedServer.isNotBlank(), modifier = Modifier.fillMaxWidth()) { Text(stringResource(R.string.bind_next)) }
                    }
                    2 -> {
                        SectionTitle(R.string.bind_step2_title)
                        Spacer(Modifier.height(16.dp))
                        Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f))) {
                            Row(Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                Column { Text(selectedServer, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold); Text(mtType, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f)) }
                            }
                        }
                        Spacer(Modifier.height(16.dp))
                        OutlinedTextField(account, { account = it }, label = { Text(stringResource(R.string.bind_account)) }, placeholder = { Text(stringResource(R.string.bind_account_hint)) }, singleLine = true, modifier = Modifier.fillMaxWidth())
                        Spacer(Modifier.height(12.dp))
                        OutlinedTextField(investorPassword, { investorPassword = it }, label = { Text(stringResource(R.string.bind_investor_password)) }, placeholder = { Text(stringResource(R.string.bind_password_hint)) }, visualTransformation = PasswordVisualTransformation(), singleLine = true, modifier = Modifier.fillMaxWidth())
                        Spacer(Modifier.height(16.dp))
                        StepNavRow(onBack = { step = 1 }, onNext = { step = 3 }, nextEnabled = account.isNotBlank() && investorPassword.isNotBlank())
                    }
                    3 -> {
                        SectionTitle(R.string.bind_step3_title)
                        Spacer(Modifier.height(16.dp))
                        Card(Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
                            Column(Modifier.padding(16.dp)) {
                                listOf(R.string.bind_platform_type to mtType, R.string.bind_server to selectedServer, R.string.bind_account to account).forEach { (res, v) ->
                                    Row(Modifier.fillMaxWidth().padding(vertical = 4.dp), horizontalArrangement = Arrangement.SpaceBetween) {
                                        Text(stringResource(res), style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                                        Text(v, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold)
                                    }
                                }
                                Spacer(Modifier.height(12.dp))
                                OutlinedTextField(label, { label = it }, label = { Text(stringResource(R.string.bind_label)) }, placeholder = { Text(stringResource(R.string.bind_label_hint)) }, singleLine = true, modifier = Modifier.fillMaxWidth())
                                Spacer(Modifier.height(12.dp))
                                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                                    Text(stringResource(R.string.bind_account_type), style = MaterialTheme.typography.bodyMedium)
                                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                        FilterChip(selected = isDemo, onClick = { isDemo = true }, label = { Text(stringResource(R.string.bind_demo)) })
                                        FilterChip(selected = !isDemo, onClick = { isDemo = false }, label = { Text(stringResource(R.string.bind_live)) })
                                    }
                                }
                            }
                        }
                        Spacer(Modifier.height(16.dp))
                        if (state.binding) Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) { Row(verticalAlignment = Alignment.CenterVertically) { CircularProgressIndicator(Modifier.size(20.dp)); Spacer(Modifier.width(12.dp)); Text(stringResource(R.string.bind_binding), style = MaterialTheme.typography.bodyMedium) } }
                        else StepNavRow(onBack = { step = 2 }, onNext = {
                            if (mtType == "MT4") vm.bindMt4Account(selectedServer, account, investorPassword, label, isDemo)
                            else vm.bindMt5Account(selectedServer, account, investorPassword, label, isDemo)
                        }, nextText = stringResource(R.string.bind_confirm))
                    }
                }
            }
        }
    }
}

@Composable
private fun SectionTitle(resId: Int) = Text(stringResource(resId), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)

@Composable
private fun StepNavRow(onBack: () -> Unit, onNext: () -> Unit, nextEnabled: Boolean = true, nextText: String = stringResource(R.string.bind_next)) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        OutlinedButton(onBack, Modifier.weight(1f)) { Text(stringResource(R.string.bind_previous)) }
        Button(onNext, Modifier.weight(1f), enabled = nextEnabled) { Text(nextText) }
    }
}

private fun checkForIssues(): Unit {} // Prevent single-abstract-method warnings
