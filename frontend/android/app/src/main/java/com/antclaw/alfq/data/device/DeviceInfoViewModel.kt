package com.antclaw.alfq.data.device

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.antclaw.alfq.R
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class DeviceInfoViewModel @Inject constructor(
    private val deviceInfoCollector: DeviceInfoCollector,
    @ApplicationContext private val appContext: Context,
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
            _error.value = appContext.getString(R.string.device_consent_required)
            return
        }

        _isLoading.value = true
        _error.value = null

        viewModelScope.launch {
            try {
                val info = deviceInfoCollector.collect()
                _deviceInfo.value = info
                if (info == null) _error.value = appContext.getString(R.string.device_error_collect)
            } catch (e: Exception) {
                _error.value = appContext.getString(R.string.device_error_generic, e.message ?: "")
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