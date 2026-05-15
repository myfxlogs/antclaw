package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.repository.SignalRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SignalDetailUiState(
    val pair: String = "", val direction: String = "neutral",
    val confidence: Int = 0, val price: String = "",
    val loading: Boolean = true, val error: String? = null
)

@HiltViewModel
class SignalDetailViewModel @Inject constructor(
    private val signalRepo: SignalRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(SignalDetailUiState())
    val uiState: StateFlow<SignalDetailUiState> = _uiState.asStateFlow()

    fun load(pair: String) {
        viewModelScope.launch {
            _uiState.value = SignalDetailUiState(pair = pair, loading = true)
            try {
                val d = signalRepo.getDetail(pair)
                _uiState.value = SignalDetailUiState(pair = pair, direction = d.direction, confidence = d.confidence, price = d.price)
            } catch (e: Exception) {
                _uiState.value = SignalDetailUiState(pair = pair, error = e.message)
            }
        }
    }
}
