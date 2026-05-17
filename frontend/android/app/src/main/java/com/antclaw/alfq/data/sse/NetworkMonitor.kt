package com.antclaw.alfq.data.sse

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.util.Log
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 网络状态监听器 — 从 SseManager 分离。
 *
 * 职责：
 * 1. 注册 ConnectivityManager.NetworkCallback
 * 2. 暴露 `isAvailable: StateFlow<Boolean>` — 无网时 SSE 暂停重连
 * 3. 生命周期：start/stop
 */
@Singleton
class NetworkMonitor @Inject constructor(
    @ApplicationContext private val appContext: Context,
) {
    companion object {
        private const val TAG = "NetworkMonitor"
    }

    private val connectivityManager: ConnectivityManager =
        appContext.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private val _isAvailable = MutableStateFlow(true)
    val isAvailable: StateFlow<Boolean> = _isAvailable.asStateFlow()

    private var callback: ConnectivityManager.NetworkCallback? = null

    fun start() {
        val request = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()
        callback = object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) {
                Log.i(TAG, "Network available")
                _isAvailable.value = true
            }
            override fun onLost(network: Network) {
                Log.i(TAG, "Network lost")
                _isAvailable.value = false
            }
            override fun onCapabilitiesChanged(network: Network, caps: NetworkCapabilities) {
                if (caps.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)) {
                    _isAvailable.value = true
                }
            }
        }
        connectivityManager.registerNetworkCallback(request, callback!!)
    }

    fun stop() {
        callback?.let { connectivityManager.unregisterNetworkCallback(it) }
        callback = null
    }
}
