package com.antclaw.alfq.data.repository

import android.util.Log
import antclaw.v1.Device
import com.antclaw.alfq.data.device.DeviceInfoCollector
import com.antclaw.alfq.data.rpc.ConnectTransportProvider
import com.connectrpc.MethodSpec
import com.connectrpc.StreamType
import com.connectrpc.getOrThrow
import javax.inject.Inject
import javax.inject.Singleton

/**
 * 设备信息上报仓库
 *
 * 登录成功后异步调用 ReportDeviceInfo RPC 补全设备信息。
 * 优先使用完整采集（需 consent），无 consent 时回落基础上报，
 * 确保管理端至少看到 device_id / brand / model / os_version 等基础字段。
 */
@Singleton
class DeviceRepository @Inject constructor(
    private val deviceInfoCollector: DeviceInfoCollector
) : DeviceReportApi {
    companion object {
        private const val TAG = "DeviceRepository"
    }

    /**
     * 登录 / 注册成功后调用，异步上报设备信息。
     * 失败静默——不影响主流程。
     */
    override suspend fun reportDeviceInfo() {
        val di = deviceInfoCollector.collect()
            ?: deviceInfoCollector.collectBasic()

        try {
            val client = ConnectTransportProvider.createProtocolClient()

            val deviceInfo = Device.DeviceInfo.newBuilder()
                .setDeviceId(di.deviceId)
                .setModel(di.model)
                .setBrand(di.brand)
                .setManufacturer(di.manufacturer)
                .setOsVersion(di.osVersion)
                .setOsType("android")
                .setAppVersion(di.appVersionName)
                .setBuildNumber(di.appVersionCode.toString())
                .setScreenWidth(di.screenWidth)
                .setScreenHeight(di.screenHeight)
                .setNetworkType(di.networkType.name.lowercase())
                .setTimezone(di.timezone)
                .setLocale(di.locale)
                .setFingerprint("${di.manufacturer}/${di.model}/${di.osVersion}")
                .build()

            val request = Device.ReportDeviceInfoRequest.newBuilder()
                .setDeviceInfo(deviceInfo)
                .build()

            val spec = MethodSpec("antclaw.v1.DeviceService/ReportDeviceInfo",
                Device.ReportDeviceInfoRequest::class,
                Device.ReportDeviceInfoResponse::class,
                StreamType.UNARY)
            val resp = client.unary(request, emptyMap(), spec)
            val success = resp.getOrThrow().success
            Log.i(TAG, "reportDeviceInfo success: $success")
        } catch (e: Exception) {
            Log.e(TAG, "reportDeviceInfo failed", e)
        }
    }
}
