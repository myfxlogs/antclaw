package com.antclaw.alfq.data.mt

data class AddMt5Request(
    val server: String,
    val account: String,
    val investorPassword: String,
    val label: String = "",
    val isDemo: Boolean = true,
)

data class Mt5Account(
    val id: String = "",
    val server: String = "",
    val account: String = "",
    val label: String = "",
    val isDemo: Boolean = true,
    val connected: Boolean = false,
    val createdAt: Long = 0,
)

data class Mt5AccountInfo(
    val id: String = "",
    val balance: Double = 0.0,
    val equity: Double = 0.0,
    val margin: Double = 0.0,
    val freeMargin: Double = 0.0,
    val marginLevel: Double = 0.0,
    val todayPnl: Double = 0.0,
    val positionCount: Int = 0,
    val updatedAt: Long = 0,
)
