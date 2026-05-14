/**
 * Device Information Test Component
 * 
 * This component demonstrates the device information collection functionality
 * and provides a UI to view the collected data.
 */

import { useState, useEffect } from 'react';
import { deviceInfoCollector, type DeviceInfo, formatDeviceInfo, getDeviceSummary } from './deviceInfo';

interface DeviceInfoTestProps {
  onDataCollected?: (info: DeviceInfo) => void;
}

export function DeviceInfoTest({ onDataCollected }: DeviceInfoTestProps) {
  const [isLoading, setIsLoading] = useState(true);
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [consentStatus, setConsentStatus] = useState<'pending' | 'granted' | 'denied'>('pending');

  useEffect(() => {
    const collectDeviceInfo = async () => {
      try {
        setIsLoading(true);
        
        // Check consent status
        if (!deviceInfoCollector.hasConsent()) {
          setConsentStatus('pending');
          await deviceInfoCollector.requestConsent();
        }
        
        if (deviceInfoCollector.hasConsent()) {
          setConsentStatus('granted');
          const info = await deviceInfoCollector.collect();
          setDeviceInfo(info);
          onDataCollected?.(info);
        } else {
          setConsentStatus('denied');
          setError('User consent not granted');
        }
      } catch (err) {
        setError(`Error collecting device info: ${err instanceof Error ? err.message : 'Unknown error'}`);
      } finally {
        setIsLoading(false);
      }
    };

    collectDeviceInfo();
  }, [onDataCollected]);

  if (isLoading) {
    return (
      <div className="device-info-test">
        <div className="loading">
          <div className="spinner"></div>
          <p>Collecting device information...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="device-info-test error">
        <p className="error-message">Error: {error}</p>
      </div>
    );
  }

  if (!deviceInfo) {
    return (
      <div className="device-info-test">
        <p>No device information available</p>
      </div>
    );
  }

  const summary = getDeviceSummary(deviceInfo);

  return (
    <div className="device-info-test">
      <h2>Device Information</h2>
      
      <div className="consent-badge">
        Consent: <span className={`status ${consentStatus}`}>{consentStatus}</span>
      </div>

      <div className="summary-cards">
        <div className="card">
          <div className="card-icon">📱</div>
          <div className="card-content">
            <div className="card-label">Device</div>
            <div className="card-value">{summary.device}</div>
          </div>
        </div>
        
        <div className="card">
          <div className="card-icon">🌐</div>
          <div className="card-content">
            <div className="card-label">Browser</div>
            <div className="card-value">{summary.browser}</div>
          </div>
        </div>
        
        <div className="card">
          <div className="card-icon">🖥️</div>
          <div className="card-content">
            <div className="card-label">OS</div>
            <div className="card-value">{summary.os}</div>
          </div>
        </div>
        
        <div className="card">
          <div className="card-icon">📐</div>
          <div className="card-content">
            <div className="card-label">Screen</div>
            <div className="card-value">{summary.screen}</div>
          </div>
        </div>
      </div>

      <div className="details-section">
        <h3>Detailed Information</h3>
        <pre className="json-output">{formatDeviceInfo(deviceInfo)}</pre>
      </div>

      <style>{`
        .device-info-test {
          padding: 20px;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
          max-width: 800px;
          margin: 0 auto;
        }

        .device-info-test h2 {
          color: #333;
          border-bottom: 2px solid #D4AF37;
          padding-bottom: 10px;
          margin-bottom: 20px;
        }

        .device-info-test h3 {
          color: #555;
          margin-top: 20px;
          margin-bottom: 10px;
        }

        .consent-badge {
          display: inline-block;
          padding: 6px 12px;
          background: #f5f5f5;
          border-radius: 20px;
          margin-bottom: 20px;
          font-size: 14px;
        }

        .consent-badge .status {
          font-weight: bold;
          margin-left: 5px;
        }

        .consent-badge .status.pending {
          color: #f59e0b;
        }

        .consent-badge .status.granted {
          color: #10b981;
        }

        .consent-badge .status.denied {
          color: #ef4444;
        }

        .summary-cards {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 15px;
          margin-bottom: 25px;
        }

        .card {
          background: linear-gradient(135deg, #fff 0%, #f9f9f9 100%);
          border-radius: 12px;
          padding: 15px;
          box-shadow: 0 2px 8px rgba(0,0,0,0.1);
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .card-icon {
          font-size: 28px;
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: #f0f0f0;
          border-radius: 10px;
        }

        .card-content {
          flex: 1;
        }

        .card-label {
          font-size: 12px;
          color: #888;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .card-value {
          font-size: 14px;
          color: #333;
          font-weight: 600;
          margin-top: 2px;
        }

        .details-section {
          background: #f8f9fa;
          border-radius: 12px;
          padding: 20px;
        }

        .json-output {
          background: #1e1e1e;
          color: #d4d4d4;
          padding: 15px;
          border-radius: 8px;
          overflow-x: auto;
          font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
          font-size: 12px;
          line-height: 1.5;
          white-space: pre-wrap;
          word-break: break-all;
        }

        .loading {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 40px;
        }

        .spinner {
          width: 40px;
          height: 40px;
          border: 4px solid #f3f3f3;
          border-top: 4px solid #D4AF37;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }

        .loading p {
          margin-top: 15px;
          color: #666;
        }

        .error-message {
          color: #dc2626;
          background: #fee2e2;
          padding: 12px 16px;
          border-radius: 8px;
          border-left: 4px solid #dc2626;
        }
      `}</style>
    </div>
  );
}
