package com.antclaw.alfq.data.rpc

data class PriceReq(
    val pair: String,
    val timeframe: String,
    val count: Int,
)

data class PriceBar(
    val close: String,
)

data class PriceResp(
    val current: String,
    val bars: List<PriceBar>,
)
