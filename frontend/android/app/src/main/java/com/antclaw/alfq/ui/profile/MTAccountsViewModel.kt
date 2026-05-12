package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.rpc.MT5AddAccountReq
import com.antclaw.alfq.data.rpc.MT5InfoReq
import com.antclaw.alfq.data.rpc.MT5InfoResp
import com.antclaw.alfq.data.rpc.MT5PositionResp
import com.antclaw.alfq.data.rpc.MT5RemoveReq
import com.antclaw.alfq.data.rpc.MT5RpcClient
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MTAccountUi(
    val id: String,
    val server: String,
    val account: String,
    val label: String,
    val isDemo: Boolean,
    val info: MT5InfoResp? = null,
    val positions: List<MT5PositionResp> = emptyList(),
    val orderCount: Int = 0,
)

data class MTAccountsUiState(
    val accounts: List<MTAccountUi> = emptyList(),
    val loading: Boolean = true,
)

@HiltViewModel
class MTAccountsViewModel @Inject constructor() : ViewModel() {

    private val rpc = MT5RpcClient()

    private val _uiState = MutableStateFlow(MTAccountsUiState())
    val uiState: StateFlow<MTAccountsUiState> = _uiState.asStateFlow()

    var showAddDialog by mutableStateOf(false)
        private set

    // Demo accounts for UI testing (replace with real RPC when backend ready)
    private val demoAccounts = listOf(
        MTAccountUi(
            id = "1", server = "ICMarkets-Demo", account = "88005522",
            label = "主要实盘", isDemo = false,
            info = MT5InfoResp(balance = 12450.0, equity = 12680.0, margin = 980.0, today_pnl = 0.032),
            positions = listOf(MT5PositionResp(ticket = 1001, symbol = "EURUSD", type = "BUY"), MT5PositionResp(ticket = 1002, symbol = "XAUUSD", type = "SELL")),
            orderCount = 247
        ),
        MTAccountUi(
            id = "2", server = "Exness-Demo", account = "11223344",
            label = "策略测试", isDemo = true,
            info = MT5InfoResp(balance = 5000.0, equity = 5040.0, today_pnl = 0.008),
            orderCount = 45
        )
    )

    init {
        viewModelScope.launch {
            // Placeholder: load from local storage
            _uiState.value = MTAccountsUiState(accounts = demoAccounts, loading = false)
        }
    }

    fun addAccount(server: String, account: String, password: String, label: String, isDemo: Boolean) {
        viewModelScope.launch {
            showAddDialog = false
            try {
                // rpc.addAccount(MT5AddAccountReq(server, account, password, label, isDemo))
                // Reload list
            } catch (e: Exception) {
                // Handle error
            }
        }
    }

    fun removeAccount(id: String) {
        viewModelScope.launch {
            try {
                // rpc.removeAccount(MT5RemoveReq(id))
            } catch (e: Exception) { }
        }
    }
}
