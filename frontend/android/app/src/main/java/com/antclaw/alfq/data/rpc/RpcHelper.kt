package com.antclaw.alfq.data.rpc

import com.connectrpc.MethodSpec
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import com.google.protobuf.MessageLite
import kotlin.reflect.KClass

/**
 * RPC 调用辅助 — 将 4 行样板压缩为 1 行。
 *
 * 用法：
 *   val resp = RpcHelper.unary("antclaw.v1.SignalsService/GetBias",
 *       biasReq, Signals.GetBiasRequest::class, Signals.GetBiasResponse::class)
 */
object RpcHelper {

    @Suppress("UNCHECKED_CAST")
    suspend fun <Req : MessageLite, Res : MessageLite> unary(
        path: String,
        req: Req,
        reqClass: KClass<Req>,
        resClass: KClass<Res>,
    ): Res {
        val spec = MethodSpec(path, reqClass, resClass, StreamType.UNARY)
        val resp = ConnectTransportProvider.createProtocolClient().unary(req, emptyMap(), spec)
        return resp.getOrThrow() as Res
    }
}
