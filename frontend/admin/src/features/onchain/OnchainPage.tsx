// /onchain —— 链上指标 + 状态卡。
import { useState } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  OnchainService,
  OnchainServiceGetMetricsRequestSchema,
  OnchainServiceGetAnalysisRequestSchema,
} from '@antclaw/proto/antclaw/v1/onchain_pb'
import { transport } from '../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'

const client = createClient(OnchainService, transport)

export default function OnchainPage() {
  const [asset, setAsset] = useState('BTC')
  const metrics = useAsync(() => client.getMetrics(create(OnchainServiceGetMetricsRequestSchema, { asset })), [asset])
  const analysis = useAsync(
    () => client.getAnalysis(create(OnchainServiceGetAnalysisRequestSchema, { asset })),
    [asset],
  )
  return (
    <PageShell
      title="链上分析"
      subtitle="CoinGecko / CoinMetrics 指标 + Regime 卡"
      actions={
        <select className="input w-32" value={asset} onChange={(e) => setAsset(e.target.value)}>
          <option value="BTC">BTC</option>
          <option value="ETH">ETH</option>
        </select>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <section>
          <h3 className="font-medium mb-2">指标</h3>
          <AsyncView state={metrics.state} render={(d) => <JsonView data={d as unknown} />} />
        </section>
        <section>
          <h3 className="font-medium mb-2">状态分析</h3>
          <AsyncView state={analysis.state} render={(d) => <JsonView data={d as unknown} />} />
        </section>
      </div>
    </PageShell>
  )
}
