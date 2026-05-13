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

private val serverList = listOf(
    "ICMarkets-Demo", "ICMarkets-Live", "RoboForex-Demo", "RoboForex-Live",
    "FBS-Demo", "FBS-Live", "Exness-Demo", "Exness-Live",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BindMtAccountScreen(onBack: () -> Unit, vm: MTAccountsViewModel = hiltViewModel()) {
    val state by vm.uiState.collectAsState()
    var step by remember { mutableIntStateOf(1) }
    var mtType by remember { mutableStateOf("MT4") }
    var server by remember { mutableStateOf("") }
    var selectedServer by remember { mutableStateOf("") }
    var account by remember { mutableStateOf("") }
    var investorPassword by remember { mutableStateOf("") }
    var label by remember { mutableStateOf("") }
    var isDemo by remember { mutableStateOf(true) }

    LaunchedEffect(state.bindSuccess) { if (state.bindSuccess) { vm.clearBindResult(); onBack() } }
    if (state.bindError != null) {
        AlertDialog(
            onDismissRequest = { vm.clearBindResult() },
            title = { Text(stringResource(R.string.mt_bind_failed)) },
            text = { Text(state.bindError ?: "") },
            confirmButton = { TextButton(onClick = { vm.clearBindResult() }) { Text(stringResource(R.string.common_ok)) } }
        )
    }

    Column(modifier = Modifier.fillMaxSize()) {
        TopAppBar(
            title = { Text(stringResource(R.string.bind_mt_title)) },
            navigationIcon = { TextButton(onClick = onBack) { Text(stringResource(R.string.common_back)) } },
            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background)
        )
        Row(modifier = Modifier.fillMaxWidth().padding(horizontal = 32.dp, vertical = 16.dp),
            horizontalArrangement = Arrangement.Center, verticalAlignment = Alignment.CenterVertically) {
            StepDot(1, step, stringResource(R.string.bind_step_platform))
            HorizontalDivider(modifier = Modifier.weight(1f).padding(horizontal = 8.dp))
            StepDot(2, step, stringResource(R.string.bind_step_credentials))
            HorizontalDivider(modifier = Modifier.weight(1f).padding(horizontal = 8.dp))
            StepDot(3, step, stringResource(R.string.bind_step_confirm))
        }

        Card(modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
            when (step) {
                1 -> Step1PlatformServer(mtType = mtType, onMtTypeChange = { mtType = it },
                    server = server, onServerChange = { server = it },
                    selectedServer = selectedServer, onServerSelect = { selectedServer = it },
                    onNext = { step = 2 })
                2 -> Step2Credentials(selectedServer = selectedServer, mtType = mtType,
                    account = account, onAccountChange = { account = it },
                    investorPassword = investorPassword, onPasswordChange = { investorPassword = it },
                    onBack = { step = 1 }, onNext = { step = 3 })
                3 -> Step3Confirm(mtType = mtType, selectedServer = selectedServer, account = account,
                    label = label, onLabelChange = { label = it },
                    isDemo = isDemo, onIsDemoChange = { isDemo = it },
                    binding = state.binding, onBack = { step = 2 },
                    onConfirm = {
                        if (mtType == "MT4") vm.bindMt4Account(selectedServer, account, investorPassword, label, isDemo)
                        else vm.bindMt5Account(selectedServer, account, investorPassword, label, isDemo)
                    })
            }
        }
    }
}

