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
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R

@Composable
fun LoginScreen(
    onLoginSuccess: (String) -> Unit,
    onRegisterClick: () -> Unit = {},
    onForgotPasswordClick: () -> Unit = {},
    viewModel: LoginViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    LaunchedEffect(Unit) { viewModel.autoLogin { token -> onLoginSuccess(token) } }

    Column(modifier = Modifier.fillMaxSize().padding(32.dp),
        verticalArrangement = Arrangement.Center, horizontalAlignment = Alignment.CenterHorizontally) {
        Text(stringResource(R.string.login_title), style = MaterialTheme.typography.displayLarge,
            color = MaterialTheme.colorScheme.primary)
        Spacer(modifier = Modifier.height(4.dp))
        Text(stringResource(R.string.app_tagline), style = MaterialTheme.typography.bodyLarge,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
        Spacer(modifier = Modifier.height(48.dp))

        OutlinedTextField(value = state.email, onValueChange = viewModel::onEmailChange,
            label = { Text(stringResource(R.string.login_email)) },
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Email, imeAction = ImeAction.Next),
            singleLine = true, modifier = Modifier.fillMaxWidth())
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedTextField(value = state.password, onValueChange = viewModel::onPasswordChange,
            label = { Text(stringResource(R.string.login_password)) },
            visualTransformation = PasswordVisualTransformation(),
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Done),
            singleLine = true, modifier = Modifier.fillMaxWidth())
        Spacer(modifier = Modifier.height(24.dp))

        if (state.error != null) {
            Text(state.error!!, color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall, modifier = Modifier.padding(bottom = 8.dp))
        }

        Button(onClick = { viewModel.login { token -> onLoginSuccess(token) } },
            enabled = !state.loading && state.email.isNotBlank() && state.password.isNotBlank(),
            modifier = Modifier.fillMaxWidth().height(52.dp)) {
            if (state.loading) {
                CircularProgressIndicator(modifier = Modifier.size(24.dp), color = MaterialTheme.colorScheme.onPrimary)
            } else {
                Text(stringResource(R.string.login_button), style = MaterialTheme.typography.titleMedium)
            }
        }

        Spacer(modifier = Modifier.height(12.dp))
        Row {
            TextButton(onClick = onRegisterClick) { Text(stringResource(R.string.login_register), style = MaterialTheme.typography.bodySmall) }
            Text(" \u00b7 ", color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f))
            TextButton(onClick = onForgotPasswordClick) { Text(stringResource(R.string.login_forgot_password), style = MaterialTheme.typography.bodySmall) }
        }
    }
}
