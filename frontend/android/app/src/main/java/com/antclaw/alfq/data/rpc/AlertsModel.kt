package com.antclaw.alfq.data.rpc

data class AlertSub(
    val id: String,
    val type: String,
    val pair: String,
    val condition: String,
    val threshold: String,
    val alert_type: String = type,
    val active: Boolean = true,
    val subscription_id: String = id,
)

data class AlertSubReq(
    val type: String,
    val pair: String,
    val condition: String,
    val threshold: String,
)

data class CreateAlertReq(
    val type: String = "",
    val pair: String = "",
)

data class AlertListResp(
    val subscriptions: List<AlertSub>,
)
