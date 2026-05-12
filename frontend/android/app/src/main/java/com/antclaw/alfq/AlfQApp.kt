package com.antclaw.alfq

import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
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
import com.antclaw.alfq.ui.theme.AlfQTheme
import com.antclaw.alfq.navigation.BottomNavBar
import com.antclaw.alfq.data.rpc.ConnectTransportProvider

@Composable
fun AlfQApp() {
    val isLoggedIn = remember { mutableStateOf(ConnectTransportProvider.getToken() != null) }

    AlfQTheme(darkTheme = true) {
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

    NavHost(
        navController = navController,
        startDestination = "feed",
        builder = {
            composable("feed") {
                FeedScreen(onSignalClick = { pair -> navController.navigate("signal/$pair") })
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
            composable(
                route = "signal/{pair}",
                arguments = listOf(navArgument("pair") { type = NavType.StringType })
            ) { backStackEntry ->
                val pair = backStackEntry.arguments?.getString("pair") ?: "EURUSD"
                SignalDetailScreen(pair = pair, onBack = { navController.popBackStack() })
            }
        }
    )

    BottomNavBar(navController = navController)
}
