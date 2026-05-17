package com.antclaw.alfq.data.repository

import antclaw.v1.Alerts
import com.antclaw.alfq.data.rpc.AlertRpc
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AlertRepository @Inject constructor(private val rpc: AlertRpc) {
    data class Sub(val id: String, val pair: String, val condition: String, val threshold: String, val type: String, val active: Boolean)
    suspend fun listSubscriptions(): List<Sub> = rpc.listSubscriptions().subscriptionsList.map { s ->
        Sub(s.subscriptionId, s.pair, s.condition, s.threshold, s.alertType, s.active)
    }
    suspend fun subscribe(type: String, pair: String, condition: String, threshold: String) {
        rpc.subscribe(Alerts.SubscribeRequest.newBuilder().setAlertType(type).setPair(pair).setCondition(condition).setThreshold(threshold).build())
    }
    suspend fun unsubscribe(id: String) { rpc.unsubscribe(id) }
}
