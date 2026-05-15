package com.antclaw.alfq.data.rpc

import antclaw.v1.Auth
import antclaw.v1.System
import com.antclaw.alfq.BuildConfig
import com.antclaw.alfq.data.local.TokenStore
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
import okhttp3.Response
import java.io.IOException

object ConnectTransportProvider {

    val baseUrl: String = BuildConfig.BASE_URL

    private var tokenProvider: (() -> String?)? = null
    private var tokenStore: TokenStore? = null
    private var sessionExpiredNotifier: com.antclaw.alfq.data.session.SessionExpiredNotifier? = null

    // ── Single-flight refresh ──
    private val refreshLock = Any()
    @Volatile private var isRefreshing = false
    @Volatile private var lastRefreshResult: String? = null

    private val noAuthOkHttpClient by lazy { OkHttpClient() }
    private val authOkHttpClient by lazy {
        OkHttpClient.Builder()
            .addInterceptor { chain ->
                val token = tokenProvider?.invoke()
                val request = if (token != null) {
                    chain.request().newBuilder()
                        .header("Authorization", "Bearer $token").build()
                } else chain.request()

                val response: Response = try {
                    chain.proceed(request)
                } catch (e: IOException) { throw e }

                if (response.code == 401) {
                    response.close()
                    val newToken = refreshTokenSingleFlight()
                    if (newToken != null) {
                        val retryRequest = chain.request().newBuilder()
                            .header("Authorization", "Bearer $newToken").build()
                        chain.proceed(retryRequest)
                    } else {
                        chain.proceed(request)
                    }
                } else {
                    response
                }
            }.build()
    }

    private val noAuthProtocolClient: ProtocolClientInterface by lazy {
        ProtocolClient(
            httpClient = ConnectOkHttpClient(unaryClient = noAuthOkHttpClient),
            config = ProtocolClientConfig(
                host = baseUrl,
                serializationStrategy = GoogleJavaProtobufStrategy(),
            ),
        )
    }

    private val authProtocolClient: ProtocolClientInterface by lazy {
        ProtocolClient(
            httpClient = ConnectOkHttpClient(unaryClient = authOkHttpClient),
            config = ProtocolClientConfig(
                host = baseUrl,
                serializationStrategy = GoogleJavaProtobufStrategy(),
            ),
        )
    }

    fun init(tokenStore: TokenStore, notifier: com.antclaw.alfq.data.session.SessionExpiredNotifier? = null) {
        this.tokenStore = tokenStore
        this.sessionExpiredNotifier = notifier
        val persistedToken = runBlocking { tokenStore.getAccessToken() }
        if (persistedToken != null) {
            tokenProvider = { runBlocking { tokenStore.getAccessToken() } }
        }
    }

    fun setToken(token: String) { tokenProvider = { token } }
    fun clearToken() { tokenProvider = null }
    fun getToken(): String? = tokenProvider?.invoke()

    /**
     * Single-flight token refresh — 多个并发 401 只发起一次刷新。
     * 第一个请求执行刷新，后续请求等待并复用结果。
     * 在 OkHttp 拦截器（非协程线程）中调用，使用 synchronized + wait/notifyAll。
     */
    private fun refreshTokenSingleFlight(): String? {
        val store = tokenStore ?: return null

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

        // Acquire the refresh lock and execute
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
                    val refreshToken = store.getRefreshToken() ?: return@runBlocking null
                    val request = Auth.RefreshRequest.newBuilder()
                        .setRefreshToken(refreshToken).build()
                    val spec = MethodSpec(
                        path = "antclaw.v1.AuthService/Refresh",
                        requestClass = Auth.RefreshRequest::class,
                        responseClass = Auth.RefreshResponse::class,
                        streamType = StreamType.UNARY,
                    )
                    val resp: ResponseMessage<Auth.RefreshResponse> = noAuthProtocolClient.unary(request, emptyMap(), spec)
                    val newAccess = resp.getOrThrow().accessToken
                    setToken(newAccess)
                    store.saveAccessToken(newAccess)
                    newAccess
                } catch (_: Exception) {
                    clearToken()
                    store.clearTokens()
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

    fun createProtocolClient(): ProtocolClientInterface = authProtocolClient

    fun healthz(): Result<String> {
        return runBlocking {
            try {
                val client = noAuthProtocolClient
                val spec = MethodSpec(
                    path = "antclaw.v1.SystemService/Healthz",
                    requestClass = System.HealthzRequest::class,
                    responseClass = System.HealthzResponse::class,
                    streamType = StreamType.UNARY,
                )
                val resp: ResponseMessage<System.HealthzResponse> = client.unary(
                    System.HealthzRequest.getDefaultInstance(), emptyMap(), spec
                )
                Result.success(resp.getOrThrow().status)
            } catch (e: Exception) { Result.failure(e) }
        }
    }
}
