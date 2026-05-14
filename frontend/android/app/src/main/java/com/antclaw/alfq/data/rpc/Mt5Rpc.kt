package com.antclaw.alfq.data.rpc

import javax.inject.Inject

// MT5 RPC client — placeholder stubs. Real RPC implementation pending backend MT5 integration.

data class MT5AddAccountReq(val server: String, val account: String, val password: String, val label: String, val isDemo: Boolean)
data class MT5InfoReq(val account_id: String)
data class MT5InfoResp(val balance: Double, val equity: Double, val margin: Double, val today_pnl: Double)
data class MT5PositionResp(val ticket: Int, val symbol: String, val type: String)
data class MT5RemoveReq(val account_id: String)

class MT5RpcClient @Inject constructor() {
    // TODO: Implement real RPC via ConnectTransportProvider.createProtocolClient()
}
