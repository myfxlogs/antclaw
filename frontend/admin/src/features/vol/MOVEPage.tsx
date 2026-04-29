// /vol/move —— MOVE 指数（美债隐含波动）。
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { JsonView } from '../_shared/JsonView'
import { fetchMove } from './api'

export default function MOVEPage() {
  const { state } = useAsync(() => fetchMove(), [])
  return (
    <PageShell title="MOVE 指数" subtitle="美债隐含波动率（Yardeni）">
      <AsyncView state={state} render={(d) => <JsonView data={d as unknown} />} />
    </PageShell>
  )
}
