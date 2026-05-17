package com.antclaw.alfq.data.sse

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.TokenManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SseDebugState(
    val baseUrl: String = "",
    val hasToken: Boolean = false,
    val sseState: String = "UNKNOWN",
    val lastError: String? = null,
    val lastDeviceReportResult: String? = null,
)

@HiltViewModel
class SseDebugViewModel @Inject constructor(
    private val sseManager: SseManager,
    private val tokenManager: TokenManager,
) : ViewModel() {

    private val _state = MutableStateFlow(SseDebugState())
    val state: StateFlow<SseDebugState> = _state.asStateFlow()

    init {
        viewModelScope.launch {
            sseManager.connectionState.collect { connState ->
                _state.value = _state.value.copy(
                    hasToken = tokenManager.getToken() != null,
                    sseState = connState.name,
                )
            }
        }
    }

    fun refresh() {
        _state.value = _state.value.copy(
            hasToken = tokenManager.getToken() != null,
        )
    }
}
