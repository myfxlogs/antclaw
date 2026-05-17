package com.antclaw.alfq.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.error.toUserErrorRes
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
    /** 用户可读错误文案的 string resource ID，null 表示无错误 */
    val errorRes: Int? = null,
)

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepo: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(LoginUiState())
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()

    fun onEmailChange(email: String) {
        _uiState.value = _uiState.value.copy(email = email, errorRes = null)
    }

    fun onPasswordChange(password: String) {
        _uiState.value = _uiState.value.copy(password = password, errorRes = null)
    }

    fun login(onSuccess: (AuthSessionResult) -> Unit) {
        val state = _uiState.value
        viewModelScope.launch {
            _uiState.value = state.copy(loading = true, errorRes = null)
            authRepo.login(state.email, state.password)
                .onSuccess { result ->
                    _uiState.value = _uiState.value.copy(loading = false)
                    onSuccess(result)
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        loading = false,
                        errorRes = e.toUserErrorRes(),
                    )
                }
        }
    }
}
