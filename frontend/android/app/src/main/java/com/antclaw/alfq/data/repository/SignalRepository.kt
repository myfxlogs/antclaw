package com.antclaw.alfq.data.repository

import com.antclaw.alfq.data.rpc.SignalRpc
import com.antclaw.alfq.data.rpc.PriceRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SignalRepository @Inject constructor(
    private val signalRpc: SignalRpc,
    private val priceRpc: PriceRpc,
) {
    data class SignalDetail(val direction: String, val confidence: Int, val price: String)
    suspend fun getDetail(pair: String): SignalDetail {
        val signal = signalRpc.getUnified(pair)
        val price = priceRpc.getPrice(pair)
        return SignalDetail(
            direction = if (signal.hasSignal()) signal.signal.direction else "neutral",
            confidence = if (signal.hasSignal()) ((signal.signal.confidence) * 100).toInt() else 0,
            price = price.current.ifEmpty { "--" },
        )
    }
}
