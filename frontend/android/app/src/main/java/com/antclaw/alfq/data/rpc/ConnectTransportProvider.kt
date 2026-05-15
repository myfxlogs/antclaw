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
                    val newToken = refreshTokenBlocking()
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

    fun init(tokenStore: TokenStore) {
        this.tokenStore = tokenStore
        val persistedToken = runBlocking { tokenStore.getAccessToken() }
        if (persistedToken != null) {
            tokenProvider = { runBlocking { tokenStore.getAccessToken() } }
        }
    }

    fun setToken(token: String) { tokenProvider = { token } }
    fun clearToken() { tokenProvider = null }
    fun getToken(): String? = tokenProvider?.invoke()

    /** 同步 token 刷新（在 OkHttp 拦截器中调用，不在协程上下文中）。 */
    private fun refreshTokenBlocking(): String? {
        val store = tokenStore ?: return null
        return runBlocking {
            try {
                val refreshToken = store.getRefreshToken() ?: return@runBlocking null
                val client = noAuthProtocolClient
                val request = Auth.RefreshRequest.newBuilder()
                    .setRefreshToken(refreshToken).build()
                val spec = MethodSpec(
                    path = "antclaw.v1.AuthService/Refresh",
                    requestClass = Auth.RefreshRequest::class,
                    responseClass = Auth.RefreshResponse::class,
                    streamType = StreamType.UNARY,
                )
                val resp: ResponseMessage<Auth.RefreshResponse> = client.unary(request, emptyMap(), spec)
                val newAccess = resp.getOrThrow().accessToken
                setToken(newAccess)
                store.saveAccessToken(newAccess)
                newAccess
            } catch (_: Exception) {
                clearToken()
                store.clearTokens()
                null
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
