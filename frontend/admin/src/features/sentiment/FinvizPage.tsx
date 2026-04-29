import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchFinviz } from './api'

export default function FinvizPage() {
  const [symbol, setSymbol] = useState('AAPL')
  const { state } = useAsync(() => fetchFinviz(symbol), [symbol])
  return (
    <PageShell
      title="Finviz 指标"
      subtitle="股票基本面 + 空头比 + 机构持仓（firecrawl）"
      actions={
        <input className="input w-32" value={symbol} onChange={(e) => setSymbol(e.target.value.toUpperCase())} />
      }
    >
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
