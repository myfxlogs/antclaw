// /ta/amt —— Auction Market Theory（日类型 / 开盘 / 旋转）。
import { useState } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { TAService, GetAmtRequestSchema } from '@antclaw/proto/antclaw/v1/ta_pb'
import { transport } from '../../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../../_shared/AsyncView'
import { JsonView } from '../../_shared/JsonView'

const c = createClient(TAService, transport)

export default function AMTPage() {
  const [pair, setPair] = useState('EURUSD')
  const { state } = useAsync(() => c.getAmt(create(GetAmtRequestSchema, { pair } as any)), [pair])
  return (
    <PageShell
      title="AMT 拍卖市场理论"
      subtitle="日类型 / 开盘 / 旋转"
      actions={<input className="input w-32" value={pair} onChange={(e) => setPair(e.target.value.toUpperCase())} />}
    >
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