@Composable
private fun StepDot(num: Int, current: Int, label: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Surface(shape = MaterialTheme.shapes.extraLarge,
            color = if (num <= current) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant,
            modifier = Modifier.size(32.dp)) {
            Box(contentAlignment = Alignment.Center) {
                if (num < current) Icon(Icons.Default.Check, contentDescription = null,
                    tint = MaterialTheme.colorScheme.onPrimary, modifier = Modifier.size(16.dp))
                else Text("$num", style = MaterialTheme.typography.labelMedium,
                    color = if (num <= current) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
        Spacer(modifier = Modifier.height(4.dp))
        Text(label, style = MaterialTheme.typography.labelSmall,
            color = if (num <= current) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

@Composable
private fun Step1PlatformServer(mtType: String, onMtTypeChange: (String) -> Unit,
    server: String, onServerChange: (String) -> Unit,
    selectedServer: String, onServerSelect: (String) -> Unit, onNext: () -> Unit) {
    Column(modifier = Modifier.padding(16.dp)) {
        Text(stringResource(R.string.bind_step1_title), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Spacer(modifier = Modifier.height(16.dp))
        Text(stringResource(R.string.bind_platform_type), style = MaterialTheme.typography.bodySmall)
        Spacer(modifier = Modifier.height(8.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            FilterChip(selected = mtType == "MT4", onClick = { onMtTypeChange("MT4") }, label = { Text("MT4") })
            FilterChip(selected = mtType == "MT5", onClick = { onMtTypeChange("MT5") }, label = { Text("MT5") })
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(stringResource(R.string.bind_server), style = MaterialTheme.typography.bodySmall)
        Spacer(modifier = Modifier.height(8.dp))
        OutlinedTextField(value = server, onValueChange = { onServerChange(it); onServerSelect("") },
            label = { Text(stringResource(R.string.bind_search_server)) }, modifier = Modifier.fillMaxWidth(), singleLine = true)
        Spacer(modifier = Modifier.height(8.dp))
        val filtered = if (server.isBlank()) serverList else serverList.filter { it.contains(server, ignoreCase = true) }
        if (filtered.isNotEmpty()) {
            LazyColumn(modifier = Modifier.heightIn(max = 160.dp)) {
                items(filtered) { srv ->
                    Surface(onClick = { onServerSelect(srv); onServerChange(srv) },
                        color = if (selectedServer == srv) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surface,
                        modifier = Modifier.fillMaxWidth()) {
                        Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            if (selectedServer == srv) {
                                Icon(Icons.Default.Check, contentDescription = null, tint = MaterialTheme.colorScheme.primary, modifier = Modifier.size(18.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                            }
                            Text(srv, style = MaterialTheme.typography.bodyMedium)
                        }
                    }
                }
            }
        }
        Spacer(modifier = Modifier.height(16.dp))
        Button(onClick = onNext, enabled = selectedServer.isNotBlank(), modifier = Modifier.fillMaxWidth()) {
            Text(stringResource(R.string.bind_next))
        }
    }
}

@Composable
private fun Step2Credentials(selectedServer: String, mtType: String,
    account: String, onAccountChange: (String) -> Unit,
    investorPassword: String, onPasswordChange: (String) -> Unit,
    onBack: () -> Unit, onNext: () -> Unit) {
    Column(modifier = Modifier.padding(16.dp)) {
        Text(stringResource(R.string.bind_step2_title), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Spacer(modifier = Modifier.height(16.dp))
        Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.3f))) {
            Row(modifier = Modifier.padding(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Column {
                    Text(selectedServer, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold)
                    Text(mtType, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
                }
            }
        }
        Spacer(modifier = Modifier.height(16.dp))
        OutlinedTextField(value = account, onValueChange = onAccountChange,
            label = { Text(stringResource(R.string.bind_account)) },
            placeholder = { Text(stringResource(R.string.bind_account_hint)) },
            modifier = Modifier.fillMaxWidth(), singleLine = true)
        Spacer(modifier = Modifier.height(12.dp))
        OutlinedTextField(value = investorPassword, onValueChange = onPasswordChange,
            label = { Text(stringResource(R.string.bind_investor_password)) },
            placeholder = { Text(stringResource(R.string.bind_password_hint)) },
            visualTransformation = PasswordVisualTransformation(),
            modifier = Modifier.fillMaxWidth(), singleLine = true)
        Spacer(modifier = Modifier.height(16.dp))
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            OutlinedButton(onClick = onBack, modifier = Modifier.weight(1f)) { Text(stringResource(R.string.bind_previous)) }
            Button(onClick = onNext, enabled = account.isNotBlank() && investorPassword.isNotBlank(), modifier = Modifier.weight(1f)) {
                Text(stringResource(R.string.bind_next))
            }
        }
    }
}

@Composable
private fun Step3Confirm(mtType: String, selectedServer: String, account: String,
    label: String, onLabelChange: (String) -> Unit,
    isDemo: Boolean, onIsDemoChange: (Boolean) -> Unit,
    binding: Boolean, onBack: () -> Unit, onConfirm: () -> Unit) {
    Column(modifier = Modifier.padding(16.dp)) {
        Text(stringResource(R.string.bind_step3_title), style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        Spacer(modifier = Modifier.height(16.dp))
        Card(modifier = Modifier.fillMaxWidth(), colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
            Column(modifier = Modifier.padding(16.dp)) {
                PreviewRow(stringResource(R.string.bind_platform_type).replace(":\u0020".toRegex(), ""), mtType)
                PreviewRow(stringResource(R.string.bind_server).replace(":\u0020".toRegex(), ""), selectedServer)
                PreviewRow(stringResource(R.string.bind_account).replace(":\u0020".toRegex(), ""), account)
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(value = label, onValueChange = onLabelChange,
                    label = { Text(stringResource(R.string.bind_label)) },
                    placeholder = { Text(stringResource(R.string.bind_label_hint)) },
                    modifier = Modifier.fillMaxWidth(), singleLine = true)
                Spacer(modifier = Modifier.height(12.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                    Text(stringResource(R.string.bind_account_type), style = MaterialTheme.typography.bodyMedium)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        FilterChip(selected = isDemo, onClick = { onIsDemoChange(true) }, label = { Text(stringResource(R.string.bind_demo)) })
                        FilterChip(selected = !isDemo, onClick = { onIsDemoChange(false) }, label = { Text(stringResource(R.string.bind_live)) })
                    }
                }
            }
        }
        Spacer(modifier = Modifier.height(16.dp))
        if (binding) {
            Box(modifier = Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    CircularProgressIndicator(modifier = Modifier.size(20.dp))
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(stringResource(R.string.bind_binding), style = MaterialTheme.typography.bodyMedium)
                }
            }
        } else {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                OutlinedButton(onClick = onBack, modifier = Modifier.weight(1f)) { Text(stringResource(R.string.bind_previous)) }
                Button(onClick = onConfirm, modifier = Modifier.weight(1f),
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.primary)) { Text(stringResource(R.string.bind_confirm)) }
            }
        }
    }
}

@Composable
private fun PreviewRow(label: String, value: String) {
    Row(modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp), horizontalArrangement = Arrangement.SpaceBetween) {
        Text(label, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
        Text(value, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Bold)
    }
}
