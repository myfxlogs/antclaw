package com.antclaw.alfq.data.mt

// 绑定请求
data class AddMt4Request(
    val server: String,
    val account: String,
    val investorPassword: String,
    val label: String = "",
    val isDemo: Boolean = true,
)

// 绑定响应 / 账户实体
data class Mt4Account(
    val id: String = "",
    val server: String = "",
    val account: String = "",
    val label: String = "",
    val isDemo: Boolean = true,
    val connected: Boolean = false,
    val createdAt: Long = 0,
)

// 账户概览（余额/净值/保证金）
data class Mt4AccountInfo(
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
