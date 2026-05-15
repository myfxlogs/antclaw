package com.antclaw.alfq.data.repository

import antclaw.v1.Auth
import com.antclaw.alfq.data.device.DeviceInfoCollector
import com.antclaw.alfq.data.local.TokenStore
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.antclaw.alfq.data.sse.SseManager
import com.connectrpc.MethodSpec
import com.connectrpc.ResponseMessage
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import javax.inject.Inject

class AuthRepository @Inject constructor(
    private val tokenStore: TokenStore,
    private val deviceInfoCollector: DeviceInfoCollector,
    private val sseManager: SseManager,
    private val deviceRepository: DeviceRepository,
) {
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    suspend fun login(email: String, password: String): Result<String> {
        return try {
            val client = ConnectTransportProvider.createProtocolClient()
            
            val clientInfo = Auth.ClientInfo.newBuilder()
                .setDeviceId(deviceInfoCollector.getDeviceId())
                .setUserAgent(deviceInfoCollector.getUserAgent())
                .setIpAddress("")
                .build()
            
            val request = Auth.LoginRequest.newBuilder()
                .setEmail(email)
                .setPassword(password)
                .setClient(clientInfo)
                .build()
                
            val spec = MethodSpec("antclaw.v1.AuthService/Login",
                Auth.LoginRequest::class, Auth.LoginResponse::class, StreamType.UNARY)
            val res: ResponseMessage<Auth.LoginResponse> = client.unary(request, emptyMap(), spec)
            val resp = res.getOrThrow()
            val accessToken = resp.accessToken
            val refreshToken = resp.refreshToken
            ConnectTransportProvider.setToken(accessToken)
            tokenStore.saveTokens(accessToken, refreshToken, resp.userId)
            onLoginSuccess()
            Result.success(accessToken)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun register(email: String, password: String): Result<String> {
        return try {
            val client = ConnectTransportProvider.createProtocolClient()
            
            val clientInfo = Auth.ClientInfo.newBuilder()
                .setDeviceId(deviceInfoCollector.getDeviceId())
                .setUserAgent(deviceInfoCollector.getUserAgent())
                .setIpAddress("")
                .build()
            
            val displayName = email.substringBefore("@")
            val request = Auth.RegisterRequest.newBuilder()
                .setEmail(email)
                .setDisplayName(displayName)
                .setPassword(password)
                .setClient(clientInfo)
                .build()
                
            val spec = MethodSpec("antclaw.v1.AuthService/Register",
                Auth.RegisterRequest::class, Auth.RegisterResponse::class, StreamType.UNARY)
            val res: ResponseMessage<Auth.RegisterResponse> = client.unary(request, emptyMap(), spec)
            val resp = res.getOrThrow()
            ConnectTransportProvider.setToken(resp.accessToken)
            tokenStore.saveTokens(resp.accessToken, resp.refreshToken, resp.userId)
            onLoginSuccess()
            Result.success(resp.accessToken)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * 登录 / 注册成功后：建立 SSE 连接 + 异步上报设备信息。
     */
    private fun onLoginSuccess() {
        sseManager.connect()
        scope.launch {
            deviceRepository.reportDeviceInfo()
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
            ConnectTransportProvider.setToken(resp.accessToken)
            tokenStore.saveAccessToken(resp.accessToken)
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
        } catch (_: Exception) { } finally {
            sseManager.disconnect()
            ConnectTransportProvider.clearToken()
            tokenStore.clearTokens()
        }
    }

    suspend fun restoreToken(): String? {
        return tokenStore.getAccessToken()?.also { ConnectTransportProvider.setToken(it) }
    }
}