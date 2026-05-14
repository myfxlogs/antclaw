/**
 * React Hook for Device Information Collection
 * 
 * This hook provides a convenient way to use the device information collector
 * in React components.
 */

import { useState, useEffect, useCallback } from 'react';
import { deviceInfoCollector, type DeviceInfo } from './deviceInfo';

export interface UseDeviceInfoResult {
  deviceInfo: DeviceInfo | null;
  isLoading: boolean;
  error: string | null;
  consentStatus: 'pending' | 'granted' | 'denied';
  refresh: () => void;
  reset: () => void;
}

/**
 * React hook for collecting and managing device information
 */
export function useDeviceInfo(): UseDeviceInfoResult {
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [consentStatus, setConsentStatus] = useState<'pending' | 'granted' | 'denied'>('pending');

  const collectInfo = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      // Check consent status
      if (!deviceInfoCollector.hasConsent()) {
        setConsentStatus('pending');
        await deviceInfoCollector.requestConsent();
      }

      if (deviceInfoCollector.hasConsent()) {
        setConsentStatus('granted');
        const info = await deviceInfoCollector.collect();
        setDeviceInfo(info);
      } else {
        setConsentStatus('denied');
        setError('User consent not granted');
      }
    } catch (err) {
      setError(`Error collecting device info: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    collectInfo();
  }, [collectInfo]);

  const refresh = useCallback(() => {
    collectInfo();
  }, [collectInfo]);

  const reset = useCallback(() => {
    deviceInfoCollector.reset();
    setDeviceInfo(null);
    setConsentStatus('pending');
    setError(null);
  }, []);

  return {
    deviceInfo,
    isLoading,
    error,
    consentStatus,
    refresh,
    reset,
  };
}

/**
 * Hook for getting device summary only
 */
export function useDeviceSummary() {
  const { deviceInfo, isLoading, error } = useDeviceInfo();

  if (!deviceInfo) {
    return {
      isLoading,
      error,
      device: null,
      browser: null,
      os: null,
      screen: null,
    };
  }

  return {
    isLoading,
    error,
    device: `${deviceInfo.deviceType} (${deviceInfo.hardwareConcurrency} cores)`,
    browser: `${deviceInfo.browserName} ${deviceInfo.browserVersion}`,
    os: `${deviceInfo.osName} ${deviceInfo.osVersion}`,
    screen: `${deviceInfo.screenWidth}x${deviceInfo.screenHeight}`,
  };
}

/**
 * Hook for network information
 */
export function useNetworkInfo() {
  const { deviceInfo, isLoading, error } = useDeviceInfo();

  return {
    isLoading,
    error,
    networkType: deviceInfo?.networkType || 'unknown',
    connection: deviceInfo?.connection || null,
  };
}
