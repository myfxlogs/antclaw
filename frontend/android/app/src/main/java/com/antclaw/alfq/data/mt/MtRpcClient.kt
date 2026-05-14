package com.antclaw.alfq.data.mt

import antclaw.v1.Mt4
import antclaw.v1.Mt5
import com.antclaw.alfq.data.rpc.RpcHelper

object MtRpcClient {

    // ========== MT4 ==========

    suspend fun addMt4Account(req: AddMt4Request): Mt4Account {
        val proto = Mt4.AddMT4AccountRequest.newBuilder()
            .setServer(req.server).setAccount(req.account)
            .setInvestorPassword(req.investorPassword)
            .setLabel(req.label).setIsDemo(req.isDemo)
            .build()
        return RpcHelper.unary("antclaw.v1.MT4Service/AddAccount", proto,
            Mt4.AddMT4AccountRequest::class, Mt4.MT4Account::class).toDomain()
    }

    suspend fun removeMt4Account(id: String): Boolean {
        val proto = Mt4.RemoveMT4AccountRequest.newBuilder().setId(id).build()
        return RpcHelper.unary("antclaw.v1.MT4Service/RemoveAccount", proto,
            Mt4.RemoveMT4AccountRequest::class, Mt4.RemoveMT4AccountResponse::class).success
    }

    suspend fun getMt4AccountInfo(id: String): Mt4AccountInfo {
        val proto = Mt4.GetMT4AccountInfoRequest.newBuilder().setId(id).build()
        return RpcHelper.unary("antclaw.v1.MT4Service/GetAccountInfo", proto,
            Mt4.GetMT4AccountInfoRequest::class, Mt4.MT4AccountInfo::class).toDomain()
    }

    // ========== MT5 ==========

    suspend fun addMt5Account(req: AddMt5Request): Mt5Account {
        val proto = Mt5.AddMT5AccountRequest.newBuilder()
            .setServer(req.server).setAccount(req.account)
            .setInvestorPassword(req.investorPassword)
            .setLabel(req.label).setIsDemo(req.isDemo)
            .build()
        return RpcHelper.unary("antclaw.v1.MT5Service/AddAccount", proto,
            Mt5.AddMT5AccountRequest::class, Mt5.MT5Account::class).toDomain()
    }

    suspend fun removeMt5Account(id: String): Boolean {
        val proto = Mt5.RemoveMT5AccountRequest.newBuilder().setId(id).build()
        return RpcHelper.unary("antclaw.v1.MT5Service/RemoveAccount", proto,
            Mt5.RemoveMT5AccountRequest::class, Mt5.RemoveMT5AccountResponse::class).success
    }

    suspend fun getMt5AccountInfo(id: String): Mt5AccountInfo {
        val proto = Mt5.GetMT5AccountInfoRequest.newBuilder().setId(id).build()
        return RpcHelper.unary("antclaw.v1.MT5Service/GetAccountInfo", proto,
            Mt5.GetMT5AccountInfoRequest::class, Mt5.MT5AccountInfo::class).toDomain()
    }
}

// Proto -> Domain 转换扩展
private fun Mt4.MT4Account.toDomain() = Mt4Account(
    id = id, server = server, account = account, label = label,
    isDemo = isDemo, connected = connected, createdAt = createdAt,
)

private fun Mt4.MT4AccountInfo.toDomain() = Mt4AccountInfo(
    id = id, balance = balance, equity = equity, margin = margin,
    freeMargin = freeMargin, marginLevel = marginLevel,
    todayPnl = todayPnl, positionCount = positionCount, updatedAt = updatedAt,
)

private fun Mt5.MT5Account.toDomain() = Mt5Account(
    id = id, server = server, account = account, label = label,
    isDemo = isDemo, connected = connected, createdAt = createdAt,
)

private fun Mt5.MT5AccountInfo.toDomain() = Mt5AccountInfo(
    id = id, balance = balance, equity = equity, margin = margin,
    freeMargin = freeMargin, marginLevel = marginLevel,
    todayPnl = todayPnl, positionCount = positionCount, updatedAt = updatedAt,
)
