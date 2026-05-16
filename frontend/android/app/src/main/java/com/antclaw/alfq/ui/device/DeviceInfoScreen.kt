package com.antclaw.alfq.ui.device

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.antclaw.alfq.R
import com.antclaw.alfq.data.device.DeviceInfoViewModel
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.rememberTopAppBarState

/**
 * Device Information Screen
 * 
 * Displays comprehensive device information collected by the DeviceInfoCollector.
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DeviceInfoScreen(
    viewModel: DeviceInfoViewModel = hiltViewModel()
) {
    val deviceInfo by viewModel.deviceInfo.collectAsState()
    val isLoading by viewModel.isLoading.collectAsState()
    val error by viewModel.error.collectAsState()
    val consentStatus by viewModel.consentStatus.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(stringResource(R.string.device_title)) },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = stringResource(R.string.device_refresh))
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background)
            )
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(16.dp),
            verticalArrangement = Arrangement.Top,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Consent Section
            when (consentStatus) {
                DeviceInfoViewModel.ConsentStatus.PENDING -> {
                    ConsentSection(onConsentGranted = { viewModel.requestConsent(true) })
                }
                DeviceInfoViewModel.ConsentStatus.DENIED -> {
                    Text(stringResource(R.string.device_consent_denied), color = MaterialTheme.colorScheme.error)
                }
                DeviceInfoViewModel.ConsentStatus.GRANTED -> {
                    // Show device info
                    if (isLoading) {
                        CircularProgressIndicator()
                    } else if (error != null) {
                        ErrorSection(error = error!!, onRetry = { viewModel.collectDeviceInfo() })
                    } else if (deviceInfo != null) {
                        DeviceInfoContent(deviceInfo = deviceInfo!!)
                    } else {
                        EmptyState(onCollect = { viewModel.collectDeviceInfo() })
                    }
                }
            }
        }
    }
}

/**
 * Consent Section Component
 */
@Composable
fun ConsentSection(onConsentGranted: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                Icons.Default.Info,
                contentDescription = "Info",
                modifier = Modifier.size(48.dp),
                tint = MaterialTheme.colorScheme.primary
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = stringResource(R.string.device_consent_title),
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(R.string.device_consent_body),
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = onConsentGranted,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(stringResource(R.string.device_consent_grant))
            }
        }
    }
}

/**
 * Error Section Component
 */
@Composable
fun ErrorSection(error: String, onRetry: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Icon(
                Icons.Default.Warning,
                contentDescription = stringResource(R.string.device_error),
                modifier = Modifier.size(48.dp),
                tint = MaterialTheme.colorScheme.error
            )
            Spacer(modifier = Modifier.height(16.dp))
            Text(
                text = stringResource(R.string.device_error),
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.error
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = error,
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(16.dp))
            Button(
                onClick = onRetry,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(stringResource(R.string.device_retry))
            }
        }
    }
}

/**
 * Empty State Component
 */
@Composable
fun EmptyState(onCollect: () -> Unit) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
        modifier = Modifier.fillMaxSize()
    ) {
        Icon(
            Icons.Default.Info,
            contentDescription = "Devices",
            modifier = Modifier.size(64.dp),
            tint = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f)
        )
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = stringResource(R.string.device_empty),
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
        )
        Spacer(modifier = Modifier.height(16.dp))
        Button(onClick = onCollect) {
            Text(stringResource(R.string.device_collect))
        }
    }
}

/**
 * Device Info Content Component
 */
@Composable
fun DeviceInfoContent(deviceInfo: com.antclaw.alfq.data.device.DeviceInfo) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Device Card
        DeviceCard(
            title = "Device",
            icon = Icons.Default.Phone,
            items = listOf(
                "Manufacturer" to deviceInfo.manufacturer,
                "Model" to deviceInfo.model,
                "Brand" to deviceInfo.brand,
                "Type" to deviceInfo.deviceType.name
            )
        )

        // OS Card
        DeviceCard(
            title = "Operating System",
            icon = Icons.Default.Build,
            items = listOf(
                "Version" to deviceInfo.osVersion,
                "API Level" to deviceInfo.apiLevel.toString(),
                "Security Patch" to deviceInfo.securityPatch
            )
        )

        // Screen Card
        DeviceCard(
            title = "Screen",
            icon = Icons.Default.Star,
            items = listOf(
                "Resolution" to "${deviceInfo.screenWidth} x ${deviceInfo.screenHeight}",
                "Density" to "${deviceInfo.densityDpi} dpi"
            )
        )

        // Network Card
        DeviceCard(
            title = "Network",
            icon = Icons.Default.Settings,
            items = listOf(
                "Type" to deviceInfo.networkType.name
            )
        )

        // Battery Card
        DeviceCard(
            title = "Battery",
            icon = Icons.Default.Star,
            items = listOf(
                "Level" to "${deviceInfo.batteryLevel}%",
                "Charging" to if (deviceInfo.isCharging) "Yes" else "No"
            )
        )

        // App Card
        DeviceCard(
            title = "Application",
            icon = Icons.Default.Star,
            items = listOf(
                "Version" to deviceInfo.appVersionName,
                "Version Code" to deviceInfo.appVersionCode.toString(),
                "Package" to deviceInfo.packageName
            )
        )

        // System Card
        DeviceCard(
            title = "System",
            icon = Icons.Default.Settings,
            items = listOf(
                "Timezone" to deviceInfo.timezone,
                "Locale" to deviceInfo.locale,
                "Emulator" to if (deviceInfo.isEmulator) "Yes" else "No"
            )
        )
    }
}

/**
 * Device Card Component
 */
@Composable
fun DeviceCard(
    title: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    items: List<Pair<String, String>>
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    icon,
                    contentDescription = title,
                    modifier = Modifier.size(24.dp),
                    tint = MaterialTheme.colorScheme.primary
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
            }
            Spacer(modifier = Modifier.height(12.dp))
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                items.forEach { (label, value) ->
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            text = label,
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                        Text(
                            text = value,
                            style = MaterialTheme.typography.bodyMedium,
                            fontWeight = FontWeight.Medium
                        )
                    }
                }
            }
        }
    }
}

/**
 * Format memory bytes to human readable string
 */
fun formatMemory(bytes: Long): String {
    return when {
        bytes >= 1024 * 1024 * 1024 -> "%.2f GB".format(bytes.toDouble() / (1024 * 1024 * 1024))
        bytes >= 1024 * 1024 -> "%.2f MB".format(bytes.toDouble() / (1024 * 1024))
        bytes >= 1024 -> "%.2f KB".format(bytes.toDouble() / 1024)
        bytes > 0 -> "$bytes B"
        else -> "Unknown"
    }
}
