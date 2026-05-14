package com.antclaw.alfq.data.rpc

import antclaw.v1.Alerts
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AlertsRpcClient @Inject constructor(
    private val client: ProtocolClientInterface,
) {
    suspend fun listSubscriptions(): AlertListResp {
        val request = Alerts.ListSubscriptionsRequest.newBuilder().build()
        val methodSpec = MethodSpec(
            path = "antclaw.v1.AlertService/ListSubscriptions",
            requestClass = Alerts.ListSubscriptionsRequest::class,
            responseClass = Alerts.ListSubscriptionsResponse::class,
            streamType = StreamType.UNARY,
        )
        val resp = client.unary<Alerts.ListSubscriptionsRequest, Alerts.ListSubscriptionsResponse>(
            request, emptyMap(), methodSpec
        ).getOrThrow()
        return AlertListResp(subscriptions = resp.subscriptionsList.map { sub ->
            AlertSub(id = sub.subscriptionId, type = sub.alertType,
                pair = sub.pair, condition = sub.condition, threshold = sub.threshold,
                active = sub.active)
        })
    }

    suspend fun subscribe(req: AlertSubReq) {
        val request = Alerts.SubscribeRequest.newBuilder()
            .setAlertType(req.type).setPair(req.pair)
            .setCondition(req.condition).setThreshold(req.threshold).build()
        client.unary<Alerts.SubscribeRequest, Alerts.SubscribeResponse>(
            request, emptyMap(), MethodSpec(
                path = "antclaw.v1.AlertService/Subscribe",
                requestClass = Alerts.SubscribeRequest::class,
                responseClass = Alerts.SubscribeResponse::class,
                streamType = StreamType.UNARY,
            )
        ).getOrThrow()
    }

    suspend fun unsubscribe(id: String) {
        val request = Alerts.UnsubscribeRequest.newBuilder().setSubscriptionId(id).build()
        client.unary<Alerts.UnsubscribeRequest, Alerts.UnsubscribeResponse>(
            request, emptyMap(), MethodSpec(
                path = "antclaw.v1.AlertService/Unsubscribe",
                requestClass = Alerts.UnsubscribeRequest::class,
                responseClass = Alerts.UnsubscribeResponse::class,
                streamType = StreamType.UNARY,
            )
        ).getOrThrow()
    }
}
