import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchMyFXBook } from './api'

export default function MyFXBookPage() {
  const [pair, setPair] = useState('EURUSD')
  const { state } = useAsync(() => fetchMyFXBook(pair), [pair])
  return (
    <PageShell
      title="MyFXBook 持仓情绪"
      subtitle="社区多空持仓比"
      actions={
        <input className="input w-32" value={pair} onChange={(e) => setPair(e.target.value.toUpperCase())} />
      }
    >
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
