package com.antclaw.alfq.ui.login

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.AuthRepository
import com.antclaw.alfq.data.repository.AuthSessionResult
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class RegisterUiState(
    val email: String = "",
    val password: String = "",
    val loading: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class RegisterViewModel @Inject constructor(
    private val authRepo: AuthRepository
) : ViewModel() {

    private val _state = MutableStateFlow(RegisterUiState())
    val state: StateFlow<RegisterUiState> = _state

    fun updateEmail(email: String) = _state.update { it.copy(email = email, error = null) }
    fun updatePassword(password: String) = _state.update { it.copy(password = password, error = null) }

    fun register(onSuccess: (AuthSessionResult) -> Unit) {
        val s = _state.value
        if (s.email.isBlank() || s.password.isBlank()) return

        viewModelScope.launch {
            _state.update { it.copy(loading = true, error = null) }
            authRepo.register(s.email, s.password).fold(
                onSuccess = { result ->
                    _state.update { it.copy(loading = false) }
                    onSuccess(result)
                },
                onFailure = { e ->
                    _state.update {
                        it.copy(
                            loading = false,
                            error = e.message ?: "注册失败，请重试"
                        )
                    }
                }
            )
        }
    }
}
