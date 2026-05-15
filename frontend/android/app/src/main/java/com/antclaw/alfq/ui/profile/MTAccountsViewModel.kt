package com.antclaw.alfq.ui.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.data.mt.*
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class MtAccountItem(
    val id: String,
    val server: String,
    val account: String,
    val label: String,
    val type: String,
    val isDemo: Boolean,
    val connected: Boolean,
    val balance: Double,
    val equity: Double,
    val createdAt: Long,
)

data class MtAccountsUiState(
    val accounts: List<MtAccountItem> = emptyList(),
    val loading: Boolean = false,
    val error: String? = null,
    val binding: Boolean = false,
    val bindError: String? = null,
    val bindSuccess: Boolean = false,
)

@HiltViewModel
class MTAccountsViewModel @Inject constructor(
    private val mtRpc: MtRpcClient,
) : ViewModel() {
    private val _uiState = MutableStateFlow(MtAccountsUiState())
    val uiState: StateFlow<MtAccountsUiState> = _uiState.asStateFlow()

    fun loadAccounts() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(loading = true, error = null)
            try {
                refreshBalances()
                _uiState.value = _uiState.value.copy(loading = false)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(loading = false, error = e.message)
            }
        }
    }

    private suspend fun refreshBalances() {
        val updated = _uiState.value.accounts.map { account ->
            try {
                when (account.type) {
                    "MT4" -> {
                        val info = mtRpc.getMt4AccountInfo(account.id)
                        account.copy(balance = info.balance, equity = info.equity)
                    }
                    "MT5" -> {
                        val info = mtRpc.getMt5AccountInfo(account.id)
                        account.copy(balance = info.balance, equity = info.equity)
                    }
                    else -> account
                }
            } catch (_: Exception) { account }
        }
        _uiState.value = _uiState.value.copy(accounts = updated)
    }

    fun bindMt4Account(server: String, account: String, password: String, label: String, isDemo: Boolean) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(binding = true, bindError = null)
            try {
                val req = AddMt4Request(server, account, password, label, isDemo)
                val a = mtRpc.addMt4Account(req)
                val item = MtAccountItem(
                    a.id, a.server, a.account, a.label, "MT4",
                    a.isDemo, a.connected, 0.0, 0.0, a.createdAt
                )
                _uiState.value = _uiState.value.copy(
                    binding = false, bindSuccess = true,
                    accounts = _uiState.value.accounts + item,
                )
                refreshBalances()
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    binding = false,
                    bindError = e.message ?: "\u7ed1\u5b9a\u5931\u8d25"
                )
            }
        }
    }

    fun bindMt5Account(server: String, account: String, password: String, label: String, isDemo: Boolean) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(binding = true, bindError = null)
            try {
                val req = AddMt5Request(server, account, password, label, isDemo)
                val a = mtRpc.addMt5Account(req)
                val item = MtAccountItem(
                    a.id, a.server, a.account, a.label, "MT5",
                    a.isDemo, a.connected, 0.0, 0.0, a.createdAt
                )
                _uiState.value = _uiState.value.copy(
                    binding = false, bindSuccess = true,
                    accounts = _uiState.value.accounts + item,
                )
                refreshBalances()
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    binding = false,
                    bindError = e.message ?: "\u7ed1\u5b9a\u5931\u8d25"
                )
            }
        }
    }

    fun removeMt4Account(id: String) {
        viewModelScope.launch {
            try {
                mtRpc.removeMt4Account(id)
                _uiState.value = _uiState.value.copy(
                    accounts = _uiState.value.accounts.filter { it.id != id }
                )
            } catch (_: Exception) {}
        }
    }

    fun removeMt5Account(id: String) {
        viewModelScope.launch {
            try {
                mtRpc.removeMt5Account(id)
                _uiState.value = _uiState.value.copy(
                    accounts = _uiState.value.accounts.filter { it.id != id }
                )
            } catch (_: Exception) {}
        }
    }

    fun clearBindResult() {
        _uiState.value = _uiState.value.copy(bindSuccess = false, bindError = null)
    }
}
