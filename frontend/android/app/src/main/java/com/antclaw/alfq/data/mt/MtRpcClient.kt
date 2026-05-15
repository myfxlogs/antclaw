package com.antclaw.alfq.data.mt

import antclaw.v1.Mt4
import antclaw.v1.Mt5
import com.antclaw.alfq.data.rpc.BaseRpcClient
import com.connectrpc.ProtocolClientInterface
import javax.inject.Inject
import javax.inject.Singleton

/** MT4 RPC */
@Singleton
class Mt4Rpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun addAccount(req: Mt4.AddMT4AccountRequest) =
        unary("MT4Service/AddAccount", req, Mt4.AddMT4AccountRequest::class, Mt4.MT4Account::class)
    suspend fun removeAccount(id: String) =
        unary("MT4Service/RemoveAccount",
            Mt4.RemoveMT4AccountRequest.newBuilder().setId(id).build(),
            Mt4.RemoveMT4AccountRequest::class, Mt4.RemoveMT4AccountResponse::class)
    suspend fun getAccountInfo(id: String) =
        unary("MT4Service/GetAccountInfo",
            Mt4.GetMT4AccountInfoRequest.newBuilder().setId(id).build(),
            Mt4.GetMT4AccountInfoRequest::class, Mt4.MT4AccountInfo::class)
}

/** MT5 RPC */
@Singleton
class Mt5Rpc @Inject constructor(client: ProtocolClientInterface) : BaseRpcClient(client) {
    suspend fun addAccount(req: Mt5.AddMT5AccountRequest) =
        unary("MT5Service/AddAccount", req, Mt5.AddMT5AccountRequest::class, Mt5.MT5Account::class)
    suspend fun removeAccount(id: String) =
        unary("MT5Service/RemoveAccount",
            Mt5.RemoveMT5AccountRequest.newBuilder().setId(id).build(),
            Mt5.RemoveMT5AccountRequest::class, Mt5.RemoveMT5AccountResponse::class)
    suspend fun getAccountInfo(id: String) =
        unary("MT5Service/GetAccountInfo",
            Mt5.GetMT5AccountInfoRequest.newBuilder().setId(id).build(),
            Mt5.GetMT5AccountInfoRequest::class, Mt5.MT5AccountInfo::class)
}

@Singleton
class MtRpcClient @Inject constructor(
    private val mt4: Mt4Rpc,
    private val mt5: Mt5Rpc,
) {

    // ========== MT4 ==========

    suspend fun addMt4Account(req: AddMt4Request): Mt4Account =
        mt4.addAccount(Mt4.AddMT4AccountRequest.newBuilder()
            .setServer(req.server).setAccount(req.account)
            .setInvestorPassword(req.investorPassword).setLabel(req.label).setIsDemo(req.isDemo)
            .build()).toDomain()

    suspend fun removeMt4Account(id: String): Boolean = mt4.removeAccount(id).success

    suspend fun getMt4AccountInfo(id: String): Mt4AccountInfo =
        mt4.getAccountInfo(id).toDomain()

    // ========== MT5 ==========

    suspend fun addMt5Account(req: AddMt5Request): Mt5Account =
        mt5.addAccount(Mt5.AddMT5AccountRequest.newBuilder()
            .setServer(req.server).setAccount(req.account)
            .setInvestorPassword(req.investorPassword).setLabel(req.label).setIsDemo(req.isDemo)
            .build()).toDomain()

    suspend fun removeMt5Account(id: String): Boolean = mt5.removeAccount(id).success

    suspend fun getMt5AccountInfo(id: String): Mt5AccountInfo =
        mt5.getAccountInfo(id).toDomain()
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
