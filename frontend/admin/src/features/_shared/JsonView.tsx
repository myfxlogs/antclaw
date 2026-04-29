// 通用对象渲染组件：兜底用，把 RPC 响应中暂未做定制 UI 的字段以缩进键值对方式展示。
import { isProtoTimestamp, formatTs } from './format'

type Primitive = string | number | boolean | null

function isPrimitive(v: unknown): v is Primitive {
  return v == null || typeof v === 'string' || typeof v === 'number' || typeof v === 'boolean'
}

function fmt(v: unknown): string {
  if (v === null || v === undefined) return '-'
  if (typeof v === 'number') return Number.isInteger(v) ? String(v) : v.toFixed(4)
  return String(v)
}

export function JsonView({ data, depth = 0 }: { data: unknown; depth?: number }) {
  if (isPrimitive(data)) {
    return <span className="text-gray-800">{fmt(data)}</span>
  }
  // bigint 不是 React-safe 的 child，需先转字符串
  if (typeof data === 'bigint') {
    return <span className="text-gray-800">{data.toString()}</span>
  }
  // 把 protobuf Timestamp 当作叶子节点格式化，避免展开 seconds/nanos
  if (isProtoTimestamp(data)) {
    return <span className="text-gray-800">{formatTs(data)}</span>
  }
  if (Array.isArray(data)) {
    if (data.length === 0) return <span className="text-gray-400">[]</span>
    const head = data.slice(0, 30)
    return (
      <div className="space-y-0.5">
        <div className="text-xs text-gray-500">数组 · {data.length} 项{data.length > 30 ? '（仅展示前 30）' : ''}</div>
        {head.map((item, i) => (
          <div key={i} className="pl-3 border-l border-gray-100">
            <span className="text-xs text-gray-400 mr-2">[{i}]</span>
            <JsonView data={item} depth={depth + 1} />
          </div>
        ))}
      </div>
    )
  }
  // object
  const obj = data as Record<string, unknown>
  const entries = Object.entries(obj).filter(([k]) => !k.startsWith('$'))
  return (
    <div className="space-y-1">
      {entries.map(([k, v]) => (
        <div key={k} className="flex gap-3">
          <span className="text-xs font-mono text-gray-500 w-40 truncate">{k}</span>
          <div className="flex-1 text-sm">
            <JsonView data={v} depth={depth + 1} />
          </div>
        </div>
      ))}
    </div>
  )
}
