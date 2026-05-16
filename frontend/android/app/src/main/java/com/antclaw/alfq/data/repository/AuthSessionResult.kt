package com.antclaw.alfq.data.repository

/**
 * 登录/注册成功后返回的完整会话凭据，供 SessionViewModel 统一消费。
 */
data class AuthSessionResult(
    val userId: String,
    val accessToken: String,
    val refreshToken: String,
    val displayName: String = "",
    val codeId: String = "",
)
