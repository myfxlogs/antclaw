package com.antclaw.alfq.ui.login

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R

@Composable
fun RegisterScreen(
    onBack: () -> Unit,
    onRegisterSuccess: (String) -> Unit,
    vm: RegisterViewModel = hiltViewModel()
) {
    val state by vm.state.collectAsState()
    LaunchedEffect(state.registerSuccess) { if (state.registerSuccess) { onRegisterSuccess(state.accessToken) } }

    Column(modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.Center) {
        Text(text = stringResource(R.string.register_title), style = MaterialTheme.typography.headlineMedium)
        Spacer(modifier = Modifier.height(32.dp))

        OutlinedTextField(value = state.email, onValueChange = { vm.updateEmail(it) },
            label = { Text(stringResource(R.string.register_email)) },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
            singleLine = true, modifier = Modifier.fillMaxWidth())
        Spacer(modifier = Modifier.height(12.dp))

        OutlinedTextField(value = state.password, onValueChange = { vm.updatePassword(it) },
            label = { Text(stringResource(R.string.register_password)) },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Text, imeAction = ImeAction.Done),
            singleLine = true, modifier = Modifier.fillMaxWidth())
        Spacer(modifier = Modifier.height(24.dp))

        Button(onClick = { vm.register() },
            enabled = state.email.isNotBlank() && state.password.isNotBlank() && !state.loading,
            modifier = Modifier.fillMaxWidth().height(50.dp)) {
            if (state.loading) { CircularProgressIndicator(modifier = Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary) }
            else { Text(stringResource(R.string.register_button)) }
        }

        state.error?.let { error ->
            Spacer(modifier = Modifier.height(12.dp))
            Text(text = error, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
        }
        Spacer(modifier = Modifier.height(16.dp))
        TextButton(onClick = onBack) { Text(stringResource(R.string.register_has_account)) }
    }
}
