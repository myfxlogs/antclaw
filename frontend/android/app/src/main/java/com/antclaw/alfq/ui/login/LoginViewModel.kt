package com.antclaw.alfq.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.AuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val email: String = "admin@antclaw.dev",
    val password: String = "Admin@12345",
    val loading: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepo: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    init {
        // 尝试恢复登录态
        viewModelScope.launch {
            val token = authRepo.restoreToken()
            if (token != null) {
                // 已有有效 token，跳过登录
            }
        }
    }

    fun onEmailChange(email: String) {
        _uiState.value = _uiState.value.copy(email = email, error = null)
    }

    fun onPasswordChange(password: String) {
        _uiState.value = _uiState.value.copy(password = password, error = null)
    }

    fun login(onSuccess: (String) -> Unit) {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.value = state.copy(loading = true, error = null)
            authRepo.login(state.email, state.password)
                .onSuccess { token ->
                    _uiState.value = _uiState.value.copy(loading = false)
                    onSuccess(token)
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        loading = false,
                        error = e.message ?: "登录失败"
                    )
                }
        }
    }
}
