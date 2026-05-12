package com.antclaw.alfq.ui.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.*
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SignalBarItem(
    val pair: String,
    val direction: String,
    val confidence: Int,
    val price: String,
)

data class FeedCard(
    val id: String,
    val type: String,
    val author: String,
    val pair: String? = null,
    val direction: String? = null,
    val confidence: Int? = null,
    val content: String = "",
    val timeAgo: String = "",
)

data class FeedUiState(
    val signalBar: List<SignalBarItem> = emptyList(),
    val cards: List<FeedCard> = emptyList(),
    val loading: Boolean = true,
    val error: String? = null,
)

@HiltViewModel
class FeedViewModel @Inject constructor() : ViewModel() {
    private val signalsClient = SignalsRpcClient()
    private val priceClient = PriceRpcClient()

    private val _uiState = MutableStateFlow(FeedUiState())
    val uiState: StateFlow<FeedUiState> = _uiState.asStateFlow()

    private val defaultPairs = listOf("EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "XAUUSD", "BTCUSD")

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                val signalBar = loadSignalBar()
                val cards = loadFeedCards()
                _uiState.value = FeedUiState(signalBar = signalBar, cards = cards, loading = false)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(loading = false, error = e.message ?: "加载失败")
            }
        }
    }

    private suspend fun loadSignalBar(): List<SignalBarItem> {
        return defaultPairs.map { pair ->
            try {
                val biasResp = signalsClient.getBias(SigBiasReq(pair, "1D"))
                val bias = biasResp.biases.firstOrNull()
                val priceResp = priceClient.getBars(PriceReq(pair, "1D", 1))
                SignalBarItem(
                    pair = pair,
                    direction = bias?.direction ?: "neutral",
                    confidence = ((bias?.confidence ?: 0.0) * 100).toInt(),
                    price = priceResp.current.ifEmpty { "--" }
                )
            } catch (_: Exception) {
                SignalBarItem(pair = pair, direction = "neutral", confidence = 0, price = "--")
            }
        }
    }

    private suspend fun loadFeedCards(): List<FeedCard> {
        val cards = mutableListOf<FeedCard>()
        for (pair in defaultPairs.take(3)) {
            try {
                val unified = signalsClient.getUnified(SigUnifiedReq(pair))
                val sig = unified.signal ?: continue
                cards.add(FeedCard(
                    id = "signal_$pair",
                    type = "signal_card",
                    author = "系统信号",
                    pair = sig.pair,
                    direction = sig.direction,
                    confidence = ((sig.confidence) * 100).toInt(),
                    content = sig.contributing_factors.joinToString(" \u00b7 "),
                    timeAgo = "刚刚"
                ))
            } catch (_: Exception) { }
        }
        return cards
    }
}
