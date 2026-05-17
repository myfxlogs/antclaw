package com.antclaw.alfq.navigation

/**
 * 集中管理所有导航路由字符串。
 * 新增页面路由只需在此处定义，避免散落在多个文件。
 */
object Routes {
    const val FEED = "feed"
    const val DISCOVER = "discover"
    const val POST = "post"
    const val POST_DETAIL = "postDetail/{postId}"
    const val ME = "me"
    const val MT_ACCOUNTS = "mt_accounts"
    const val BIND_MT_ACCOUNT = "bind_mt_account"
    const val ALERTS = "alerts"
    const val CHAT = "chat"
    const val NOTIFICATIONS = "notifications"
    const val NOTIFICATION_PREFS = "notification_prefs"
    const val SETTINGS_LANGUAGE = "settings/language"
    const val DEVICE_INFO = "device_info"
    const val SSE_DEBUG = "sse_debug"
    const val SIGNAL = "signal/{pair}"
    const val PROFILE = "profile/{userId}"
    const val LOGIN = "login"
    const val REGISTER = "register"

    // 带参数路由的构建方法
    fun postDetail(postId: String) = "postDetail/$postId"
    fun signal(pair: String) = "signal/$pair"
    fun profile(userId: String) = "profile/$userId"
}
