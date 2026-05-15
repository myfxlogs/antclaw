package com.antclaw.alfq.ui.alerts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.AlertRepository
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
class AlertViewModel @Inject constructor(
    private val alertRepo: AlertRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(AlertUiState())
    val uiState: StateFlow<AlertUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val items = alertRepo.listSubscriptions().map { s ->
                    AlertSubItem(s.id, s.pair, s.condition, s.threshold, s.type, s.active)
                }
                _uiState.value = AlertUiState(subscriptions = items)
            } catch (e: Exception) {
                _uiState.value = AlertUiState(error = e.message)
            }
        }
    }

    fun subscribe(type: String, pair: String, condition: String, threshold: String) {
        viewModelScope.launch {
            try { alertRepo.subscribe(type, pair, condition, threshold); load() } catch (_: Exception) {}
        }
    }

    fun unsubscribe(id: String) {
        viewModelScope.launch {
            try { alertRepo.unsubscribe(id); load() } catch (_: Exception) {}
        }
    }
}
