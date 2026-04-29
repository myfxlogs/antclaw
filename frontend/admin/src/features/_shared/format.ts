// 通用渲染工具：把 RPC 返回的非 React-友好的值（protobuf Timestamp / bigint / 普通对象）
// 安全转换为可作为 React child 的字符串。
//
// 之所以单独抽出：
//   - protobuf-es v2 的 Timestamp 是 { seconds: bigint, nanos: number, $typeName }，
//     直接 {ts} 渲染会触发 React #31 (Objects are not valid as a React child)。
//   - bigint 也无法被 React 直接渲染。
//   - 各页面以前散布着 String(ts) / new Date(...) 的特判，重复且易漏。

export function isProtoTimestamp(v: unknown): v is { seconds: bigint | number; nanos?: number } {
  if (v == null || typeof v !== 'object') return false
  const o = v as Record<string, unknown>
  return (
    (typeof o.seconds === 'bigint' || typeof o.seconds === 'number') &&
    (o.nanos === undefined || typeof o.nanos === 'number')
  )
}

/** Timestamp → JS Date；非法值返回 null。 */
export function tsToDate(ts: unknown): Date | null {
  if (!isProtoTimestamp(ts)) return null
  const sec = typeof ts.seconds === 'bigint' ? Number(ts.seconds) : ts.seconds
  if (!Number.isFinite(sec)) return null
  return new Date(sec * 1000 + Math.floor((ts.nanos ?? 0) / 1e6))
}

/** Timestamp → 本地化字符串；空值返回 fallback。 */
export function formatTs(ts: unknown, fallback = '-'): string {
  const d = tsToDate(ts)
  if (!d) return fallback
  return d.toLocaleString('zh-CN', { hour12: false })
}

/** Timestamp → YYYY-MM-DD；空值返回 fallback。 */
export function formatDate(ts: unknown, fallback = '-'): string {
  const d = tsToDate(ts)
  if (!d) return fallback
  return d.toISOString().slice(0, 10)
}

/** 兜底：把任意值变成 React 安全的字符串。 */
export function toReactText(v: unknown): string {
  if (v == null) return '-'
  if (typeof v === 'bigint') return v.toString()
  if (isProtoTimestamp(v)) return formatTs(v, '-')
  if (typeof v === 'object') {
    try {
      return JSON.stringify(v)
    } catch {
      return '[object]'
    }
  }
  return String(v)
}
