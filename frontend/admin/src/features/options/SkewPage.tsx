// /options/skew —— 风险逆转 (RR) 与蝶式 (BF) 25 delta。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { fetchSkew } from './api'

export default function SkewPage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchSkew(asset), [asset])
  return (
    <PageShell
      title="期权偏度 Skew"
      subtitle="25Δ Risk Reversal 与 Butterfly；正值表示看涨偏度"
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
          <div className="grid grid-cols-2 gap-4">
            <Stat label="25Δ Risk Reversal" value={d.rr25d.toFixed(4)} />
            <Stat label="25Δ Butterfly" value={d.bf25d.toFixed(4)} />
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
      <div className="text-2xl font-bold mt-1">{props.value}</div>
    </div>
  )
}
