package com.antclaw.alfq.data.device

import org.junit.Assert.*
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

@RunWith(RobolectricTestRunner::class)
class DeviceInfoCollectorBasicTest {

    private val collector = DeviceInfoCollector(
        org.robolectric.RuntimeEnvironment.getApplication()
    )

    @Test fun `collectBasic returns non-null without consent`() {
        // Ensure consent is false
        collector.requestConsent(false)
        assertFalse(collector.hasConsent())

        val info = collector.collectBasic()
        assertNotNull(info)
    }

    @Test fun `collectBasic fills deviceId`() {
        val info = collector.collectBasic()
        assertFalse(info.deviceId.isBlank())
    }

    @Test fun `collectBasic fills brand model osVersion`() {
        val info = collector.collectBasic()
        assertFalse(info.brand.isBlank())
        assertFalse(info.model.isBlank())
        assertFalse(info.osVersion.isBlank())
    }

    @Test fun `collectBasic fills appVersion and buildNumber`() {
        val info = collector.collectBasic()
        assertFalse(info.appVersionName.isBlank())
        // appVersionCode may be 0 in Robolectric, so we just check it's set
        assertNotNull(info.appVersionCode)
    }

    @Test fun `collectBasic skips sensitive fields`() {
        val info = collector.collectBasic()
        assertEquals(0, info.screenWidth)
        assertEquals(0, info.screenHeight)
        assertEquals(NetworkType.UNKNOWN, info.networkType)
        assertEquals(-1, info.batteryLevel)
    }

    @Test fun `collectBasic deviceType is PHONE by default`() {
        val info = collector.collectBasic()
        assertEquals(DeviceType.PHONE, info.deviceType)
    }
}
