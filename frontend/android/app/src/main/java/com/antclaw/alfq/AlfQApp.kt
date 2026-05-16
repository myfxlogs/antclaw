package com.antclaw.alfq

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.NavType
import androidx.navigation.navArgument
import com.antclaw.alfq.data.repository.AuthSessionResult
import com.antclaw.alfq.ui.chat.ChatScreen
import com.antclaw.alfq.ui.feed.FeedScreen
import com.antclaw.alfq.ui.feed.SignalDetailScreen
import com.antclaw.alfq.ui.post.PostDetailScreen
import com.antclaw.alfq.ui.discover.DiscoverScreen
import com.antclaw.alfq.ui.post.PostScreen
import com.antclaw.alfq.ui.profile.MeScreen
import com.antclaw.alfq.ui.profile.MTAccountsScreen
import com.antclaw.alfq.ui.profile.ProfileScreen
import com.antclaw.alfq.ui.alerts.AlertScreen
import com.antclaw.alfq.ui.login.LoginScreen
import com.antclaw.alfq.ui.login.RegisterScreen
import com.antclaw.alfq.ui.notification.NotificationCenterScreen
import com.antclaw.alfq.ui.notification.NotificationPrefsScreen
import com.antclaw.alfq.ui.notification.NotificationViewModel
import com.antclaw.alfq.ui.settings.LanguagePickerScreen
import com.antclaw.alfq.ui.device.DeviceInfoScreen
import com.antclaw.alfq.ui.debug.SseDebugScreen
import com.antclaw.alfq.ui.mt.BindMtAccountScreen
import com.antclaw.alfq.ui.session.SessionState
import com.antclaw.alfq.ui.session.SessionViewModel
import com.antclaw.alfq.ui.theme.AlfQTheme
import com.antclaw.alfq.navigation.BottomNavBar

// ── Constants ──
private const val DEFAULT_PAIR = "EURUSD"
private const val DEFAULT_USER_ID = "me"

@Composable
fun AlfQApp() {
    val sessionVm: SessionViewModel = hiltViewModel()
    val session by sessionVm.session.collectAsStateWithLifecycle()

    AlfQTheme(darkTheme = isSystemInDarkTheme()) {
        when (session.state) {
            SessionState.AUTHENTICATED -> MainContent(
                onLogout = { sessionVm.logout() },
                sessionVm = sessionVm,
            )
            else -> AuthContent(
                onLoginSuccess = { result ->
                    sessionVm.onLoginSuccess(
                        userId = result.userId,
                        accessToken = result.accessToken,
                        refreshToken = result.refreshToken,
                        displayName = result.displayName,
                    )
                },
            )
        }
    }
}

@Composable
fun AuthContent(onLoginSuccess: (AuthSessionResult) -> Unit) {
    val authNavController = rememberNavController()
    NavHost(navController = authNavController, startDestination = "login") {
        composable("login") {
            LoginScreen(
                onLoginSuccess = onLoginSuccess,
                onRegisterClick = { authNavController.navigate("register") },
            )
        }
        composable("register") {
            RegisterScreen(
                onBack = { authNavController.popBackStack() },
                onRegisterSuccess = onLoginSuccess,
            )
        }
    }
}

@Composable
fun MainContent(onLogout: () -> Unit, sessionVm: SessionViewModel) {
    val navController = rememberNavController()
    val notifVm: NotificationViewModel = hiltViewModel()
    val notifState by notifVm.state.collectAsState()
    val session by sessionVm.session.collectAsStateWithLifecycle()
    LifecycleAware(
        onStart = {
            sessionVm.onForeground()
            notifVm.setForeground(true)
        },
        onStop = {
            sessionVm.onBackground()
            notifVm.setForeground(false)
        },
    )

    Scaffold(
        bottomBar = {
            BottomNavBar(
                navController = navController,
                notificationCount = notifState.unreadCount,
                onPostClick = { navController.navigate("post") },
            )
        }
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = "feed",
            modifier = Modifier.padding(padding),
        ) {
            composable("feed") {
                FeedScreen(
                    notificationCount = notifState.unreadCount,
                    onPostClick = { postId -> navController.navigate("postDetail/$postId") },
                    onAuthorClick = { userId -> navController.navigate("profile/$userId") },
                    onNotificationClick = { navController.navigate("notifications") },
                    onSearchClick = { navController.navigate("discover") },
                    onMeClick = { navController.navigate("me") },
                )
            }
            composable("discover") {
                DiscoverScreen(onTraderClick = { userId -> navController.navigate("profile/$userId") })
            }
            composable(
                route = "postDetail/{postId}",
                arguments = listOf(navArgument("postId") { type = NavType.StringType })
            ) { backStackEntry ->
                val postId = backStackEntry.arguments?.getString("postId") ?: ""
                PostDetailScreen(postId = postId, onBack = { navController.popBackStack() })
            }
            composable("post") { PostScreen(onClose = { navController.popBackStack() }) }
            composable("me") {
                ProfileScreen(
                    userId = session.userId.ifEmpty { "me" },
                    onBack = { },
                    onPostClick = { postId -> navController.navigate("postDetail/$postId") },
                    onSettingsClick = { navController.navigate("settings/language") },
                )
            }
            composable("mt_accounts") {
                MTAccountsScreen(
                    onBack = { navController.popBackStack() },
                    onBindClick = { navController.navigate("bind_mt_account") }
                )
            }
            composable("bind_mt_account") { BindMtAccountScreen(onBack = { navController.popBackStack() }) }
            composable("alerts") { AlertScreen(onBack = { navController.popBackStack() }) }
            composable("chat") { ChatScreen(onBack = { navController.popBackStack() }) }
            composable("notifications") {
                NotificationCenterScreen(
                    onBack = { navController.popBackStack() },
                    onNotificationClick = { notif ->
                        val d = notif.data
                        when {
                            d["post_id"] != null -> navController.navigate("postDetail/${d["post_id"]}")
                            d["user_id"] != null -> navController.navigate("profile/${d["user_id"]}")
                            d["signal_id"] != null -> navController.navigate("signal/${d["signal_id"]}")
                        }
                    },
                )
            }
            composable("notification_prefs") { NotificationPrefsScreen(onBack = { navController.popBackStack() }) }
            composable("settings/language") { LanguagePickerScreen(onBack = { navController.popBackStack() }) }
            composable("device_info") { DeviceInfoScreen() }
            composable("sse_debug") { SseDebugScreen(onBack = { navController.popBackStack() }) }
            composable(
                route = "signal/{pair}",
                arguments = listOf(navArgument("pair") { type = NavType.StringType })
            ) { backStackEntry ->
                val pair = backStackEntry.arguments?.getString("pair") ?: DEFAULT_PAIR
                SignalDetailScreen(pair = pair, onBack = { navController.popBackStack() })
            }
            composable(
                route = "profile/{userId}",
                arguments = listOf(navArgument("userId") { type = NavType.StringType })
            ) { backStackEntry ->
                val userId = backStackEntry.arguments?.getString("userId") ?: DEFAULT_USER_ID
                ProfileScreen(userId = userId, onBack = { navController.popBackStack() }, onPostClick = { postId -> navController.navigate("postDetail/$postId") })
            }
        }
    }
}

/**
 * Reusable lifecycle-aware effect — observes START/STOP for foreground/background callbacks.
 */
@Composable
private fun LifecycleAware(onStart: () -> Unit, onStop: () -> Unit) {
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_START -> onStart()
                Lifecycle.Event.ON_STOP -> onStop()
                else -> {}
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }
}
