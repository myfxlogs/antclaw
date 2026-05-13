package com.antclaw.alfq.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.text.font.FontWeight
import com.antclaw.alfq.ui.theme.BullGreen

/** 交易统计数值 + 标签列，用于个人主页和交易员档案。 */
@Composable
fun StatCell(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(value, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, color = BullGreen)
        Text(label, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.5f))
    }
}

/** 关注/粉丝/胜率 三列行，MeScreen 和 ProfileScreen 共享。 */
@Composable
fun TraderStatRow(
    followingLabel: String, followingCount: Int,
    followersLabel: String, followerCount: Int,
    winRateLabel: String = "", winRateValue: String = "",
) {
    Row(horizontalArrangement = Arrangement.spacedBy(24.dp)) {
        Column {
            Text("$followingCount", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Text(followingLabel, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
        Column {
            Text("$followerCount", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Text(followersLabel, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
        }
        if (winRateValue.isNotEmpty()) {
            Column {
                Text(winRateValue, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                Text(winRateLabel, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}
