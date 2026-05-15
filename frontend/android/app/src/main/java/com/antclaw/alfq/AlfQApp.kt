package com.antclaw.alfq

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.NavType
import androidx.navigation.navArgument
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
import com.antclaw.alfq.ui.mt.BindMtAccountScreen
import com.antclaw.alfq.ui.theme.AlfQTheme
import com.antclaw.alfq.navigation.BottomNavBarWithFAB
import com.antclaw.alfq.data.rpc.ConnectTransportProvider

// ── Constants ──
private const val DEFAULT_PAIR = "EURUSD"
private const val DEFAULT_USER_ID = "me"

@Composable
fun AlfQApp() {
    val isLoggedIn = remember { mutableStateOf(ConnectTransportProvider.getToken() != null) }
    AlfQTheme(darkTheme = isSystemInDarkTheme()) {
        if (!isLoggedIn.value) {
            val authNavController = rememberNavController()
            NavHost(navController = authNavController, startDestination = "login") {
                composable("login") {
                    LoginScreen(
                        onLoginSuccess = { token ->
                            ConnectTransportProvider.setToken(token)
                            isLoggedIn.value = true
                        },
                        onRegisterClick = { authNavController.navigate("register") },
                    )
                }
                composable("register") {
                    RegisterScreen(
                        onBack = { authNavController.popBackStack() },
                        onRegisterSuccess = { token ->
                            ConnectTransportProvider.setToken(token)
                            isLoggedIn.value = true
                        },
                    )
                }
            }
        } else {
            MainContent(onLogout = { isLoggedIn.value = false })
        }
    }
}

@Composable
fun MainContent(onLogout: () -> Unit) {
    val navController = rememberNavController()
    val notifVm: NotificationViewModel = hiltViewModel()
    val notifState by notifVm.state.collectAsState()
    LifecycleAware({ notifVm.setForeground(true) }, { notifVm.setForeground(false) })

    Scaffold(
        bottomBar = {
            BottomNavBarWithFAB(
                navController = navController,
                notificationCount = notifState.unreadCount,
                onChatClick = { navController.navigate("chat") }
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
                )
            }
            composable("discover") {
                DiscoverScreen(onTraderClick = { userId -> navController.navigate("profile/$userId") })
            }
            composable("social") {
                FeedScreen(
                    notificationCount = notifState.unreadCount,
                    onPostClick = { postId -> navController.navigate("postDetail/$postId") },
                    onAuthorClick = { userId -> navController.navigate("profile/$userId") },
                    onNotificationClick = { navController.navigate("notifications") },
                    onSearchClick = { navController.navigate("discover") },
                )
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
                MeScreen(
                    onLogout = onLogout,
                    onNavigateToMTAccounts = { navController.navigate("mt_accounts") },
                    onNavigateToAlerts = { navController.navigate("alerts") },
                    onNavigateToSettings = { navController.navigate("settings/language") },
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
            composable("notifications") { NotificationCenterScreen(onBack = { navController.popBackStack() }) }
            composable("notification_prefs") { NotificationPrefsScreen(onBack = { navController.popBackStack() }) }
            composable("settings/language") { LanguagePickerScreen(onBack = { navController.popBackStack() }) }
            composable("device_info") { DeviceInfoScreen() }
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
                ProfileScreen(userId = userId, onBack = { navController.popBackStack() })
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
