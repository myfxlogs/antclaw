package com.antclaw.alfq.data.device

/**
 * 设备信息数据类 - 精简版
 */
data class DeviceInfo(
    // 设备标识
    val deviceId: String,
    val deviceType: DeviceType,
    
    // 硬件信息
    val manufacturer: String,
    val model: String,
    val brand: String,
    
    // 操作系统
    val osVersion: String,
    val apiLevel: Int,
    val securityPatch: String,
    
    // 屏幕信息
    val screenWidth: Int,
    val screenHeight: Int,
    val densityDpi: Int,
    
    // 网络信息
    val networkType: NetworkType,
    
    // 电池信息
    val batteryLevel: Int,
    val isCharging: Boolean,
    
    // 应用信息
    val appVersionName: String,
    val appVersionCode: Int,
    val packageName: String,
    
    // 系统信息
    val timezone: String,
    val locale: String,
    val isEmulator: Boolean,
    
    // 收集元数据
    val collectedAt: Long,
    val sessionId: String
)

/**
 * 设备类型
 */
enum class DeviceType {
    PHONE, TABLET, UNKNOWN
}

/**
 * 网络类型
 */
enum class NetworkType {
    WIFI, CELLULAR_2G, CELLULAR_3G, CELLULAR_4G, UNKNOWN
}

/**
 * 设备信息摘要
 */
data class DeviceInfoSummary(
    val deviceName: String,
    val osInfo: String,
    val screenInfo: String,
    val networkInfo: String,
    val batteryInfo: String
)