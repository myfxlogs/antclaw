package com.antclaw.alfq.data.local

import android.content.Context
import android.content.res.Configuration
import android.os.Build
import java.util.Locale

/**
 * 应用语言管理：设备语言检测 → 持久化 → 运行时切换。
 *
 * 支持的语言（ISO 639）：
 *   en — English（兜底）
 *   zh — 简体中文
 *   vi — Tiếng Việt
 *   th — ไทย
 *   id — Bahasa Indonesia
 *   ms — Bahasa Melayu
 *   ja — 日本語
 */
object LocaleManager {

    const val PREFS_KEY = "app_locale"
    const val KEY_SELECTED_LANGUAGE = "selected_language"
    const val KEY_FIRST_LAUNCH_DONE = "first_launch_done"

    val SUPPORTED_LOCALES = setOf("en", "zh", "vi", "th", "id", "ms", "ja")

    fun init(context: Context) {
        val prefs = context.getSharedPreferences(PREFS_KEY, Context.MODE_PRIVATE)
        val firstLaunchDone = prefs.getBoolean(KEY_FIRST_LAUNCH_DONE, false)
        if (!firstLaunchDone) {
            val deviceLang = detectDeviceLanguage(context)
            val bestMatch = bestMatchLocale(deviceLang)
            saveSelectedLanguage(context, bestMatch)
            prefs.edit().putBoolean(KEY_FIRST_LAUNCH_DONE, true).apply()
        }
        applyLocale(context, getSelectedLanguage(context))
    }

    fun getSelectedLanguage(context: Context): String {
        return context.getSharedPreferences(PREFS_KEY, Context.MODE_PRIVATE)
            .getString(KEY_SELECTED_LANGUAGE, "en") ?: "en"
    }

    fun setSelectedLanguage(context: Context, languageTag: String) {
        saveSelectedLanguage(context, bestMatchLocale(languageTag))
    }

    fun applyLocale(context: Context, languageTag: String): Context {
        val locale = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            Locale.forLanguageTag(languageTag)
        } else {
            Locale(languageTag)
        }
        Locale.setDefault(locale)
        val config = Configuration(context.resources.configuration)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.JELLY_BEAN_MR1) {
            config.setLocale(locale)
        } else {
            @Suppress("DEPRECATION")
            config.locale = locale
        }
        return context.createConfigurationContext(config)
    }

    fun getAvailableLanguages(): List<Pair<String, String>> = listOf(
        "en" to "English",
        "zh" to "简体中文",
        "vi" to "Tiếng Việt",
        "th" to "ไทย",
        "id" to "Bahasa Indonesia",
        "ms" to "Bahasa Melayu",
        "ja" to "日本語",
    )

    private fun saveSelectedLanguage(context: Context, languageTag: String) {
        context.getSharedPreferences(PREFS_KEY, Context.MODE_PRIVATE)
            .edit().putString(KEY_SELECTED_LANGUAGE, languageTag).apply()
    }

    private fun detectDeviceLanguage(context: Context): String {
        val config = context.resources.configuration
        val locale = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            config.locales[0]
        } else {
            @Suppress("DEPRECATION")
            config.locale
        }
        return locale?.language ?: "en"
    }

    private fun bestMatchLocale(languageTag: String): String {
        val lang = languageTag.substringBefore("-").lowercase(Locale.ENGLISH)
        return if (lang in SUPPORTED_LOCALES) lang else "en"
    }
}
