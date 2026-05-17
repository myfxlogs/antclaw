package com.antclaw.alfq.ui.theme

import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

// ==================== anttrader 品牌色 ====================
// 来源：/opt/antclaw/emulator/anttrader/frontend 色彩基准
val Primary       = Color(0xFFD4AF37)   // 主品牌金
val PrimaryDark   = Color(0xFFB8960B)   // 深金
val PrimaryLight  = Color(0xFFE6C65C)   // 浅金
val AccentDark    = Color(0xFF141D22)   // 深石墨（主文本 / 强调色）

// 背景色（浅色）
val BgMain        = Color(0xFFFFFFFF)   // 主背景
val BgCard        = Color(0xFFFFFFFF)   // 卡片背景
val BgSecondary   = Color(0xFFF5F7F9)   // 次级背景
val BgTertiary    = Color(0xFFE8ECF0)   // 三级背景

// 文本色（浅色）
val TextPrimary   = Color(0xFF141D22)   // 正文
val TextSecondary = Color(0xFF5A6B75)   // 辅助信息
val TextMuted     = Color(0xFF8A9AA5)   // 弱文本

// 功能色
val SuccessGreen  = Color(0xFF00A651)   // 成功 / 看涨
val ErrorRed      = Color(0xFFE53935)   // 错误 / 强风险
val InfoBlue      = Color(0xFF2196F3)   // 信息提示
val WarnOrange    = Color(0xFFFFA000)   // 警告

// 信号方向色（看涨/看跌与 SuccessGreen/ErrorRed 统一）
val BullGreen = SuccessGreen
val BearRed   = ErrorRed

// ── 深色背景 ──
private val DarkBgMain      = Color(0xFF0D1117)
private val DarkBgCard      = Color(0xFF161B22)
private val DarkBgSecondary = Color(0xFF1C2128)
private val DarkBgTertiary  = Color(0xFF21262D)
private val DarkTextPrimary = Color(0xFFE6EDF3)
private val DarkTextSecondary = Color(0xFF8B949E)
private val DarkSurface     = Color(0xFF161B22)

// ==================== 浅色主题 ====================

private val AlfQLightColors = lightColorScheme(
    primary              = Primary,
    onPrimary            = Color.White,
    primaryContainer     = PrimaryLight,
    onPrimaryContainer   = AccentDark,
    secondary            = AccentDark,
    onSecondary          = Color.White,
    secondaryContainer   = BgSecondary,
    onSecondaryContainer = TextPrimary,
    tertiary             = WarnOrange,
    onTertiary           = Color.White,
    tertiaryContainer    = Color(0xFFFFE0B2),
    onTertiaryContainer  = Color(0xFFE65100),
    background           = BgMain,
    onBackground         = TextPrimary,
    surface              = Color.White,
    onSurface            = TextPrimary,
    surfaceVariant       = BgSecondary,
    onSurfaceVariant     = TextSecondary,
    outline              = Color(0x14000000),
    outlineVariant       = BgTertiary,
    error                = ErrorRed,
    onError              = Color.White,
)

// ==================== 深色主题 ====================

private val AlfQDarkColors = darkColorScheme(
    primary              = PrimaryLight,
    onPrimary            = AccentDark,
    primaryContainer     = PrimaryDark,
    onPrimaryContainer   = Color(0xFFFFF3C4),
    secondary            = Color(0xFF8B949E),
    onSecondary          = AccentDark,
    secondaryContainer   = DarkBgSecondary,
    onSecondaryContainer = DarkTextPrimary,
    tertiary             = WarnOrange,
    onTertiary           = AccentDark,
    tertiaryContainer    = Color(0xFF3D2E00),
    onTertiaryContainer  = Color(0xFFFFE0B2),
    background           = DarkBgMain,
    onBackground         = DarkTextPrimary,
    surface              = DarkSurface,
    onSurface            = DarkTextPrimary,
    surfaceVariant       = DarkBgSecondary,
    onSurfaceVariant     = DarkTextSecondary,
    outline              = Color(0x30FFFFFF),
    outlineVariant       = DarkBgTertiary,
    error                = Color(0xFFEF5350),
    onError              = Color.White,
)

// ==================== 间距系统 ====================
val SpacingXs = 4.dp
val SpacingSm = 8.dp
val SpacingMd = 16.dp
val SpacingLg = 24.dp

// ==================== 主题入口 ====================

@Composable
fun AlfQTheme(
    darkTheme: Boolean = false,
    content: @Composable () -> Unit
) {
    MaterialTheme(
        colorScheme = if (darkTheme) AlfQDarkColors else AlfQLightColors,
        typography = Typography(),
        content = content,
    )
}
