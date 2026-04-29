// /options/iv-surface —— 隐含波动率曲面（按到期 + 行权价聚合）。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { fetchIVSurface } from './api'

export default function IVSurfacePage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchIVSurface(asset), [asset])
  return (
    <PageShell
      title="IV Surface"
      subtitle="Deribit 期权隐含波动率三维点；按到期分组展示前 6 档"
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
        render={(d) => {
          const byExpiry = new Map<string, { strike: number; iv: number }[]>()
          for (const p of d.points) {
            const k = `${p.dte}d`
            const arr = byExpiry.get(k) || []
            arr.push({ strike: p.strike, iv: p.iv })
            byExpiry.set(k, arr)
          }
          const expiries = Array.from(byExpiry.keys()).slice(0, 6)
          return (
            <div className="space-y-4">
              <div className="text-sm text-gray-500">共 {d.points.length} 个点，{byExpiry.size} 档到期</div>
              <div className="grid gap-3">
                {expiries.map((exp) => (
                  <ExpirySlice key={exp} expiry={exp} points={byExpiry.get(exp)!} />
                ))}
              </div>
            </div>
          )
        }}
      />
    </PageShell>
  )
}

function ExpirySlice(props: { expiry: string; points: { strike: number; iv: number }[] }) {
  const sorted = [...props.points].sort((a, b) => a.strike - b.strike)
  const minIV = Math.min(...sorted.map((p) => p.iv))
  const maxIV = Math.max(...sorted.map((p) => p.iv))
  return (
    <div className="border rounded p-3">
      <div className="font-medium text-sm mb-2">{props.expiry} · IV {minIV.toFixed(2)}~{maxIV.toFixed(2)}</div>
      <div className="flex items-end gap-0.5 h-16">
        {sorted.map((p) => {
          const range = maxIV - minIV || 1
          const h = ((p.iv - minIV) / range) * 100
          return (
            <div
              key={p.strike}
              className="flex-1 bg-blue-500"
              title={`${p.strike} → IV ${p.iv.toFixed(3)}`}
              style={{ height: `${Math.max(h, 5)}%` }}
            />
          )
        })}
      </div>
    </div>
  )
}
