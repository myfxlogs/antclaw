package com.antclaw.alfq.data.rpc

import antclaw.v1.Alerts
import antclaw.v1.AlfqChat
import antclaw.v1.Price
import antclaw.v1.Signals
import antclaw.v1.UserOuterClass
import com.connectrpc.ProtocolClientInterface
import javax.inject.Inject
import javax.inject.Singleton

/** AlertService */
@Singleton
class AlertRpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun listSubscriptions() =
        unary("AlertService/ListSubscriptions",
            Alerts.ListSubscriptionsRequest.getDefaultInstance(),
            Alerts.ListSubscriptionsRequest::class, Alerts.ListSubscriptionsResponse::class)

    suspend fun subscribe(req: Alerts.SubscribeRequest) =
        unary("AlertService/Subscribe", req,
            Alerts.SubscribeRequest::class, Alerts.SubscribeResponse::class)

    suspend fun unsubscribe(id: String) =
        unary("AlertService/Unsubscribe",
            Alerts.UnsubscribeRequest.newBuilder().setSubscriptionId(id).build(),
            Alerts.UnsubscribeRequest::class, Alerts.UnsubscribeResponse::class)
}

/** ChatService */
@Singleton
class ChatRpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun listConversations() =
        unary("ChatService/ListConversations",
            AlfqChat.ListConversationsRequest.getDefaultInstance(),
            AlfqChat.ListConversationsRequest::class, AlfqChat.ConversationList::class)
}

/** UserService */
@Singleton
class UserRpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun getMe() =
        unary("UserService/GetMe",
            UserOuterClass.GetMeRequest.getDefaultInstance(),
            UserOuterClass.GetMeRequest::class, UserOuterClass.GetMeResponse::class)
}

/** SignalsService */
@Singleton
class SignalRpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun getUnified(pair: String) =
        unary("SignalsService/GetUnified",
            Signals.GetUnifiedRequest.newBuilder().setPair(pair).build(),
            Signals.GetUnifiedRequest::class, Signals.GetUnifiedResponse::class)
}

/** PriceService */
@Singleton
class PriceRpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun getPrice(pair: String, timeframe: String = "1D", count: Int = 1) =
        unary("PriceService/GetPrice",
            Price.GetPriceRequest.newBuilder().setPair(pair).setTimeframe(timeframe).setCount(count).build(),
            Price.GetPriceRequest::class, Price.GetPriceResponse::class)
}
