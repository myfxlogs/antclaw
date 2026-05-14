import { useEffect, useRef, useCallback } from 'react';
import { getToken } from '../../features/auth/api';

const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080';
const SSE_URL = `${API_BASE_URL}/sse/notifications`;
const RECONNECT_DELAY = 3000;

type NotificationHandler = (data: string) => void;

export function useSse(onNotification?: NotificationHandler) {
  const abortRef = useRef<AbortController | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout>>();
  const mountedRef = useRef(true);

  const connect = useCallback(() => {
    const token = getToken();
    if (!token || !mountedRef.current) return;

    // Abort previous connection
    abortRef.current?.abort();
    abortRef.current = new AbortController();

    fetch(SSE_URL, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Accept': 'text/event-stream',
      },
      signal: abortRef.current.signal,
    })
      .then(async (response) => {
        if (!response.ok || !response.body) {
          throw new Error(`SSE connection failed: ${response.status}`);
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (mountedRef.current) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          let eventType = '';
          let eventData = '';

          for (const line of lines) {
            if (line.startsWith('event: ')) {
              eventType = line.slice(7).trim();
            } else if (line.startsWith('data: ')) {
              eventData = line.slice(6).trim();
            } else if (line === '' && eventData) {
              // Empty line = end of event
              if (eventType === 'notification' && onNotification) {
                onNotification(eventData);
              }
              eventType = '';
              eventData = '';
            }
            // Ignore `: ping` lines (heartbeat)
          }
        }
      })
      .catch((err) => {
        if (err.name === 'AbortError') return;
        console.error('SSE error:', err);
        // Auto-reconnect
        if (mountedRef.current) {
          reconnectTimerRef.current = setTimeout(connect, RECONNECT_DELAY);
        }
      });
  }, [onNotification]);

  useEffect(() => {
    mountedRef.current = true;
    connect();
    return () => {
      mountedRef.current = false;
      abortRef.current?.abort();
      clearTimeout(reconnectTimerRef.current);
    };
  }, [connect]);

  return { reconnect: connect };
}
