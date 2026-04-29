import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchInsider } from './api'

export default function InsiderTradesPage() {
  const { state } = useAsync(() => fetchInsider(), [])
  return (
    <PageShell title="内部人交易" subtitle="OpenInsider · firecrawl 抽取">
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
