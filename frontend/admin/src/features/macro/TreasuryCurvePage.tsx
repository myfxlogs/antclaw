// /macro/treasury —— 美债收益率曲线。
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { fetchTreasuryCurve } from './api'

export default function TreasuryCurvePage() {
  const { state } = useAsync(() => fetchTreasuryCurve(), [])
  return (
    <PageShell title="美债收益率曲线" subtitle="home.treasury.gov 真实数据">
      <AsyncView
        state={state}
        render={(d: any) => {
          const tenors: { tenor: string; yield: number }[] = d.tenors || d.points || []
          const max = Math.max(...tenors.map((t) => t.yield), 1)
          return (
            <div>
              <div className="text-xs text-gray-500 mb-2">{tenors.length} 个期限</div>
              <div className="space-y-1">
                {tenors.map((t, i) => (
                  <div key={i} className="flex items-center gap-2 text-sm">
                    <span className="w-16">{t.tenor}</span>
                    <div className="flex-1 bg-gray-100 rounded h-3 overflow-hidden">
                      <div className="bg-emerald-500 h-3" style={{ width: `${(t.yield / max) * 100}%` }} />
                    </div>
                    <span className="w-16 text-right">{t.yield.toFixed(2)}%</span>
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
