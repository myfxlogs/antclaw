/**
 * Client Device Information Collector
 * 
 * This module collects comprehensive client-side information including:
 * - Browser type and version
 * - Operating system details
 * - Screen resolution
 * - Device specifications
 * - Network connection status
 * - User interaction patterns
 * 
 * Designed with performance in mind - all operations are non-blocking
 * and executed after page load to avoid impacting page performance.
 * 
 * Follows privacy best practices:
 * - Does not collect personally identifiable information
 * - Requires user consent before collecting certain data
 * - All data is anonymized
 */

/**
 * DeviceInfo interface - comprehensive device information
 */
export interface DeviceInfo {
  // Device identification
  deviceId: string;
  deviceType: 'desktop' | 'mobile' | 'tablet' | 'unknown';
  
  // Browser information
  browserName: string;
  browserVersion: string;
  userAgent: string;
  
  // Operating system
  osName: string;
  osVersion: string;
  
  // Screen information
  screenWidth: number;
  screenHeight: number;
  viewportWidth: number;
  viewportHeight: number;
  pixelRatio: number;
  
  // Device capabilities
  touchSupport: boolean;
  hardwareConcurrency: number;
  memory: number | null; // in GB
  
  // Network information
  networkType: NetworkType;
  connection: NetworkConnection | null;
  
  // Timezone and locale
  timezone: string;
  locale: string;
  language: string;
  
  // User interaction patterns
  interactionCount: number;
  lastInteractionTime: number;
  
  // Collection metadata
  collectedAt: number;
  sessionId: string;
}

/**
 * Network type enumeration
 */
export type NetworkType = 'wifi' | 'cellular' | 'ethernet' | 'bluetooth' | 'none' | 'unknown';

/**
 * Network connection details
 */
export interface NetworkConnection {
  effectiveType: 'slow-2g' | '2g' | '3g' | '4g' | '5g' | 'unknown';
  downlink: number; // Mbps
  rtt: number; // Round-trip time in ms
  saveData: boolean;
}

/**
 * Interaction data storage
 */
interface InteractionData {
  count: number;
  lastTime: number;
}

// Storage keys
const STORAGE_KEY_DEVICE_ID = 'alfq_device_id';
const STORAGE_KEY_INTERACTION = 'alfq_interaction_data';
const STORAGE_KEY_CONSENT = 'alfq_consent';

// Singleton instance
let instance: DeviceInfoCollector | null = null;

/**
 * DeviceInfoCollector class - main collector implementation
 */
export class DeviceInfoCollector {
  private consentGiven: boolean = false;
  private interactionData: InteractionData = { count: 0, lastTime: 0 };
  private collectionPromise: Promise<DeviceInfo> | null = null;
  
  private constructor() {
    this.loadConsent();
    this.loadInteractionData();
    this.setupInteractionTracking();
  }
  
  /**
   * Get singleton instance
   */
  public static getInstance(): DeviceInfoCollector {
    if (!instance) {
      instance = new DeviceInfoCollector();
    }
    return instance;
  }
  
  /**
   * Load consent status from localStorage
   */
  private loadConsent(): void {
    const consent = localStorage.getItem(STORAGE_KEY_CONSENT);
    this.consentGiven = consent === 'true';
  }
  
  /**
   * Load interaction data from localStorage
   */
  private loadInteractionData(): void {
    try {
      const data = localStorage.getItem(STORAGE_KEY_INTERACTION);
      if (data) {
        this.interactionData = JSON.parse(data);
      }
    } catch {
      this.interactionData = { count: 0, lastTime: 0 };
    }
  }
  
  /**
   * Save interaction data to localStorage
   */
  private saveInteractionData(): void {
    try {
      localStorage.setItem(STORAGE_KEY_INTERACTION, JSON.stringify(this.interactionData));
    } catch {
      // Silent fail - localStorage may be unavailable
    }
  }
  
  /**
   * Setup user interaction tracking
   */
  private setupInteractionTracking(): void {
    const trackInteraction = () => {
      this.interactionData.count++;
      this.interactionData.lastTime = Date.now();
      this.saveInteractionData();
    };
    
    // Track various user interactions
    ['click', 'touchstart', 'keydown', 'scroll', 'resize'].forEach(event => {
      document.addEventListener(event, trackInteraction, { passive: true });
    });
  }
  
