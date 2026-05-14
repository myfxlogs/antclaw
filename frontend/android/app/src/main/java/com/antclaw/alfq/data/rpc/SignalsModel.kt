package com.antclaw.alfq.data.rpc

// ── Request types ──

data class SigBiasReq(
    val pair: String,
    val timeframe: String,
)

data class SigUnifiedReq(
    val pair: String,
)

data class SigXFactorsReq(
    val pair: String,
)

// ── Item types ──

data class BiasItem(
    val direction: String,
    val confidence: Double,
)

data class UnifiedSignal(
    val pair: String,
    val direction: String,
    val confidence: Double,
    val contributing_factors: List<String>,
)

data class FactorData(
    val name: String,
    val weight: Double,
)

// ── Response types ──

data class SigBiasResp(
    val biases: List<BiasItem>,
)

data class SigUnifiedResp(
    val signal: UnifiedSignal?,
)

data class SigXFactorsResp(
    val factors: List<FactorData>,
)
