// useSSE：订阅 EventSource 事件并把最近 N 条聚合到 state；卸载自动 close。
import { useEffect, useState } from 'react'

const RAW_BASE = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8082'

// 兼容三种部署：
//   1) "/"          —— 同源（nginx 反代），SSE 路径 = "/sse/<ch>"
//   2) "/api"       —— 同源带前缀，SSE = "/api/sse/<ch>"
//   3) "http://..." —— 跨源，SSE = "<host>/sse/<ch>"
// 直接字符串拼接会产生 "//sse/x"，浏览器误解析为协议相对 URL（host=sse）→ ERR_NAME_NOT_RESOLVED。
function joinSSE(channel: string): string {
  const base = RAW_BASE.replace(/\/+$/, '') // 去掉末尾所有斜杠
  return `${base}/sse/${channel}`
}

export function useSSE<T = unknown>(channel: string, max = 50) {
  const [items, setItems] = useState<T[]>([])
  const [error, setError] = useState<string | null>(null)
  useEffect(() => {
    const url = joinSSE(channel)
    const es = new EventSource(url, { withCredentials: false })
    es.onmessage = (ev) => {
      try {
        const obj = JSON.parse(ev.data)
        setItems((prev: T[]) => [obj as T, ...prev].slice(0, max))
      } catch {
        // 非 JSON 行直接忽略
      }
    }
    es.onerror = () => {
      setError('SSE 连接异常或已断开')
    }
    return () => es.close()
  }, [channel, max])
  return { items, error }
}
