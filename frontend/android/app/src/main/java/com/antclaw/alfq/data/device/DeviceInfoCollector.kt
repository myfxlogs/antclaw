package com.antclaw.alfq.data.device

import android.annotation.SuppressLint
import android.content.Context
import android.content.pm.PackageManager
import android.os.BatteryManager
import android.os.Build
import android.provider.Settings
import android.util.DisplayMetrics
import android.view.WindowManager
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.util.*
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 设备信息收集器 - 精简版
 * 
 * 负责收集设备信息和生成设备标识符
 */
@Singleton
class DeviceInfoCollector @Inject constructor(
    @ApplicationContext private val context: Context
) {
    companion object {
        private const val PREFS_NAME = "alfq_device_prefs"
        private const val KEY_DEVICE_ID = "device_id"
        private const val KEY_CONSENT = "consent_given"
        private const val CACHE_TIMEOUT = 300000L // 5分钟缓存
    }

    private var consentGiven: Boolean = false
    private var cachedDeviceInfo: DeviceInfo? = null

    init {
        consentGiven = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            .getBoolean(KEY_CONSENT, false)
    }

    // 同意管理
    fun requestConsent(given: Boolean) {
        consentGiven = given
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
            .edit().putBoolean(KEY_CONSENT, given).apply()
    }

    fun hasConsent(): Boolean = consentGiven

    // 设备ID管理
    @SuppressLint("HardwareIds")
    fun getDeviceId(): String {
        val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        var deviceId = prefs.getString(KEY_DEVICE_ID, null)

        if (deviceId.isNullOrEmpty()) {
            deviceId = runCatching {
                Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
            }.getOrNull().takeIf { !it.isNullOrEmpty() } 
                ?: "alfq_${UUID.randomUUID().toString().replace("-", "")}"
            
            prefs.edit().putString(KEY_DEVICE_ID, deviceId).apply()
        }

        return deviceId
    }

    fun getSessionId(): String = "session_${System.currentTimeMillis().toString(36)}${UUID.randomUUID().toString().substring(0, 8)}"

    fun getUserAgent(): String {
        val appVersion = runCatching {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.packageManager.getPackageInfo(context.packageName, PackageManager.PackageInfoFlags.of(0)).versionName
            } else {
                @Suppress("DEPRECATION")
                context.packageManager.getPackageInfo(context.packageName, 0).versionName
            }
        }.getOrNull() ?: "1.0.0"
        
        return "AntClaw/$appVersion (Android ${Build.VERSION.SDK_INT}; ${Build.MODEL})"
    }

    // ── 基础设备信息（不需要 consent）──
    // 用于登录后立即补全管理端设备列表基础字段
    fun collectBasic(): DeviceInfo {
        return DeviceInfo(
            deviceId = getDeviceId(),
            deviceType = DeviceType.PHONE,
            manufacturer = Build.MANUFACTURER,
            model = Build.MODEL,
            brand = Build.BRAND,
            osVersion = Build.VERSION.RELEASE,
            apiLevel = Build.VERSION.SDK_INT,
            securityPatch = Build.VERSION.SECURITY_PATCH,
            screenWidth = 0,
            screenHeight = 0,
            densityDpi = 0,
            networkType = NetworkType.UNKNOWN,
            batteryLevel = -1,
            isCharging = false,
            appVersionName = getAppVersionName(),
            appVersionCode = getAppVersionCode(),
            packageName = context.packageName,
            timezone = TimeZone.getDefault().id,
            locale = Locale.getDefault().language,
            isEmulator = isRunningOnEmulator(),
            collectedAt = System.currentTimeMillis(),
            sessionId = getSessionId(),
        )
    }

    // 设备信息收集（完整版，需 consent）
    suspend fun collect(): DeviceInfo? {
        if (!consentGiven) return null
        
        cachedDeviceInfo?.let { 
            if (System.currentTimeMillis() - it.collectedAt < CACHE_TIMEOUT) return it 
        }

        return withContext(Dispatchers.IO) {
            runCatching {
                val displayMetrics = getDisplayMetrics()
                val info = collectBasic().copy(
                    deviceType = getDeviceType(displayMetrics),
                    screenWidth = displayMetrics.widthPixels,
                    screenHeight = displayMetrics.heightPixels,
                    densityDpi = displayMetrics.densityDpi,
                    networkType = getNetworkType(),
                    batteryLevel = getBatteryLevel(),
                    isCharging = isCharging(),
                    collectedAt = System.currentTimeMillis(),
                    sessionId = getSessionId(),
                )
                cachedDeviceInfo = info
                info
            }.getOrElse {
                android.util.Log.e("DeviceInfoCollector", "Error collecting device info", it)
                null
            }
        }
    }

    // 获取显示指标
    private fun getDisplayMetrics(): DisplayMetrics {
        return context.resources.displayMetrics
    }

    // 获取设备类型
    private fun getDeviceType(displayMetrics: DisplayMetrics): DeviceType {
        val smallestWidth = displayMetrics.widthPixels.coerceAtMost(displayMetrics.heightPixels)
        val smallestWidthDp = smallestWidth / displayMetrics.density
        return when {
            smallestWidthDp >= 600 -> DeviceType.TABLET
            else -> DeviceType.PHONE
        }
    }

    // 获取网络类型
    private fun getNetworkType(): NetworkType {
        return try {
            val connectivityManager = 
                context.getSystemService(Context.CONNECTIVITY_SERVICE) as android.net.ConnectivityManager
            
            val capabilities = connectivityManager.getNetworkCapabilities(connectivityManager.activeNetwork)
            when {
                capabilities?.hasTransport(android.net.NetworkCapabilities.TRANSPORT_WIFI) == true -> NetworkType.WIFI
                capabilities?.hasTransport(android.net.NetworkCapabilities.TRANSPORT_CELLULAR) == true -> {
                    @Suppress("DEPRECATION")
                    @SuppressLint("MissingPermission")
                    when ((context.getSystemService(Context.TELEPHONY_SERVICE) as android.telephony.TelephonyManager).networkType) {
                        android.telephony.TelephonyManager.NETWORK_TYPE_LTE -> NetworkType.CELLULAR_4G
                        android.telephony.TelephonyManager.NETWORK_TYPE_HSDPA,
                        android.telephony.TelephonyManager.NETWORK_TYPE_HSPA,
                        android.telephony.TelephonyManager.NETWORK_TYPE_HSPAP,
                        android.telephony.TelephonyManager.NETWORK_TYPE_UMTS -> NetworkType.CELLULAR_3G
                        else -> NetworkType.CELLULAR_2G
                    }
                }
                else -> NetworkType.UNKNOWN
            }
        } catch (_: Exception) {
            NetworkType.UNKNOWN
        }
    }

    // 获取电池电量
    private fun getBatteryLevel(): Int {
        return runCatching {
            (context.getSystemService(Context.BATTERY_SERVICE) as BatteryManager)
                .getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY)
        }.getOrDefault(-1)
    }

    // 是否充电中
    private fun isCharging(): Boolean {
        return runCatching {
            val batteryManager = context.getSystemService(Context.BATTERY_SERVICE) as BatteryManager
            val status = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_STATUS)
            status == BatteryManager.BATTERY_STATUS_CHARGING || status == BatteryManager.BATTERY_STATUS_FULL
        }.getOrDefault(false)
    }

    // 获取应用版本名称
    private fun getAppVersionName(): String {
        return runCatching {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.packageManager.getPackageInfo(context.packageName, PackageManager.PackageInfoFlags.of(0)).versionName
            } else {
                @Suppress("DEPRECATION")
                context.packageManager.getPackageInfo(context.packageName, 0).versionName
            }
        }.getOrNull() ?: "unknown"
    }

    // 获取应用版本号
    private fun getAppVersionCode(): Int {
        return runCatching {
            val info = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                context.packageManager.getPackageInfo(context.packageName, PackageManager.PackageInfoFlags.of(0))
            } else {
                @Suppress("DEPRECATION")
                context.packageManager.getPackageInfo(context.packageName, 0)
            }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) info.longVersionCode.toInt()
            else @Suppress("DEPRECATION") info.versionCode
        }.getOrDefault(0)
    }

    // 是否模拟器
    private fun isRunningOnEmulator(): Boolean {
        return Build.FINGERPRINT.startsWith("generic") ||
               Build.FINGERPRINT.startsWith("unknown") ||
               Build.MODEL.contains("Emulator") ||
               Build.MODEL.contains("Android SDK built for x86") ||
               "google_sdk" == Build.PRODUCT
    }

    // 获取摘要
    fun getSummary(deviceInfo: DeviceInfo): DeviceInfoSummary {
        return DeviceInfoSummary(
            deviceName = "${deviceInfo.manufacturer} ${deviceInfo.model}",
            osInfo = "Android ${deviceInfo.osVersion} (API ${deviceInfo.apiLevel})",
            screenInfo = "${deviceInfo.screenWidth}x${deviceInfo.screenHeight} @ ${deviceInfo.densityDpi}dpi",
            networkInfo = deviceInfo.networkType.name,
            batteryInfo = "${deviceInfo.batteryLevel}% ${if (deviceInfo.isCharging) "(充电中)" else ""}"
        )
    }

    // 重置数据
    fun reset() {
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE).edit().clear().apply()
        consentGiven = false
        cachedDeviceInfo = null
    }
}