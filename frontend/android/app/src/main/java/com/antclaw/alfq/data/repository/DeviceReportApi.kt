package com.antclaw.alfq.data.repository

/** 设备上报能力接口 — 解耦 SessionViewModel 与具体实现，便于测试注入。 */
interface DeviceReportApi {
    suspend fun reportDeviceInfo()
}
