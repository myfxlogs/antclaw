package com.antclaw.alfq

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
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
import com.antclaw.alfq.ui.login.RegisterScreen
import com.antclaw.alfq.ui.notification.NotificationCenterScreen
import com.antclaw.alfq.ui.notification.NotificationPrefsScreen
import com.antclaw.alfq.ui.notification.NotificationViewModel
import com.antclaw.alfq.ui.settings.LanguagePickerScreen
import com.antclaw.alfq.ui.theme.AlfQTheme
import com.antclaw.alfq.navigation.BottomNavBarWithFAB
import com.antclaw.alfq.data.rpc.ConnectTransportProvider

@Composable
fun AlfQApp() {
    val isLoggedIn = remember { mutableStateOf(ConnectTransportProvider.getToken() != null) }
    AlfQTheme(darkTheme = false) {
        if (!isLoggedIn.value) {
            val authNavController = rememberNavController()
            NavHost(navController = authNavController, startDestination = "login") {
                composable("login") {
                    LoginScreen(
                        onLoginSuccess = { token -> ConnectTransportProvider.setToken(token); isLoggedIn.value = true },
                        onRegisterClick = { authNavController.navigate("register") },
                    )
                }
                composable("register") {
                    RegisterScreen(
                        onBack = { authNavController.popBackStack() },
                        onRegisterSuccess = { token -> ConnectTransportProvider.setToken(token); isLoggedIn.value = true },
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
    val notifVm: NotificationViewModel = viewModel()
    val notifState by notifVm.state.collectAsState()
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
                        onSearchClick = { navController.navigate("discover") },
                    )
                }
                composable("discover") {
                    DiscoverScreen(onTraderClick = { userId -> navController.navigate("profile/$userId") })
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
                    MTAccountsScreen(onBack = { navController.popBackStack() }, onBindClick = { navController.navigate("bind_mt_account") })
                }
                composable("bind_mt_account") {
                    com.antclaw.alfq.ui.mt.BindMtAccountScreen(onBack = { navController.popBackStack() })
                }
                composable("alerts") { AlertScreen(onBack = { navController.popBackStack() }) }
                composable("notifications") { NotificationCenterScreen(onBack = { navController.popBackStack() }) }
                composable("notification_prefs") { NotificationPrefsScreen(onBack = { navController.popBackStack() }) }
                composable("settings/language") { LanguagePickerScreen(onBack = { navController.popBackStack() }) }
                composable(route = "signal/{pair}", arguments = listOf(navArgument("pair") { type = NavType.StringType })) { backStackEntry ->
                    val pair = backStackEntry.arguments?.getString("pair") ?: "EURUSD"
                    SignalDetailScreen(pair = pair, onBack = { navController.popBackStack() })
                }
                composable(route = "profile/{userId}", arguments = listOf(navArgument("userId") { type = NavType.StringType })) { backStackEntry ->
                    val userId = backStackEntry.arguments?.getString("userId") ?: "me"
                    com.antclaw.alfq.ui.profile.ProfileScreen(userId = userId, onBack = { navController.popBackStack() })
                }
            }
        )
    }
}
