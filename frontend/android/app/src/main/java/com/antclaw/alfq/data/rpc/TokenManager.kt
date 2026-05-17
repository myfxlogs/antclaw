package com.antclaw.alfq.data.rpc

import com.antclaw.alfq.data.session.SessionExpiredNotifier
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Token 生命周期 facade — 组合 [TokenSnapshot] + [RefreshTokenCoordinator]。
 *
 * 为 SessionViewModel 等调用方提供简洁 API，内部委托给子组件。
 * 若调用方只需读 token，可直接注入 [TokenSnapshot]。
 */
@Singleton
class TokenManager @Inject constructor(
    private val snapshot: TokenSnapshot,
    private val refreshCoordinator: RefreshTokenCoordinator,
) {
    fun getToken(): String? = snapshot.getToken()
    fun setToken(token: String) = snapshot.setToken(token)
    fun clearToken() = snapshot.clearToken()
    fun restore(notifier: SessionExpiredNotifier?) = refreshCoordinator.restore(notifier)
}
