package com.antclaw.alfq.data.rpc

import android.util.Log
import antclaw.v1.Auth
import com.antclaw.alfq.BuildConfig
import com.antclaw.alfq.data.local.TokenStoreApi
import com.antclaw.alfq.data.session.SessionExpiredNotifier
import com.connectrpc.MethodSpec
import com.connectrpc.ProtocolClientConfig
import com.connectrpc.ProtocolClientInterface
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.extensions.GoogleJavaProtobufStrategy
import com.connectrpc.getOrThrow
import com.connectrpc.impl.ProtocolClient
import com.connectrpc.okhttp.ConnectOkHttpClient
import kotlinx.coroutines.runBlocking
import okhttp3.OkHttpClient
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Single-flight token 刷新协调器。
 *
 * 职责：
 * 1. 从持久化恢复 token 到 [TokenSnapshot]（启动时调用一次）
 * 2. 多线程安全的 single-flight token 刷新（供 [AuthInterceptor] 401 时调用）
 * 3. 刷新失败时通知 [SessionExpiredNotifier] 并清空 token
 *
 * 刷新使用独立的 no-auth ProtocolClient（不经过 AuthInterceptor）。
 */
@Singleton
class RefreshTokenCoordinator @Inject constructor(
    private val tokenStore: TokenStoreApi,
    private val snapshot: TokenSnapshot,
) {
    companion object {
        private const val TAG = "RefreshTokenCoord"
    }

    private var sessionExpiredNotifier: SessionExpiredNotifier? = null

    // ── Single-flight refresh 锁 ──
    private val refreshLock = Any()
    @Volatile private var isRefreshing = false
    @Volatile private var lastRefreshResult: String? = null

    // ── 启动恢复 ──

    fun restore(notifier: SessionExpiredNotifier?) {
        sessionExpiredNotifier = notifier
        snapshot.setToken(runBlocking { tokenStore.getAccessToken() } ?: return)
    }

    // ── Single-flight refresh（线程安全） ──

    fun refreshBlocking(): String? {
        // Fast path: another thread already finished refreshing
        if (!isRefreshing && lastRefreshResult != null) {
            val cached = lastRefreshResult
            lastRefreshResult = null
            return cached
        }

        // Wait if another thread is refreshing
        if (isRefreshing) {
            synchronized(refreshLock) {
                while (isRefreshing) {
                    (refreshLock as java.lang.Object).wait(5000)
                }
            }
            val cached = lastRefreshResult
            lastRefreshResult = null
            return cached
        }

        // Acquire lock
        synchronized(refreshLock) {
            if (isRefreshing) {
                while (isRefreshing) (refreshLock as java.lang.Object).wait(5000)
                val cached = lastRefreshResult; lastRefreshResult = null; return cached
            }
            isRefreshing = true
        }

        try {
            val result = runBlocking {
                try {
                    val rt = tokenStore.getRefreshToken() ?: return@runBlocking null
                    val request = Auth.RefreshRequest.newBuilder()
                        .setRefreshToken(rt).build()
                    val spec = MethodSpec(
                        path = "antclaw.v1.AuthService/Refresh",
                        requestClass = Auth.RefreshRequest::class,
                        responseClass = Auth.RefreshResponse::class,
                        streamType = StreamType.UNARY,
                    )
                    val noAuthClient: ProtocolClientInterface = ProtocolClient(
                        httpClient = ConnectOkHttpClient(unaryClient = OkHttpClient()),
                        config = ProtocolClientConfig(
                            host = BuildConfig.BASE_URL,
                            serializationStrategy = GoogleJavaProtobufStrategy(),
                        ),
                    )
                    val resp: ResponseMessage<Auth.RefreshResponse> =
                        noAuthClient.unary(request, emptyMap(), spec)
                    val newAccess = resp.getOrThrow().accessToken
                    snapshot.setToken(newAccess)
                    tokenStore.saveAccessToken(newAccess)
                    newAccess
                } catch (e: Exception) {
                    Log.e(TAG, "Token refresh failed", e)
                    snapshot.clearToken()
                    tokenStore.clearTokens()
                    sessionExpiredNotifier?.notifySessionExpired()
                    null
                }
            }
            lastRefreshResult = result
            return result
        } finally {
            synchronized(refreshLock) {
                isRefreshing = false
                (refreshLock as java.lang.Object).notifyAll()
            }
        }
    }
}
