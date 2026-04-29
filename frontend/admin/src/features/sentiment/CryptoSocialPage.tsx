import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchCryptoSocial } from './api'

export default function CryptoSocialPage() {
  const [asset, setAsset] = useState('BTC')
  const { state } = useAsync(() => fetchCryptoSocial(asset), [asset])
  return (
    <PageShell
      title="加密社交"
      subtitle="LunarCrush / CryptoCompare 社交热度（firecrawl）"
      actions={
        <select className="input w-32" value={asset} onChange={(e) => setAsset(e.target.value)}>
          <option value="BTC">BTC</option>
          <option value="ETH">ETH</option>
          <option value="SOL">SOL</option>
        </select>
      }
    >
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
