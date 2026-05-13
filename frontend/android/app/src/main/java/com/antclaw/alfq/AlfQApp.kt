package com.antclaw.alfq

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.runtime.DisposableEffect
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.NavType
import androidx.navigation.navArgument
import com.antclaw.alfq.ui.feed.FeedScreen
import com.antclaw.alfq.ui.feed.SignalDetailScreen
import com.antclaw.alfq.ui.discover.DiscoverScreen
import com.antclaw.alfq.ui.post.PostScreen
import com.antclaw.alfq.ui.profile.MeScreen
import com.antclaw.alfq.ui.profile.MTAccountsScreen
import com.antclaw.alfq.ui.alerts.AlertScreen
import com.antclaw.alfq.ui.login.LoginScreen
import com.antclaw.alfq.ui.notification.NotificationCenterScreen
import com.antclaw.alfq.ui.notification.NotificationPrefsScreen
import com.antclaw.alfq.ui.notification.NotificationViewModel
import com.antclaw.alfq.ui.theme.AlfQTheme
import com.antclaw.alfq.navigation.BottomNavBarWithFAB
import com.antclaw.alfq.data.rpc.ConnectTransportProvider

@Composable
fun AlfQApp() {
    val isLoggedIn = remember { mutableStateOf(ConnectTransportProvider.getToken() != null) }

    AlfQTheme(darkTheme = false) {
        if (!isLoggedIn.value) {
            LoginScreen(
                onLoginSuccess = { token ->
                    ConnectTransportProvider.setTokenProvider { token }
                    isLoggedIn.value = true
                }
            )
        } else {
            MainContent(
                onLogout = {
                    ConnectTransportProvider.setTokenProvider { null }
                    isLoggedIn.value = false
                }
            )
        }
    }
}

@Composable
fun MainContent(onLogout: () -> Unit) {
    val navController = rememberNavController()
    val notifVm: NotificationViewModel = viewModel()
    val notifState by notifVm.state.collectAsState()

    // SSE 前后台生命周期管理
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_START -> notifVm.onForeground()
                Lifecycle.Event.ON_STOP -> notifVm.onBackground()
                else -> {}
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    Scaffold(
        bottomBar = {
            BottomNavBarWithFAB(
                navController = navController,
                notificationCount = notifState.unreadCount,
                onPostClick = { navController.navigate("post") },
                onChatClick = { navController.navigate("notifications") }
            )
        }
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = "feed",
            modifier = Modifier.padding(padding),
            builder = {
                composable("feed") {
                    FeedScreen(
                        notificationCount = notifState.unreadCount,
                        onSignalClick = { pair -> navController.navigate("signal/$pair") },
                        onNotificationClick = { navController.navigate("notifications") },
                    )
                }
                composable("discover") { DiscoverScreen() }
                composable("post") { PostScreen() }
                composable("me") {
                    MeScreen(
                        onLogout = onLogout,
                        onNavigateToMTAccounts = { navController.navigate("mt_accounts") },
                        onNavigateToAlerts = { navController.navigate("alerts") }
                    )
                }
                composable("mt_accounts") {
                    MTAccountsScreen(onBack = { navController.popBackStack() })
                }
                composable("alerts") {
                    AlertScreen(onBack = { navController.popBackStack() })
                }
                composable("notifications") {
                    NotificationCenterScreen(
                        onBack = { navController.popBackStack() },
                    )
                }
                composable("notification_prefs") {
                    NotificationPrefsScreen(
                        onBack = { navController.popBackStack() },
                    )
                }
                composable(
                    route = "signal/{pair}",
                    arguments = listOf(navArgument("pair") { type = NavType.StringType })
                ) { backStackEntry ->
                    val pair = backStackEntry.arguments?.getString("pair") ?: "EURUSD"
                    SignalDetailScreen(pair = pair, onBack = { navController.popBackStack() })
                }
            }
        )
    }
}