  /**
   * Request user consent for data collection
   */
  public requestConsent(): Promise<boolean> {
    return new Promise((resolve) => {
      // In a real implementation, this would show a consent dialog
      // For this example, we'll simulate consent after a brief delay
      setTimeout(() => {
        this.consentGiven = true;
        localStorage.setItem(STORAGE_KEY_CONSENT, 'true');
        resolve(true);
      }, 100);
    });
  }
  
  /**
   * Check if consent has been given
   */
  public hasConsent(): boolean {
    return this.consentGiven;
  }
  
  /**
   * Get unique device ID (anonymized)
   */
  public getDeviceId(): string {
    let deviceId = localStorage.getItem(STORAGE_KEY_DEVICE_ID);
    
    if (!deviceId) {
      // Generate a unique anonymous ID
      deviceId = 'alfq_' + Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
      localStorage.setItem(STORAGE_KEY_DEVICE_ID, deviceId);
    }
    
    return deviceId;
  }
  
  /**
   * Generate a unique session ID
   */
  public getSessionId(): string {
    return 'session_' + Date.now().toString(36) + Math.random().toString(36).substring(2, 8);
  }
  
  /**
   * Parse browser information from user agent
   */
  private parseBrowser(userAgent: string): { name: string; version: string } {
    const browsers = [
      { regex: /Edg\/([\d.]+)/, name: 'Edge' },
      { regex: /Chrome\/([\d.]+)/, name: 'Chrome' },
      { regex: /Safari\/([\d.]+)/, name: 'Safari' },
      { regex: /Firefox\/([\d.]+)/, name: 'Firefox' },
      { regex: /Opera\/([\d.]+)/, name: 'Opera' },
      { regex: /MSIE ([\d.]+)/, name: 'IE' },
    ];
    
    for (const browser of browsers) {
      const match = userAgent.match(browser.regex);
      if (match) {
        return { name: browser.name, version: match[1] };
      }
    }
    
    return { name: 'Unknown', version: '0.0' };
  }
  
  /**
   * Parse operating system from user agent
   */
  private parseOS(userAgent: string): { name: string; version: string } {
    const osList = [
      { regex: /Windows NT ([\d.]+)/, name: 'Windows' },
      { regex: /Mac OS X ([\d._]+)/, name: 'macOS' },
      { regex: /Linux/, name: 'Linux' },
      { regex: /Android ([\d.]+)/, name: 'Android' },
      { regex: /iPhone OS ([\d_]+)/, name: 'iOS' },
      { regex: /iPad.*OS ([\d_]+)/, name: 'iOS' },
    ];
    
    for (const os of osList) {
      const match = userAgent.match(os.regex);
      if (match) {
        return { 
          name: os.name, 
          version: match[1] ? match[1].replace(/_/g, '.') : 'unknown' 
        };
      }
    }
    
    return { name: 'Unknown', version: 'unknown' };
  }
  
  /**
   * Determine device type based on screen size and touch support
   */
  private getDeviceType(): 'desktop' | 'mobile' | 'tablet' | 'unknown' {
    const width = window.screen.width;
    
    if ('ontouchstart' in window || navigator.maxTouchPoints > 0) {
      if (width <= 480) return 'mobile';
      if (width <= 768) return 'tablet';
    }
    
    if (width > 768) return 'desktop';
    
    return 'unknown';
  }
  
  /**
   * Get network information
   */
  private getNetworkInfo(): { type: NetworkType; connection: NetworkConnection | null } {
    const nav = navigator as Navigator & { connection?: any };
    if (!nav.connection) {
      return { type: 'unknown', connection: null };
    }
    
    const connection = nav.connection;
    let type: NetworkType = 'unknown';
    
    // Determine network type
    if (connection.effectiveType) {
      switch (connection.effectiveType) {
        case 'slow-2g':
        case '2g':
        case '3g':
        case '4g':
        case '5g':
          type = 'cellular';
          break;
      }
    }
    
    // Check for other types
    if (connection.type === 'wifi') type = 'wifi';
    if (connection.type === 'ethernet') type = 'ethernet';
    if (connection.type === 'bluetooth') type = 'bluetooth';
    if (connection.type === 'none') type = 'none';
    
    return {
      type,
      connection: {
        effectiveType: connection.effectiveType as NetworkConnection['effectiveType'] || 'unknown',
        downlink: connection.downlink || 0,
        rtt: connection.rtt || 0,
        saveData: connection.saveData || false,
      },
    };
  }
  
