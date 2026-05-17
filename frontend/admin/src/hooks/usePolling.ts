// usePolling: polling with backoff, pause-when-hidden, and manual refresh (A13-P2-02)
import { useEffect, useRef, useState, useCallback } from 'react'

interface UsePollingOptions<T> {
  fetcher: () => Promise<T>
  intervalMs: number        // base interval between fetches
  enabled?: boolean         // default true
  pauseWhenHidden?: boolean // default true
}

interface UsePollingResult<T> {
  data: T | null
  loading: boolean
  error: string | null
  lastUpdated: Date | null
  refresh: () => Promise<void>
}

export function usePolling<T>({ fetcher, intervalMs, enabled = true, pauseWhenHidden = true }: UsePollingOptions<T>): UsePollingResult<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const errorCountRef = useRef(0)
  const timerRef = useRef<ReturnType<typeof setTimeout>>()
  const fetcherRef = useRef(fetcher)
  fetcherRef.current = fetcher

  const doFetch = useCallback(async () => {
    setLoading(true)
    try {
      const result = await fetcherRef.current()
      setData(result)
      setError(null)
      setLastUpdated(new Date())
      errorCountRef.current = 0
    } catch (e: any) {
      setError(e?.message || String(e))
      errorCountRef.current++
    } finally {
      setLoading(false)
    }
  }, [])

  // Calculate backoff interval: base * 2^errors, capped at 5 min
  const backoffMs = Math.min(intervalMs * Math.pow(2, errorCountRef.current), 300000)

  useEffect(() => {
    if (!enabled) return

    let visibilityPaused = false

    const handleVisibility = () => {
      if (pauseWhenHidden && document.hidden) {
        visibilityPaused = true
        if (timerRef.current) clearTimeout(timerRef.current)
      } else if (visibilityPaused) {
        visibilityPaused = false
        doFetch()
        schedule()
      }
    }

    const schedule = () => {
      const delay = errorCountRef.current > 2 ? backoffMs : intervalMs
      timerRef.current = setTimeout(() => {
        doFetch()
        schedule()
      }, delay)
    }

    doFetch()
    schedule()

    if (pauseWhenHidden) {
      document.addEventListener('visibilitychange', handleVisibility)
    }

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      document.removeEventListener('visibilitychange', handleVisibility)
    }
  }, [enabled, doFetch, backoffMs, intervalMs, pauseWhenHidden])

  return { data, loading, error, lastUpdated, refresh: doFetch }
}
