package com.antclaw.alfq.ui.alerts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.Alerts
import com.antclaw.alfq.data.rpc.RpcHelper
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AlertUiState(
    val loading: Boolean = false, val error: String? = null,
    val subscriptions: List<AlertSubItem> = emptyList()
)
data class AlertSubItem(
    val subscriptionId: String, val pair: String, val condition: String,
    val threshold: String, val alertType: String, val active: Boolean
)

@HiltViewModel
class AlertViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(AlertUiState())
    val uiState: StateFlow<AlertUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val resp = RpcHelper.unary(
                    "antclaw.v1.AlertService/ListSubscriptions",
                    Alerts.ListSubscriptionsRequest.getDefaultInstance(),
                    Alerts.ListSubscriptionsRequest::class,
                    Alerts.ListSubscriptionsResponse::class)
                val items = resp.subscriptionsList.map { s ->
                    AlertSubItem(s.subscriptionId, s.pair, s.condition, s.threshold, s.alertType, s.active)
                }
                _uiState.value = _uiState.value.copy(loading = false, subscriptions = items)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(loading = false, error = e.message)
            }
        }
    }

    fun subscribe(type: String, pair: String, condition: String, threshold: String) {
        viewModelScope.launch {
            try {
                val req = Alerts.SubscribeRequest.newBuilder()
                    .setAlertType(type).setPair(pair).setCondition(condition).setThreshold(threshold).build()
                RpcHelper.unary("antclaw.v1.AlertService/Subscribe", req,
                    Alerts.SubscribeRequest::class, Alerts.SubscribeResponse::class)
                load()
            } catch (_: Exception) {}
        }
    }

    fun unsubscribe(id: String) {
        viewModelScope.launch {
            try {
                val req = Alerts.UnsubscribeRequest.newBuilder().setSubscriptionId(id).build()
                RpcHelper.unary("antclaw.v1.AlertService/Unsubscribe", req,
                    Alerts.UnsubscribeRequest::class, Alerts.UnsubscribeResponse::class)
                load()
            } catch (_: Exception) {}
        }
    }
}
