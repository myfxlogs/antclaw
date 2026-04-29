// /options/gex —— 显示总 GEX、零伽马点和分桶柱状图。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { fetchGEX } from './api'

export default function GEXPage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchGEX(asset), [asset])
  return (
    <PageShell
      title="Gamma Exposure"
      subtitle="按行权价聚合的总伽马敞口；零伽马点提示市场重力位"
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
      <AsyncView
        state={state}
        render={(d) => (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <Stat label="总 GEX" value={d.totalGex.toFixed(2)} />
              <Stat label="零伽马点" value={d.zeroGamma.toFixed(2)} />
            </div>
            <div>
              <h3 className="font-medium mb-2">分桶 ({d.strikes.length})</h3>
              <BucketBars
                buckets={d.strikes.map((s) => ({ strike: s.strike, gex: s.callGex + s.putGex }))}
              />
            </div>
          </div>
        )}
      />
    </PageShell>
  )
}

function Stat(props: { label: string; value: string }) {
  return (
    <div className="bg-gray-50 rounded p-4">
      <div className="text-xs text-gray-500">{props.label}</div>
      <div className="text-xl font-bold mt-1">{props.value}</div>
    </div>
  )
}

function BucketBars(props: { buckets: { strike: number; gex: number }[] }) {
  const items = props.buckets.slice(0, 24)
  const max = Math.max(...items.map((b) => Math.abs(b.gex)), 1)
  return (
    <div className="space-y-1">
      {items.map((b) => {
        const w = (Math.abs(b.gex) / max) * 100
        const positive = b.gex >= 0
        return (
          <div key={b.strike} className="flex items-center gap-2 text-xs">
            <span className="w-20 text-right text-gray-600">{b.strike}</span>
            <div className="flex-1 h-3 bg-gray-100 rounded overflow-hidden">
              <div
                className={`h-3 ${positive ? 'bg-green-500' : 'bg-red-500'}`}
                style={{ width: `${w}%` }}
              />
            </div>
            <span className="w-20 text-gray-600">{b.gex.toFixed(0)}</span>
          </div>
        )
      })}
    </div>
  )
}
