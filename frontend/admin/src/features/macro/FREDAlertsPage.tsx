// /macro/fred-alerts —— 订阅 SSE /sse/macro_alerts。
import { PageShell } from '../_shared/AsyncView'
import { useSSE } from '../_shared/sse'

type MacroAlert = { kind: string; series: string; severity: string; message: string }

export default function FREDAlertsPage() {
  const { items, error } = useSSE<MacroAlert>('macro_alerts', 50)
  return (
    <PageShell
      title="FRED / 宏观告警"
      subtitle={`实时 SSE 通道：/sse/macro_alerts ${error ? `（${error}）` : ''}`}
    >
      {items.length === 0 ? (
        <div className="text-sm text-gray-400">等待事件...</div>
      ) : (
        <ul className="space-y-1">
          {items.map((a, i) => (
            <li key={i} className="flex items-center gap-2 text-sm border-b py-1">
              <span className="px-2 py-0.5 text-xs rounded bg-orange-100 text-orange-700">{a.kind}</span>
              <span className="font-mono text-xs text-gray-500">{a.series}</span>
              <span className="text-xs text-gray-400">[{a.severity}]</span>
              <span className="text-gray-700">{a.message}</span>
            </li>
          ))}
        </ul>
      )}
    </PageShell>
  )
}
