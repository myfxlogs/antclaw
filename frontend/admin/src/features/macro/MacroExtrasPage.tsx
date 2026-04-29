// /macro/extras —— 选源 + 选 series → 拉时序。
import { useState } from 'react'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { formatDate, toReactText } from '../_shared/format'
import { fetchMacroSeries } from './api'

const SOURCES = ['bis', 'imf', 'wb', 'ecb', 'eurostat', 'oecd', 'snb', 'te', 'dtcc', 'treasury']

export default function MacroExtrasPage() {
  const [source, setSource] = useState('wb')
  const [seriesId, setSeriesId] = useState('NY.GDP.MKTP.CD')
  const [submit, setSubmit] = useState({ source, seriesId })
  const { state } = useAsync(() => fetchMacroSeries(submit.source, submit.seriesId), [submit])

  return (
    <PageShell title="宏观全谱" subtitle="BIS / IMF / WB / ECB / Eurostat / OECD 等">
      <div className="space-y-4">
        <div className="flex flex-wrap gap-2 items-end">
          <label className="text-sm">
            <span className="text-xs text-gray-500 block">源</span>
            <select className="input" value={source} onChange={(e) => setSource(e.target.value)}>
              {SOURCES.map((s) => <option key={s} value={s}>{s.toUpperCase()}</option>)}
            </select>
          </label>
          <label className="text-sm flex-1 min-w-64">
            <span className="text-xs text-gray-500 block">
              Series ID <span className="text-gray-400">（WB: 直接填 indicator 默认 WLD；或 "USA/NY.GDP.MKTP.CD"）</span>
            </span>
            <input className="input" value={seriesId} onChange={(e) => setSeriesId(e.target.value)} />
          </label>
          <button
            onClick={() => setSubmit({ source, seriesId })}
            className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm"
          >
            查询
          </button>
        </div>
        <AsyncView
          state={state}
          render={(d: any) => {
            const points: { date: unknown; value: unknown }[] = d.points || d.observations || []
            return (
              <div>
                <div className="text-xs text-gray-500 mb-2">{points.length} 个观测值</div>
                <div className="text-xs font-mono bg-gray-50 rounded p-3 max-h-72 overflow-auto">
                  {points.slice(-30).map((p, i) => (
                    <div key={i}>{formatDate(p.date, toReactText(p.date))} → {toReactText(p.value)}</div>
                  ))}
                </div>
              </div>
            )
          }}
        />
      </div>
    </PageShell>
  )
}
