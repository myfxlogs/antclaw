// /vol/cross —— Cross-asset volatility 概览：VIX (股) + DVOL (BTC)。
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchVix, fetchDvol } from './api'

export default function CrossVolPage() {
  const vix = useAsync(() => fetchVix(), [])
  const dvol = useAsync(() => fetchDvol('BTC'), [])
  return (
    <PageShell title="跨市场波动率" subtitle="VIX (CBOE) 与 DVOL (Deribit BTC) 对照">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <section>
          <h3 className="font-medium mb-2">VIX</h3>
          <AsyncView state={vix.state} render={(d) => <JsonView data={d as unknown} />} />
        </section>
        <section>
          <h3 className="font-medium mb-2">DVOL · BTC</h3>
          <AsyncView state={dvol.state} render={(d) => <JsonView data={d as unknown} />} />
        </section>
      </div>
    </PageShell>
  )
}
