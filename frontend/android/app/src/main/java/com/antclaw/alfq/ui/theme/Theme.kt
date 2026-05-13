package com.antclaw.alfq.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

// ==================== anttrader 品牌色 ====================
// 来源：/opt/antclaw/emulator/anttrader/frontend 色彩基准
val Primary        = Color(0xFFD4AF37)   // 主品牌金
val PrimaryDark    = Color(0xFFB8960B)   // 深金
val PrimaryLight   = Color(0xFFE6C65C)   // 浅金
val AccentDark     = Color(0xFF141D22)   // 深石墨（主文本 / 强调色）

// 背景色
val BgMain         = Color(0xFFFFFFFF)   // 主背景
val BgCard         = Color(0xFFFFFFFF)   // 卡片背景
val BgSecondary    = Color(0xFFF5F7F9)   // 次级背景
val BgTertiary     = Color(0xFFE8ECF0)   // 三级背景

// 文本色
val TextPrimary    = Color(0xFF141D22)   // 正文
val TextSecondary  = Color(0xFF5A6B75)   // 辅助信息
val TextMuted      = Color(0xFF8A9AA5)   // 弱文本（时间、说明）

// 功能色
val SuccessGreen   = Color(0xFF00A651)   // 成功 / 看涨
val ErrorRed        = Color(0xFFE53935)   // 错误 / 强风险
val InfoBlue        = Color(0xFF2196F3)   // 信息提示
val WarnOrange      = Color(0xFFFFA000)   // 警告

// 信号方向色（看涨/看跌与 SuccessGreen/ErrorRed 统一）
val BullGreen = SuccessGreen
val BearRed   = ErrorRed

// ==================== 浅色主题 ====================

private val AlfQLightColors = lightColorScheme(
    primary            = Primary,
    onPrimary          = Color.White,
    primaryContainer   = PrimaryLight,
    onPrimaryContainer = AccentDark,

    secondary          = AccentDark,
    onSecondary        = Color.White,
    secondaryContainer = BgSecondary,
    onSecondaryContainer = TextPrimary,

    tertiary           = WarnOrange,
    onTertiary         = Color.White,
    tertiaryContainer  = Color(0xFFFFE0B2),
    onTertiaryContainer = Color(0xFFE65100),

    background         = BgMain,
    onBackground       = TextPrimary,

    surface            = Color.White,
    onSurface          = TextPrimary,
    surfaceVariant     = BgSecondary,
    onSurfaceVariant   = TextSecondary,

    outline            = Color(0x14000000),
    outlineVariant     = BgTertiary,

    error              = ErrorRed,
    onError            = Color.White,
)

// ==================== 暗色主题（交易终端质感） ====================

private val AlfQDarkColors = darkColorScheme(
    primary            = Primary,
    onPrimary          = AccentDark,
    primaryContainer   = PrimaryDark,
    onPrimaryContainer = Color.White,

    secondary          = PrimaryLight,
    onSecondary        = AccentDark,
    secondaryContainer = AccentDark,
    onSecondaryContainer = Color.White,

    tertiary           = WarnOrange,
    onTertiary         = Color.White,

    background         = Color(0xFF020617),   // 深蓝黑底
    onBackground       = Color(0xFFE8ECF0),

    surface            = Color(0xFF0F172A),
    onSurface          = Color(0xFFE8ECF0),
    surfaceVariant     = AccentDark,
    onSurfaceVariant   = TextMuted,

    outline            = Color(0xFF334155),
    outlineVariant     = Color(0xFF334155),

    error              = ErrorRed,
    onError            = Color.White,
)

// ==================== 间距系统 ====================
val SpacingXs = 4.dp
val SpacingSm = 8.dp
val SpacingMd = 16.dp
val SpacingLg = 24.dp
val SpacingXl = 32.dp
val SpacingXx = 48.dp

// ==================== 圆角系统 ====================
val CornerSm = 4.dp
val CornerMd = 8.dp
val CornerLg = 16.dp
val CornerXl = 24.dp
val CornerFull = 999.dp

// ==================== 阴影系统 ====================
val ElevationNone = 0.dp
val ElevationSm = 2.dp
val ElevationMd = 4.dp
val ElevationLg = 8.dp
val ElevationXl = 16.dp

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
