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
import androidx.compose.material.icons.filled.ChatBubbleOutline
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import androidx.navigation.compose.currentBackStackEntryAsState

data class NavTab(val route: String, val label: String, val icon: ImageVector)

val tabs = listOf(
    NavTab("feed", "首页", Icons.Default.Home),
    NavTab("discover", "发现", Icons.Default.Search),
    NavTab("me", "我的", Icons.Default.Person),
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
            // Home
            NavigationItem(
                selected = currentRoute == "feed",
                onClick = {
                    if (currentRoute != "feed") {
                        navController.navigate("feed") {
                            popUpTo("feed") { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                },
                icon = { Icon(Icons.Default.Home, contentDescription = "首页") },
                label = "首页"
            )

            // Discover
            NavigationItem(
                selected = currentRoute == "discover",
                onClick = {
                    if (currentRoute != "discover") {
                        navController.navigate("discover") {
                            popUpTo("feed") { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                },
                icon = { Icon(Icons.Default.Search, contentDescription = "发现") },
                label = "发现"
            )

            // Post - FAB
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
                    contentDescription = "发布",
                    tint = MaterialTheme.colorScheme.onPrimary,
                    modifier = Modifier.size(24.dp)
                )
            }

            // Chat / 通知
            Box(modifier = Modifier.width(56.dp).clickable(onClick = onChatClick).padding(vertical = 8.dp)) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Box(modifier = Modifier.size(24.dp), contentAlignment = Alignment.Center) {
                        Icon(Icons.Default.ChatBubbleOutline, contentDescription = "消息")
                    }
                    Text(
                        "消息",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                if (notificationCount > 0) {
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .offset(x = 6.dp, y = (-2).dp)
                            .background(Color(0xFFE53935), RoundedCornerShape(10.dp))
                            .padding(horizontal = 5.dp, vertical = 1.dp),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = if (notificationCount > 99) "99+" else notificationCount.toString(),
                            color = Color.White,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold,
                        )
                    }
                }
            }

            // Me
            NavigationItem(
                selected = currentRoute == "me",
                onClick = {
                    if (currentRoute != "me") {
                        navController.navigate("me") {
                            popUpTo("feed") { saveState = true }
                            launchSingleTop = true
                            restoreState = true
                        }
                    }
                },
                icon = { Icon(Icons.Default.Person, contentDescription = "我的") },
                label = "我的"
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
        ) {
            icon()
        }
        Text(
            label,
            style = MaterialTheme.typography.labelSmall,
            color = if (selected) MaterialTheme.colorScheme.onSurface else MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
fun BottomNavBar(navController: NavController, notificationCount: Int = 0) {
    BottomNavBarWithFAB(
        navController = navController,
        notificationCount = notificationCount
    )
}
