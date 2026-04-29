import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchCBOE } from './api'

export default function CBOEPutCallPage() {
  const { state } = useAsync(() => fetchCBOE(), [])
  return (
    <PageShell title="CBOE Put/Call" subtitle="CBOE 整体期权情绪比">
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
