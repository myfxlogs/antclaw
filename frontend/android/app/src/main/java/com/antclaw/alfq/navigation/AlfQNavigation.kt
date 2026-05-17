package com.antclaw.alfq.navigation

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Email
import androidx.compose.foundation.layout.size
import androidx.compose.material3.*
import androidx.compose.ui.Modifier
import androidx.compose.runtime.*
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import androidx.navigation.compose.currentBackStackEntryAsState
import com.antclaw.alfq.R

/** X 风格底部导航栏 — 使用 Material3 NavigationBar + NavigationBarItem。 */
@Composable
fun BottomNavBar(
    navController: NavController,
    notificationCount: Int = 0,
    onPostClick: () -> Unit = {},
) {
    val backStack by navController.currentBackStackEntryAsState()
    val currentRoute = backStack?.destination?.route ?: "feed"

    NavigationBar(containerColor = MaterialTheme.colorScheme.background) {
        NavigationBarItem(
            selected = currentRoute == "feed",
            onClick = {
                if (currentRoute != "feed") {
                    navController.navigate("feed") {
                        popUpTo("feed") { inclusive = false; saveState = true }
                        launchSingleTop = true
                        restoreState = true
                    }
                }
            },
            icon = { Icon(Icons.Default.Home, contentDescription = stringResource(R.string.nav_home)) },
            label = { Text(stringResource(R.string.nav_home)) },
        )
        NavigationBarItem(
            selected = currentRoute == "discover",
            onClick = {
                if (currentRoute != "discover") {
                    navController.navigate("discover") {
                        popUpTo("feed") { inclusive = false; saveState = true }
                        launchSingleTop = true
                        restoreState = true
                    }
                }
            },
            icon = { Icon(Icons.Default.Search, contentDescription = stringResource(R.string.nav_discover)) },
            label = { Text(stringResource(R.string.nav_discover)) },
        )

        // 发贴 FAB（居中，使用 NavigationBarItem 保持无障碍语义）
        NavigationBarItem(
            selected = currentRoute == "post",
            onClick = onPostClick,
            icon = {
                SmallFloatingActionButton(
                    onClick = onPostClick,
                    containerColor = MaterialTheme.colorScheme.primary,
                    contentColor = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.size(40.dp),
                ) { Icon(Icons.Default.Add, contentDescription = stringResource(R.string.nav_post), modifier = Modifier.size(20.dp)) }
            },
            label = { Text(stringResource(R.string.nav_post)) },
        )

        NavigationBarItem(
            selected = currentRoute == "notifications",
            onClick = {
                if (currentRoute != "notifications") {
                    navController.navigate("notifications") {
                        popUpTo("feed") { inclusive = false; saveState = true }
                        launchSingleTop = true
                        restoreState = true
                    }
                }
            },
            icon = {
                BadgedBox(badge = { if (notificationCount > 0) Badge { Text("$notificationCount") } }) {
                    Icon(Icons.Default.Notifications, contentDescription = stringResource(R.string.nav_notifications))
                }
            },
            label = { Text(stringResource(R.string.nav_notifications)) },
        )
        NavigationBarItem(
            selected = currentRoute == "chat",
            onClick = {
                if (currentRoute != "chat") {
                    navController.navigate("chat") {
                        popUpTo("feed") { inclusive = false; saveState = true }
                        launchSingleTop = true
                        restoreState = true
                    }
                }
            },
            icon = { Icon(Icons.Default.Email, contentDescription = stringResource(R.string.nav_messages)) },
            label = { Text(stringResource(R.string.nav_messages)) },
        )
    }
}
