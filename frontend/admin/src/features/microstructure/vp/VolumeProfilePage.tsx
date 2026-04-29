// /microstructure/vp —— 体积区间分布。
import { useState } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { TAService, GetVolumeProfileRequestSchema } from '@antclaw/proto/antclaw/v1/ta_pb'
import { transport } from '../../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../../_shared/AsyncView'

const c = createClient(TAService, transport)

export default function VolumeProfilePage() {
  const [pair, setPair] = useState('EURUSD')
  const { state } = useAsync(() => c.getVolumeProfile(create(GetVolumeProfileRequestSchema, { pair } as any)), [pair])
  return (
    <PageShell
      title="体积区间分布"
      subtitle="POC / VAH / VAL · 价格-成交量直方图"
      actions={<input className="input w-32" value={pair} onChange={(e) => setPair(e.target.value.toUpperCase())} />}
    >
      <AsyncView
        state={state}
        render={(d: any) => {
          const bins: { price: number; volume: number }[] = d.bins || d.profile || []
          const maxV = Math.max(...bins.map((b) => b.volume), 1)
          const poc = d.poc ?? d.pointOfControl ?? 0
          return (
            <div className="space-y-3">
              <div className="text-sm text-gray-600">POC：<b>{Number(poc).toFixed(4)}</b> · 共 {bins.length} 桶</div>
              <div className="space-y-0.5">
                {bins.map((b, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs">
                    <span className="w-20 text-right text-gray-600">{b.price.toFixed(4)}</span>
                    <div className="flex-1 bg-gray-100 rounded h-3 overflow-hidden">
                      <div className="bg-cyan-500 h-3" style={{ width: `${(b.volume / maxV) * 100}%` }} />
                    </div>
                    <span className="w-16 text-right text-gray-500">{Number(b.volume).toFixed(0)}</span>
                  </div>
                ))}
              </div>
            </div>
          )
        }}
      />
    </PageShell>
  )
}
