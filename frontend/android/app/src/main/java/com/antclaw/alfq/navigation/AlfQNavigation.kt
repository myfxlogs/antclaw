package com.antclaw.alfq.navigation

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Email
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import androidx.navigation.compose.currentBackStackEntryAsState
import com.antclaw.alfq.R

private data class NavTab(val route: String, val labelRes: Int, val icon: ImageVector)

private val mainTabs = listOf(
    NavTab("feed", R.string.nav_home, Icons.Default.Home),
    NavTab("discover", R.string.nav_discover, Icons.Default.Search),
)

@Composable
fun BottomNavBarWithFAB(
    navController: NavController,
    notificationCount: Int = 0,
    onPostClick: () -> Unit = {},
    onChatClick: () -> Unit = {}
) {
    val backStack by navController.currentBackStackEntryAsState()
    val currentRoute = backStack?.destination?.route ?: "feed"

    val navigateTab: (String) -> Unit = { route ->
        if (currentRoute != route) {
            navController.navigate(route) {
                popUpTo("feed") { saveState = true }
                launchSingleTop = true
                restoreState = true
            }
        }
    }

    Surface(
        color = MaterialTheme.colorScheme.background,
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp)
                .padding(horizontal = 8.dp),
            horizontalArrangement = Arrangement.SpaceEvenly,
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Order: 首页 发现 +
            mainTabs.forEach { tab ->
                NavigationItem(
                    selected = currentRoute == tab.route,
                    onClick = { navigateTab(tab.route) },
                    icon = { Icon(tab.icon, contentDescription = stringResource(tab.labelRes)) },
                    label = stringResource(tab.labelRes)
                )
            }

            // FAB post button
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.primary)
                    .clickable(onClick = onPostClick),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Default.Add,
                    contentDescription = stringResource(R.string.nav_post),
                    tint = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.size(24.dp)
                )
            }

            // 消息
            NavigationItem(
                selected = currentRoute == "notifications",
                onClick = onChatClick,
                icon = { Icon(Icons.Default.Email, contentDescription = stringResource(R.string.nav_messages)) },
                label = stringResource(R.string.nav_messages),
            )

            // 我的
            NavigationItem(
                selected = currentRoute == "me",
                onClick = { navigateTab("me") },
                icon = { Icon(Icons.Default.Person, contentDescription = stringResource(R.string.nav_me)) },
                label = stringResource(R.string.nav_me),
            )
        }
    }
}

@Composable
private fun NavigationItem(
    selected: Boolean,
    onClick: () -> Unit,
    icon: @Composable () -> Unit,
    label: String
) {
    Column(
        modifier = Modifier
            .width(56.dp)
            .clickable(onClick = onClick)
            .padding(vertical = 8.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Box(
            modifier = Modifier.size(24.dp),
            contentAlignment = Alignment.Center
        ) { icon() }
        Text(
            label,
            style = MaterialTheme.typography.labelSmall,
            color = if (selected) MaterialTheme.colorScheme.onSurface else MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

