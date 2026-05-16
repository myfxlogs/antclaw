package com.antclaw.alfq.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.AuthRepository
import com.antclaw.alfq.data.repository.AuthSessionResult
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class LoginUiState(
    val email: String = "",
    val password: String = "",
    val loading: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepo: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    /** 从持久存储恢复 token 和 userId，尝试自动登录。 */
    fun autoLogin(onSuccess: (AuthSessionResult) -> Unit) {
        viewModelScope.launch {
            val token = authRepo.restoreToken()
            if (token != null) {
                onSuccess(
                    AuthSessionResult(
                        userId = authRepo.restoredUserId() ?: "",
                        accessToken = token,
                        refreshToken = "",
                    )
                )
            }
        }
    }

    fun onEmailChange(email: String) {
        _uiState.value = _uiState.value.copy(email = email, error = null)
    }

    fun onPasswordChange(password: String) {
        _uiState.value = _uiState.value.copy(password = password, error = null)
    }

    fun login(onSuccess: (AuthSessionResult) -> Unit) {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.value = state.copy(loading = true, error = null)
            authRepo.login(state.email, state.password)
                .onSuccess { result ->
                    _uiState.value = _uiState.value.copy(loading = false)
                    onSuccess(result)
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
