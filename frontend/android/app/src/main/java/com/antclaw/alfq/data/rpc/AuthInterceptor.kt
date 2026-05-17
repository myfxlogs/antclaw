package com.antclaw.alfq.data.rpc

import okhttp3.Interceptor
import okhttp3.Response
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

/**
 * OkHttp 认证拦截器。
 *
 * 职责：
 * 1. 每次请求注入 `Authorization: Bearer <token>`。
 * 2. 收到 401 时触发 [RefreshTokenCoordinator.refreshBlocking] 并重试一次。
 *
 * 不持有 OkHttpClient 本身 — 由 [ProtocolClientFactory] 装配到 client builder。
 */
@Singleton
class AuthInterceptor @Inject constructor(
    private val snapshot: TokenSnapshot,
    private val refreshCoordinator: RefreshTokenCoordinator,
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val token = snapshot.getToken()
        val request = if (token != null) {
            chain.request().newBuilder()
                .header("Authorization", "Bearer $token").build()
        } else chain.request()

        val response: Response = try {
            chain.proceed(request)
        } catch (e: IOException) { throw e }

        if (response.code == 401) {
            response.close()
            val newToken = refreshCoordinator.refreshBlocking()
            if (newToken != null) {
                val retryRequest = chain.request().newBuilder()
                    .header("Authorization", "Bearer $newToken").build()
                return chain.proceed(retryRequest)
            }
            return chain.proceed(request)
        }
        return response
    }
}
