package com.antclaw.alfq.data.device

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

class DeviceInfoViewModel @Inject constructor(
    private val deviceInfoCollector: DeviceInfoCollector
) : ViewModel() {

    private val _deviceInfo = MutableStateFlow<DeviceInfo?>(null)
    val deviceInfo = _deviceInfo.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading = _isLoading.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error = _error.asStateFlow()

    private val _consentStatus = MutableStateFlow<ConsentStatus>(
        if (deviceInfoCollector.hasConsent()) ConsentStatus.GRANTED else ConsentStatus.PENDING
    )
    val consentStatus = _consentStatus.asStateFlow()

    fun requestConsent(given: Boolean) {
        deviceInfoCollector.requestConsent(given)
        _consentStatus.value = if (given) ConsentStatus.GRANTED else ConsentStatus.DENIED
        if (given) collectDeviceInfo()
    }

    fun collectDeviceInfo() {
        if (!deviceInfoCollector.hasConsent()) {
            _error.value = "需要用户同意才能收集设备信息"
            return
        }

        _isLoading.value = true
        _error.value = null

        viewModelScope.launch {
            try {
                val info = deviceInfoCollector.collect()
                _deviceInfo.value = info
                if (info == null) _error.value = "设备信息收集失败"
            } catch (e: Exception) {
                _error.value = "收集设备信息时发生错误: ${e.message}"
            } finally {
                _isLoading.value = false
            }
        }
    }

    fun getSummary(): DeviceInfoSummary? = _deviceInfo.value?.let { deviceInfoCollector.getSummary(it) }

    fun refresh() = collectDeviceInfo()

    fun reset() {
        deviceInfoCollector.reset()
        _deviceInfo.value = null
        _consentStatus.value = ConsentStatus.PENDING
        _error.value = null
    }

    enum class ConsentStatus { PENDING, GRANTED, DENIED }
}