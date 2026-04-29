// /macro/fedwatch —— FOMC 加息概率柱状图。
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { fetchFOMC } from './api'

export default function FedWatchPage() {
  const { state } = useAsync(() => fetchFOMC(), [])
  return (
    <PageShell title="FedWatch" subtitle="CME 联邦基金期货隐含的 FOMC 加息概率">
      <AsyncView
        state={state}
        render={(d: any) => {
          const probs: { rate: string; probability: number }[] = d.probabilities || d.buckets || []
          const total = probs.reduce((s, p) => s + (p.probability || 0), 0)
          return (
            <div className="space-y-2">
              <div className="text-xs text-gray-500">合计：{total.toFixed(1)} % · 共 {probs.length} 档</div>
              {probs.map((p, i) => (
                <div key={i} className="flex items-center gap-2">
                  <span className="w-32 text-sm">{p.rate}</span>
                  <div className="flex-1 bg-gray-100 rounded h-4 overflow-hidden">
                    <div className="bg-indigo-500 h-4" style={{ width: `${Math.min(p.probability, 100)}%` }} />
                  </div>
                  <span className="w-16 text-right text-sm">{(p.probability ?? 0).toFixed(1)}%</span>
                </div>
              ))}
            </div>
          )
        }}
      />
    </PageShell>
  )
}
