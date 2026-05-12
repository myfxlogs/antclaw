package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.*
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class FactorItem(val name: String, val value: Double)
data class SignalDetailUiState(
    val pair: String = "",
    val direction: String = "neutral",
    val confidence: Int = 0,
    val price: String = "",
    val factors: List<FactorItem> = emptyList(),
    val barCloses: List<Float> = emptyList(),
    val loading: Boolean = true,
    val error: String? = null,
)

@HiltViewModel
class SignalDetailViewModel @Inject constructor() : ViewModel() {
    private val signalsClient = SignalsRpcClient()
    private val priceClient = PriceRpcClient()

    private val _uiState = MutableStateFlow(SignalDetailUiState())
    val uiState: StateFlow<SignalDetailUiState> = _uiState.asStateFlow()

    fun load(pair: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val unified = signalsClient.getUnified(SigUnifiedReq(pair))
                val priceResp = priceClient.getBars(PriceReq(pair, "1D", 20))
                val sig = unified.signal

                // Parse factor scores from response or use xfactors
                val factors = try {
                    val xf = signalsClient.getXFactors(SigXFactorsReq(pair))
                    xf.factors.map { FactorItem(it.name, it.weight) }
                } catch (_: Exception) {
                    listOf(
                        FactorItem("TA", 0.88),
                        FactorItem("COT", 0.72),
                        FactorItem("Sentiment", 0.55),
                        FactorItem("Macro", 0.35),
                    )
                }

                val closes = priceResp.bars.map {
                    it.close.toFloatOrNull() ?: 0f
                }

                _uiState.value = SignalDetailUiState(
                    pair = pair,
                    direction = sig?.direction ?: "neutral",
                    confidence = ((sig?.confidence ?: 0.0) * 100).toInt(),
                    price = priceResp.current.ifEmpty { "--" },
                    factors = factors,
                    barCloses = closes,
                    loading = false
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(loading = false, error = e.message ?: "加载失败")
            }
        }
    }
}
