// /vol/term —— DVOL 期限结构（用 GetDvol 返回里包含的 term 数据）。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchDvol } from './api'

export default function TermStructurePage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchDvol(asset), [asset])
  return (
    <PageShell
      title="期限结构"
      subtitle="DVOL 隐含波动率期限结构"
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
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
