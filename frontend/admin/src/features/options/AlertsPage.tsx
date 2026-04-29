// /options/alerts —— 期权异动告警；优先 SSE，回退至 Unary 拉取。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { useSSE } from '../_shared/sse'
import { fetchIVAlerts } from './api'

type StreamAlert = { kind: string; asset: string; severity: string; message: string }

export default function AlertsPage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchIVAlerts(asset), [asset])
  const sse = useSSE<StreamAlert>('options_alerts', 30)

  return (
    <PageShell
      title="期权告警"
      subtitle={`SSE 实时通道：/sse/options_alerts ${sse.error ? `（${sse.error}）` : ''}`}
      actions={
        <select
          value={asset}
          onChange={(e) => setAsset(e.target.value)}
          className="border rounded px-3 py-1.5 text-sm"
        >
          <option value="BTC">BTC</option>
          <option value="ETH">ETH</option>
        </select>
      }
    >
      <div className="space-y-6">
        <section>
          <h3 className="font-medium mb-2">实时事件</h3>
          {sse.items.length === 0 ? (
            <div className="text-sm text-gray-400">等待 SSE 事件...</div>
          ) : (
            <ul className="space-y-1">
              {sse.items.map((a, i) => (
                <li key={i} className="flex items-center gap-2 text-sm border-b py-1">
                  <span className="px-2 py-0.5 text-xs rounded bg-orange-100 text-orange-700">
                    {a.kind}
                  </span>
                  <span className="font-mono text-xs text-gray-500">{a.asset}</span>
                  <span className="text-gray-700">{a.message}</span>
                </li>
              ))}
            </ul>
          )}
        </section>
        <section>
          <h3 className="font-medium mb-2">最近告警快照</h3>
          <AsyncView
            state={state}
            render={(d) => (
              <ul className="space-y-1">
                {(d.alerts || []).map((a, i) => (
                  <li key={i} className="text-sm">
                    <span className="font-mono text-xs text-gray-500 mr-2">{a.kind}</span>
                    <span className="text-xs text-gray-400 mr-2">[{a.severity}]</span>
                    {a.message}
                  </li>
                ))}
              </ul>
            )}
            emptyText="暂无告警"
          />
        </section>
      </div>
    </PageShell>
  )
}