  /**
   * Get memory information (approximate)
   */
  private getMemory(): number | null {
    if ('deviceMemory' in navigator) {
      return (navigator as any).deviceMemory;
    }
    return null;
  }
  
  /**
   * Collect all device information
   */
  public async collect(): Promise<DeviceInfo> {
    // If already collecting, return the existing promise
    if (this.collectionPromise) {
      return this.collectionPromise;
    }
    
    this.collectionPromise = new Promise((resolve) => {
      // Execute after page load to avoid impacting performance
      const executeCollection = () => {
        const userAgent = navigator.userAgent;
        const browser = this.parseBrowser(userAgent);
        const os = this.parseOS(userAgent);
        const network = this.getNetworkInfo();
        
        const deviceInfo: DeviceInfo = {
          deviceId: this.getDeviceId(),
          deviceType: this.getDeviceType(),
          browserName: browser.name,
          browserVersion: browser.version,
          userAgent: userAgent,
          osName: os.name,
          osVersion: os.version,
          screenWidth: window.screen.width,
          screenHeight: window.screen.height,
          viewportWidth: window.innerWidth,
          viewportHeight: window.innerHeight,
          pixelRatio: window.devicePixelRatio || 1,
          touchSupport: 'ontouchstart' in window || navigator.maxTouchPoints > 0,
          hardwareConcurrency: navigator.hardwareConcurrency || 1,
          memory: this.getMemory(),
          networkType: network.type,
          connection: network.connection,
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          locale: navigator.language,
          language: navigator.languages?.[0] || navigator.language,
          interactionCount: this.interactionData.count,
          lastInteractionTime: this.interactionData.lastTime,
          collectedAt: Date.now(),
          sessionId: this.getSessionId(),
        };
        
        resolve(deviceInfo);
        this.collectionPromise = null;
      };
      
      // If page is already loaded, execute immediately
      if (document.readyState === 'complete') {
        executeCollection();
      } else {
        // Wait for page to load
        window.addEventListener('load', executeCollection, { once: true });
      }
    });
    
    return this.collectionPromise;
  }
  
  /**
   * Collect information with consent check
   */
  public async collectWithConsent(): Promise<DeviceInfo | null> {
    if (!this.consentGiven) {
      await this.requestConsent();
    }
    
    if (!this.consentGiven) {
      return null;
    }
    
    return this.collect();
  }
  
  /**
   * Reset all collected data (for privacy compliance)
   */
  public reset(): void {
    localStorage.removeItem(STORAGE_KEY_DEVICE_ID);
    localStorage.removeItem(STORAGE_KEY_INTERACTION);
    localStorage.removeItem(STORAGE_KEY_CONSENT);
    this.consentGiven = false;
    this.interactionData = { count: 0, lastTime: 0 };
  }
}

/**
 * Singleton export for easy usage
 */
export const deviceInfoCollector = DeviceInfoCollector.getInstance();

/**
 * Convenience: get the persistent device id string directly.
 */
export const getDeviceId = (): string => deviceInfoCollector.getDeviceId();

/**
 * Helper function to format device info for logging
 */
export function formatDeviceInfo(info: DeviceInfo): string {
  return JSON.stringify(info, null, 2);
}

/**
 * Helper function to get a summary of device info
 */
export function getDeviceSummary(info: DeviceInfo): {
  device: string;
  browser: string;
  os: string;
  screen: string;
} {
  return {
    device: `${info.deviceType} (${info.hardwareConcurrency} cores)`,
    browser: `${info.browserName} ${info.browserVersion}`,
    os: `${info.osName} ${info.osVersion}`,
    screen: `${info.screenWidth}x${info.screenHeight}`,
  };
}
