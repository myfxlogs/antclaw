package com.antclaw.alfq.navigation

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Email
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import androidx.navigation.compose.currentBackStackEntryAsState
import com.antclaw.alfq.R

/** X 风格底部栏：首页 | 发现 | 发贴 | 通知 | 消息 */
@Composable
fun BottomNavBar(
    navController: NavController,
    notificationCount: Int = 0,
    onPostClick: () -> Unit = {},
) {
    val backStack by navController.currentBackStackEntryAsState()
    val currentRoute = backStack?.destination?.route ?: "feed"

    val navigateTab: (String) -> Unit = { route ->
        if (currentRoute != route) {
            navController.navigate(route) {
                popUpTo("feed") { inclusive = false; saveState = true }
                launchSingleTop = true
                restoreState = true
            }
        }
    }

    Surface(color = MaterialTheme.colorScheme.background) {
        Row(
            modifier = Modifier.fillMaxWidth().height(56.dp).padding(horizontal = 8.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            NavItem(Icons.Default.Home, stringResource(R.string.nav_home), currentRoute == "feed") { navigateTab("feed") }
            NavItem(Icons.Default.Search, stringResource(R.string.nav_discover), currentRoute == "discover") { navigateTab("discover") }

            // 发贴 FAB
            SmallFloatingActionButton(
                onClick = onPostClick,
                containerColor = MaterialTheme.colorScheme.primary,
                contentColor = MaterialTheme.colorScheme.onPrimary,
                modifier = Modifier.size(40.dp),
            ) { Icon(Icons.Default.Add, contentDescription = stringResource(R.string.nav_post), modifier = Modifier.size(20.dp)) }

            NavItem(Icons.Default.Notifications, stringResource(R.string.nav_notifications), currentRoute == "notifications") { navigateTab("notifications") }
            NavItem(Icons.Default.Email, stringResource(R.string.nav_messages), currentRoute == "chat") { navigateTab("chat") }
        }
    }
}

@Composable
private fun NavItem(icon: ImageVector, label: String, selected: Boolean, onClick: () -> Unit) {
    Column(
        modifier = Modifier.width(48.dp).clickable(onClick = onClick).padding(vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        Icon(icon, contentDescription = label, modifier = Modifier.size(24.dp),
            tint = if (selected) MaterialTheme.colorScheme.onSurface else MaterialTheme.colorScheme.onSurfaceVariant)
        Text(label, style = MaterialTheme.typography.labelSmall,
            color = if (selected) MaterialTheme.colorScheme.onSurface else MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
