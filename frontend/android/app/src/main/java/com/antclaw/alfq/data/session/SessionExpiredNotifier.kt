package com.antclaw.alfq.data.session

import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Session 过期通知总线 — 解耦 RefreshTokenCoordinator（无协程上下文）与 ViewModel 层。
 *
 * 用法：
 *   RefreshTokenCoordinator.refresh 失败 → notifier.notifySessionExpired()
 *   SessionViewModel 订阅 → 收到事件 → onSessionExpired() → emit(RequireLogin)
 */
@Singleton
open class SessionExpiredNotifier @Inject constructor() {
    private val _events = MutableSharedFlow<Unit>(extraBufferCapacity = 4)
    val events: SharedFlow<Unit> = _events.asSharedFlow()

    open fun notifySessionExpired() {
        _events.tryEmit(Unit)
    }
}
