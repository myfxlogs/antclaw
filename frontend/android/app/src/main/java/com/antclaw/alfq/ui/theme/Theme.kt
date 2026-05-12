package com.antclaw.alfq.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

// -- 交易员暗色配色 --
private val DarkBackground = Color(0xFF0D1117)
private val DarkSurface = Color(0xFF161B22)
private val DarkSurfaceVariant = Color(0xFF21262D)
private val BullGreen = Color(0xFF26A69A)
private val BearRed = Color(0xFFEF5350)
private val GoldAccent = Color(0xFFFFC107)

private val AlfQDarkColors = darkColorScheme(
    primary = BullGreen,
    secondary = GoldAccent,
    tertiary = BearRed,
    background = DarkBackground,
    surface = DarkSurface,
    surfaceVariant = DarkSurfaceVariant,
    onPrimary = Color.Black,
    onSecondary = Color.Black,
    onTertiary = Color.White,
    onBackground = Color(0xFFE6EDF3),
    onSurface = Color(0xFFE6EDF3),
    outline = Color(0xFF30363D),
)

private val AlfQLightColors = lightColorScheme(
    primary = Color(0xFF00897B),
    secondary = Color(0xFFFFA000),
    tertiary = BearRed,
    background = Color(0xFFF6F8FA),
    surface = Color.White,
    surfaceVariant = Color(0xFFE8ECF0),
    onPrimary = Color.White,
    onSecondary = Color.Black,
    onTertiary = Color.White,
    onBackground = Color(0xFF1F2328),
    onSurface = Color(0xFF1F2328),
    outline = Color(0xFFD0D7DE),
)

@Composable
fun AlfQTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val colorScheme = if (darkTheme) AlfQDarkColors else AlfQLightColors
    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography(),
        content = content
    )
}
