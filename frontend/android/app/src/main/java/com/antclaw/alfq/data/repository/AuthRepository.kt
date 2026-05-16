package com.antclaw.alfq.data.repository

import antclaw.v1.Auth
import com.antclaw.alfq.data.device.DeviceInfoCollector
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject

class AuthRepository @Inject constructor(
    private val tokenStore: TokenStore,
    private val deviceInfoCollector: DeviceInfoCollector,
) {

    // ── 公共工具 ──

    /** 构造 Login/Register 共用 ClientInfo。 */
    private fun buildClientInfo(): Auth.ClientInfo = Auth.ClientInfo.newBuilder()
        .setDeviceId(deviceInfoCollector.getDeviceId())
        .setUserAgent(deviceInfoCollector.getUserAgent())
        .setIpAddress("")
        .build()

    suspend fun login(email: String, password: String): Result<AuthSessionResult> {
        return try {
            val client = ConnectTransportProvider.createProtocolClient()
            val request = Auth.LoginRequest.newBuilder()
                .setEmail(email)
                .setPassword(password)
                .setClient(buildClientInfo())
                .build()
                
            val spec = MethodSpec("antclaw.v1.AuthService/Login",
                Auth.LoginRequest::class, Auth.LoginResponse::class, StreamType.UNARY)
            val res: ResponseMessage<Auth.LoginResponse> = client.unary(request, emptyMap(), spec)
            val resp = res.getOrThrow()

            Result.success(
                AuthSessionResult(
                    userId = resp.userId,
                    accessToken = resp.accessToken,
                    refreshToken = resp.refreshToken,
                    codeId = resp.codeId,
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun register(email: String, password: String): Result<AuthSessionResult> {
        return try {
            val client = ConnectTransportProvider.createProtocolClient()
            val displayName = email.substringBefore("@")
            val request = Auth.RegisterRequest.newBuilder()
                .setEmail(email)
                .setDisplayName(displayName)
                .setPassword(password)
                .setClient(buildClientInfo())
                .build()
                
            val spec = MethodSpec("antclaw.v1.AuthService/Register",
                Auth.RegisterRequest::class, Auth.RegisterResponse::class, StreamType.UNARY)
            val res: ResponseMessage<Auth.RegisterResponse> = client.unary(request, emptyMap(), spec)
            val resp = res.getOrThrow()

            Result.success(
                AuthSessionResult(
                    userId = resp.userId,
                    accessToken = resp.accessToken,
                    refreshToken = resp.refreshToken,
                    displayName = displayName,
                    codeId = resp.codeId,
                )
            )
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun refresh(): Result<String> {
        return try {
            val client = ConnectTransportProvider.createProtocolClient()
            val refreshToken = tokenStore.getRefreshToken()
                ?: return Result.failure(Exception("No refresh token available"))
            val request = Auth.RefreshRequest.newBuilder().setRefreshToken(refreshToken).build()
            val spec = MethodSpec("antclaw.v1.AuthService/Refresh",
                Auth.RefreshRequest::class, Auth.RefreshResponse::class, StreamType.UNARY)
            val res: ResponseMessage<Auth.RefreshResponse> = client.unary(request, emptyMap(), spec)
            val resp = res.getOrThrow()
            Result.success(resp.accessToken)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun logout() {
        try {
            val client = ConnectTransportProvider.createProtocolClient()
            val spec = MethodSpec("antclaw.v1.AuthService/Logout",
                Auth.LogoutRequest::class, Auth.LogoutResponse::class, StreamType.UNARY)
            client.unary(Auth.LogoutRequest.getDefaultInstance(), emptyMap(), spec)
        } catch (_: Exception) { }
    }

    suspend fun restoreToken(): String? {
        return tokenStore.getAccessToken()
    }

    /** 从持久存储恢复 userId（供 autoLogin 使用）。 */
    suspend fun restoredUserId(): String? {
        return tokenStore.getUserId()
    }
}
