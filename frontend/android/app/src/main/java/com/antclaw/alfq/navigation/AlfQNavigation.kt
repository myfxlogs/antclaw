package com.antclaw.alfq.navigation

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Home
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.navigation.NavController
import androidx.navigation.compose.currentBackStackEntryAsState

data class NavTab(val route: String, val label: String, val icon: ImageVector)

val tabs = listOf(
    NavTab("feed", "Feed", Icons.Default.Home),
    NavTab("discover", "发现", Icons.Default.Search),
    NavTab("post", "发布", Icons.Default.Add),
    NavTab("me", "我", Icons.Default.Person),
)

@Composable
fun BottomNavBar(navController: NavController, notificationCount: Int = 0) {
    val backStack by navController.currentBackStackEntryAsState()
    val currentRoute = backStack?.destination?.route ?: "feed"

    NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
        tabs.forEach { tab ->
            NavigationBarItem(
                selected = currentRoute == tab.route,
                onClick = {
                    if (currentRoute != tab.route) {
                        navController.navigate(tab.route) {
                            popUpTo("feed") { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                },
                icon = {
                    if (tab.route == "feed" && notificationCount > 0) {
                        BadgedBox(badge = {
                            Badge { Text("$notificationCount") }
                        }) {
                            Icon(tab.icon, contentDescription = tab.label)
                        }
                    } else {
                        Icon(tab.icon, contentDescription = tab.label)
                    }
                },
                label = { Text(tab.label) },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = MaterialTheme.colorScheme.primary,
                    unselectedIconColor = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f),
                    selectedTextColor = MaterialTheme.colorScheme.primary,
                    indicatorColor = MaterialTheme.colorScheme.primary.copy(alpha = 0.15f),
                )
            )
        }
    }
}
