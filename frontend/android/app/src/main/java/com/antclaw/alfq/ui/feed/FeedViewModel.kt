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

data class SignalBarItem(
    val pair: String, val direction: String = "neutral",
    val confidence: Int = 0, val price: String = "--"
)
data class FeedCard(
    val id: String, val author: String = "", val pair: String? = null,
    val direction: String = "neutral", val confidence: Int = 0,
    val content: String = "", val timeAgo: String = ""
)
data class FeedUiState(
    val signalBar: List<SignalBarItem> = emptyList(),
    val cards: List<FeedCard> = emptyList(),
    val loading: Boolean = false, val error: String? = null
)

@HiltViewModel
class FeedViewModel @Inject constructor() : ViewModel() {
    private val _uiState = MutableStateFlow(FeedUiState())
    val uiState: StateFlow<FeedUiState> = _uiState.asStateFlow()
    private val defaultPairs = listOf("EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "XAUUSD", "BTCUSD")

    private val authorSystem = "\u7cfb\u7edf\u4fe1\u53f7"
    private val timeJustNow = "\u521a\u521a"

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                _uiState.value = FeedUiState(signalBar = loadSignalBar(), cards = loadFeedCards())
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(error = e.message)
            }
        }
    }

    private suspend fun loadSignalBar(): List<SignalBarItem> = defaultPairs.map { pair ->
        try {
            val biasReq = Signals.GetBiasRequest.newBuilder().setPair(pair).setTimeframe("1D").build()
            val biasResp = RpcHelper.unary(
                "antclaw.v1.SignalsService/GetBias", biasReq,
                Signals.GetBiasRequest::class, Signals.GetBiasResponse::class)
            val bias = biasResp.biasesList.firstOrNull()
            val priceReq = Price.GetPriceRequest.newBuilder().setPair(pair).setTimeframe("1D").setCount(1).build()
            val priceResp = RpcHelper.unary(
                "antclaw.v1.PriceService/GetPrice", priceReq,
                Price.GetPriceRequest::class, Price.GetPriceResponse::class)
            SignalBarItem(pair, bias?.direction ?: "neutral",
                ((bias?.confidence ?: 0.0) * 100).toInt(), priceResp.current.ifEmpty { "--" })
        } catch (_: Exception) { SignalBarItem(pair) }
    }

    private suspend fun loadFeedCards(): List<FeedCard> = defaultPairs.take(3).mapNotNull { pair ->
        try {
            val req = Signals.GetUnifiedRequest.newBuilder().setPair(pair).build()
            val resp = RpcHelper.unary(
                "antclaw.v1.SignalsService/GetUnified", req,
                Signals.GetUnifiedRequest::class, Signals.GetUnifiedResponse::class)
            if (!resp.hasSignal()) return@mapNotNull null
            FeedCard("signal_$pair", authorSystem, resp.signal.pair, resp.signal.direction,
                ((resp.signal.confidence) * 100).toInt(),
                resp.signal.contributingFactorsList.joinToString(" \u00b7 "), timeJustNow)
        } catch (_: Exception) { null }
    }
}
