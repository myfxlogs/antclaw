package com.antclaw.alfq.data.rpc

import antclaw.v1.Signals
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject
import javax.inject.Singleton

// ── Client ──

@Singleton
class SignalsRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun getBias(req: SigBiasReq): SigBiasResp {
        val request = Signals.GetBiasRequest.newBuilder()
            .setPair(req.pair)
            .setTimeframe(req.timeframe)
            .build()

        val methodSpec = MethodSpec(
            path = "antclaw.v1.SignalsService/GetBias",
            requestClass = Signals.GetBiasRequest::class,
            responseClass = Signals.GetBiasResponse::class,
            streamType = StreamType.UNARY,
        )

        val response: ResponseMessage<Signals.GetBiasResponse> =
            client.unary(request, emptyMap(), methodSpec)

        val resp = response.getOrThrow()
        return SigBiasResp(
            biases = resp.biasesList.map { b ->
                BiasItem(direction = b.direction, confidence = b.confidence)
            }
        )
    }

    suspend fun getUnified(req: SigUnifiedReq): SigUnifiedResp {
        val request = Signals.GetUnifiedRequest.newBuilder()
            .setPair(req.pair)
            .build()

        val methodSpec = MethodSpec(
            path = "antclaw.v1.SignalsService/GetUnified",
            requestClass = Signals.GetUnifiedRequest::class,
            responseClass = Signals.GetUnifiedResponse::class,
            streamType = StreamType.UNARY,
        )

        val response: ResponseMessage<Signals.GetUnifiedResponse> =
            client.unary(request, emptyMap(), methodSpec)

        val resp = response.getOrThrow()
        return if (resp.hasSignal()) {
            SigUnifiedResp(
                signal = UnifiedSignal(
                    pair = resp.signal.pair,
                    direction = resp.signal.direction,
                    confidence = resp.signal.confidence,
                    contributing_factors = resp.signal.contributingFactorsList,
                )
            )
        } else {
            SigUnifiedResp(signal = null)
        }
    }

    suspend fun getXFactors(req: SigXFactorsReq): SigXFactorsResp {
        val request = Signals.GetXFactorsRequest.newBuilder()
            .setPair(req.pair)
            .build()

        val methodSpec = MethodSpec(
            path = "antclaw.v1.SignalsService/GetXFactors",
            requestClass = Signals.GetXFactorsRequest::class,
            responseClass = Signals.GetXFactorsResponse::class,
            streamType = StreamType.UNARY,
        )

        val response: ResponseMessage<Signals.GetXFactorsResponse> =
            client.unary(request, emptyMap(), methodSpec)

        val resp = response.getOrThrow()
        return SigXFactorsResp(
            factors = resp.factorsList.map { f ->
                FactorData(name = f.name, weight = f.weight)
            }
        )
    }
}
