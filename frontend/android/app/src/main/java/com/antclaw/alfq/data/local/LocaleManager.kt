package com.antclaw.alfq.data.local

import android.annotation.SuppressLint
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
 *   zh-TW — 繁體中文
 */
object LocaleManager {

    const val PREFS_KEY = "app_locale"
    const val KEY_SELECTED_LANGUAGE = "selected_language"
    const val KEY_FIRST_LAUNCH_DONE = "first_launch_done"

    val SUPPORTED_LOCALES = setOf("en", "zh", "zh-TW", "vi", "th", "id", "ms", "ja")

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

    @SuppressLint("AppBundleLocaleChanges")
    fun applyLocale(context: Context, languageTag: String): Context {
        val locale = buildLocale(languageTag)
        Locale.setDefault(locale)
        
        val config = Configuration(context.resources.configuration)
        
        config.setLocales(android.os.LocaleList(locale))
        return context.createConfigurationContext(config)
    }

    private fun buildLocale(languageTag: String): Locale {
        // zh-TW 用 Locale("zh","TW") 确保匹配 values-zh-rTW
        if (languageTag == "zh-TW") return Locale("zh", "TW")
        return Locale.forLanguageTag(languageTag)
    }

    fun getAvailableLanguages(): List<Pair<String, String>> = listOf(
        "en" to "English",
        "zh" to "简体中文",
        "vi" to "Tiếng Việt",
        "th" to "ไทย",
        "id" to "Bahasa Indonesia",
        "ms" to "Bahasa Melayu",
        "ja" to "日本語",
        "zh-TW" to "繁體中文",
    )

    private fun saveSelectedLanguage(context: Context, languageTag: String) {
        context.getSharedPreferences(PREFS_KEY, Context.MODE_PRIVATE)
            .edit().putString(KEY_SELECTED_LANGUAGE, languageTag).apply()
    }

    private fun detectDeviceLanguage(context: Context): String {
        val config = context.resources.configuration
        val locale = config.locales[0]
        return if (locale != null) {
            val lang = locale.language.lowercase(Locale.ENGLISH)
            val country = locale.country.uppercase(Locale.ENGLISH)
            // 繁体中文设备用 zh-TW 标签区分
            if (lang == "zh" && country == "TW") "zh-TW"
            else if (lang == "zh" && (country == "HK" || country == "MO")) "zh-TW"
            else lang
        } else "en"
    }

    private fun bestMatchLocale(languageTag: String): String {
        val normalized = languageTag.replace('_', '-')
        val locale = Locale.forLanguageTag(normalized)
        val lang = locale.language.lowercase(Locale.ENGLISH)
        val country = locale.country.uppercase(Locale.ENGLISH)
        if (lang == "zh" && (country == "TW" || country == "HK" || country == "MO")) return "zh-TW"
        if (lang == "zh") return "zh"
        return if (lang in SUPPORTED_LOCALES) lang else "en"
    }
}