package com.antclaw.alfq.ui.alerts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.AlertSub
import com.antclaw.alfq.data.rpc.AlertSubReq
import com.antclaw.alfq.data.rpc.AlertsRpcClient
import com.antclaw.alfq.data.rpc.CreateAlertReq
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AlertUiState(
    val subscriptions: List<AlertSub> = emptyList(),
    val loading: Boolean = true,
)

@HiltViewModel
class AlertViewModel @Inject constructor(
    private val client: AlertsRpcClient,
) : ViewModel() {

    private val _uiState = MutableStateFlow(AlertUiState())
    val uiState: StateFlow<AlertUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true)
            try {
                val resp = client.listSubscriptions()
                _uiState.value = AlertUiState(subscriptions = resp.subscriptions, loading = false)
            } catch (_: Exception) {
                _uiState.value = AlertUiState(loading = false)
            }
        }
    }

    fun subscribe(type: String, pair: String, condition: String, threshold: String) {
        viewModelScope.launch {
            try {
                client.subscribe(AlertSubReq(type, pair, condition, threshold))
                load()
            } catch (_: Exception) { }
        }
    }

    fun unsubscribe(id: String) {
        viewModelScope.launch {
            try {
                client.unsubscribe(id)
                load()
            } catch (_: Exception) { }
        }
    }
}
