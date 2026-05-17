package com.antclaw.alfq.data.rpc

import javax.inject.Inject
import javax.inject.Singleton

/**
 * Token 内存快照 — 线程安全的读写持有者。
 *
 * 不涉及持久化、刷新、通知。仅负责 get/set/clear。
 */
@Singleton
class TokenSnapshot @Inject constructor() {
    @Volatile var value: String? = null

    fun getToken(): String? = value
    fun setToken(token: String) { value = token }
    fun clearToken() { value = null }
}
