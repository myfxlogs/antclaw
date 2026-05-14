/**
 * Device Information Collection Test Script
 * 
 * This script tests the device information collection functionality
 * and verifies all specified data points are collected correctly.
 */

import { deviceInfoCollector, type DeviceInfo } from './deviceInfo';

/**
 * Test function to verify device information collection
 */
export async function testDeviceInfoCollection(): Promise<{
  success: boolean;
  errors: string[];
  deviceInfo: DeviceInfo | null;
}> {
  const errors: string[] = [];
  
  console.log('=== Device Information Collection Test ===\n');
  
  try {
    // Step 1: Request consent
    console.log('Step 1: Requesting user consent...');
    const consent = await deviceInfoCollector.requestConsent();
    console.log(`Consent granted: ${consent}`);
    
    if (!consent) {
      errors.push('User consent not granted');
      return { success: false, errors, deviceInfo: null };
    }
    
    // Step 2: Collect device information
    console.log('\nStep 2: Collecting device information...');
    const deviceInfo = await deviceInfoCollector.collect();
    
    console.log('\n=== Collected Device Information ===');
    console.log(JSON.stringify(deviceInfo, null, 2));
    
    // Step 3: Validate collected data
    console.log('\n=== Validation Results ===');
    
    // Validate required fields
    const validationChecks = [
      { field: 'deviceId', value: deviceInfo.deviceId, validator: (v: string) => v.length > 0 },
      { field: 'deviceType', value: deviceInfo.deviceType, validator: (v: string) => ['desktop', 'mobile', 'tablet', 'unknown'].includes(v) },
      { field: 'browserName', value: deviceInfo.browserName, validator: (v: string) => v.length > 0 },
      { field: 'browserVersion', value: deviceInfo.browserVersion, validator: (v: string) => v.length > 0 },
      { field: 'osName', value: deviceInfo.osName, validator: (v: string) => v.length > 0 },
      { field: 'osVersion', value: deviceInfo.osVersion, validator: (v: string) => v.length > 0 },
      { field: 'screenWidth', value: deviceInfo.screenWidth, validator: (v: number) => v > 0 },
      { field: 'screenHeight', value: deviceInfo.screenHeight, validator: (v: number) => v > 0 },
      { field: 'viewportWidth', value: deviceInfo.viewportWidth, validator: (v: number) => v > 0 },
      { field: 'viewportHeight', value: deviceInfo.viewportHeight, validator: (v: number) => v > 0 },
      { field: 'pixelRatio', value: deviceInfo.pixelRatio, validator: (v: number) => v >= 1 },
      { field: 'touchSupport', value: deviceInfo.touchSupport, validator: (v: boolean) => typeof v === 'boolean' },
      { field: 'hardwareConcurrency', value: deviceInfo.hardwareConcurrency, validator: (v: number) => v >= 1 },
      { field: 'timezone', value: deviceInfo.timezone, validator: (v: string) => v.length > 0 },
      { field: 'locale', value: deviceInfo.locale, validator: (v: string) => v.length > 0 },
      { field: 'language', value: deviceInfo.language, validator: (v: string) => v.length > 0 },
      { field: 'collectedAt', value: deviceInfo.collectedAt, validator: (v: number) => v > 0 },
      { field: 'sessionId', value: deviceInfo.sessionId, validator: (v: string) => v.length > 0 },
    ];
    
    let passedChecks = 0;
    let failedChecks = 0;
    
    for (const { field, value, validator } of validationChecks) {
      try {
        const isValid = (validator as (v: any) => boolean)(value);
        if (isValid) {
          console.log(`✓ ${field}: ${typeof value === 'string' ? `"${value}"` : value}`);
          passedChecks++;
        } else {
          console.log(`✗ ${field}: Invalid value - ${typeof value === 'string' ? `"${value}"` : value}`);
          errors.push(`Invalid ${field}: ${value}`);
          failedChecks++;
        }
      } catch (err) {
        console.log(`✗ ${field}: Error validating - ${err}`);
        errors.push(`Error validating ${field}: ${err}`);
        failedChecks++;
      }
    }
    
    // Additional validation for network connection
    if (deviceInfo.connection) {
      console.log(`✓ connection.effectiveType: ${deviceInfo.connection.effectiveType}`);
      console.log(`✓ connection.downlink: ${deviceInfo.connection.downlink} Mbps`);
      console.log(`✓ connection.rtt: ${deviceInfo.connection.rtt} ms`);
      console.log(`✓ connection.saveData: ${deviceInfo.connection.saveData}`);
      passedChecks += 4;
    } else {
      console.log(`~ connection: Not available in this browser`);
    }
    
    // Check memory (optional)
    if (deviceInfo.memory !== null) {
      console.log(`✓ memory: ${deviceInfo.memory} GB`);
      passedChecks++;
    } else {
      console.log(`~ memory: Not available in this browser`);
    }
    
    // Summary
    console.log(`\n=== Test Summary ===`);
    console.log(`Total checks: ${validationChecks.length + 5}`);
    console.log(`Passed: ${passedChecks}`);
    console.log(`Failed: ${failedChecks}`);
    
    if (failedChecks === 0) {
      console.log('\n✓ All tests passed! Device information collection is working correctly.');
      return { success: true, errors: [], deviceInfo };
    } else {
      console.log('\n✗ Some tests failed. Please review the errors above.');
      return { success: false, errors, deviceInfo };
    }
    
  } catch (err) {
    const errorMessage = `Unexpected error during test: ${err instanceof Error ? err.message : 'Unknown error'}`;
    console.error(errorMessage);
    errors.push(errorMessage);
    return { success: false, errors, deviceInfo: null };
  }
}

/**
 * Run tests when this file is imported
 */
if (typeof window !== 'undefined') {
  // Only run tests in browser environment
  document.addEventListener('DOMContentLoaded', async () => {
    try {
      await testDeviceInfoCollection();
    } catch (err) {
      console.error('Error running device info tests:', err);
    }
  });
}
