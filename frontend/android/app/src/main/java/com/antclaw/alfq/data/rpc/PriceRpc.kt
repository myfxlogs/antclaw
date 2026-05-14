package com.antclaw.alfq.data.rpc

import antclaw.v1.Price
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PriceRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun getBars(req: PriceReq): PriceResp {
        val request = Price.GetPriceRequest.newBuilder()
            .setPair(req.pair).setTimeframe(req.timeframe).setCount(req.count).build()
        val resp = client.unary<Price.GetPriceRequest, Price.GetPriceResponse>(
            request, emptyMap(), MethodSpec(
                path = "antclaw.v1.PriceService/GetPrice",
                requestClass = Price.GetPriceRequest::class,
                responseClass = Price.GetPriceResponse::class,
                streamType = StreamType.UNARY,
            )
        ).getOrThrow()
        return PriceResp(current = resp.current,
            bars = resp.barsList.map { PriceBar(close = it.close) })
    }
}
