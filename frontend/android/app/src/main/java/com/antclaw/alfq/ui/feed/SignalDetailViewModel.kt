package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import antclaw.v1.Signals
import antclaw.v1.Price
import com.antclaw.alfq.data.rpc.RpcHelper
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
class SignalDetailViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(SignalDetailUiState())
    val uiState: StateFlow<SignalDetailUiState> = _uiState.asStateFlow()

    fun load(pair: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null, pair = pair)
            try {
                val unifiedReq = Signals.GetUnifiedRequest.newBuilder().setPair(pair).build()
                val signal = RpcHelper.unary(
                    "antclaw.v1.SignalsService/GetUnified", unifiedReq,
                    Signals.GetUnifiedRequest::class, Signals.GetUnifiedResponse::class)

                val priceReq = Price.GetPriceRequest.newBuilder().setPair(pair).setTimeframe("1D").setCount(1).build()
                val priceResp = RpcHelper.unary(
                    "antclaw.v1.PriceService/GetPrice", priceReq,
                    Price.GetPriceRequest::class, Price.GetPriceResponse::class)

                _uiState.value = SignalDetailUiState(
                    pair = pair,
                    direction = if (signal.hasSignal()) signal.signal.direction else "neutral",
                    confidence = if (signal.hasSignal()) ((signal.signal.confidence) * 100).toInt() else 0,
                    price = priceResp.current.ifEmpty { "--" },
                    loading = false
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(loading = false, error = e.message)
            }
        }
    }
}
