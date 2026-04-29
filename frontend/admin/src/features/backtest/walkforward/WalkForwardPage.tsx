// /backtest/walkforward —— 配置 → 触发 → 表格展示折结果与交易明细。
import { useState } from 'react'
import { PageShell } from '../../_shared/AsyncView'
import { runWalkforward, getWalkforwardResult, getTrades, WFParams } from './api'

const today = () => new Date().toISOString().slice(0, 10)
const daysAgo = (n: number) => {
  const d = new Date()
  d.setDate(d.getDate() - n)
  return d.toISOString().slice(0, 10)
}

export default function WalkForwardPage() {
  const [p, setP] = useState<WFParams>({
    strategy: 'sma_crossover',
    symbols: ['EURUSD'],
    fromDate: daysAgo(300),
    toDate: today(),
    folds: 3,
    trainRatio: 0.7,
  })
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [folds, setFolds] = useState<any[]>([])
  const [trades, setTrades] = useState<any[]>([])

  const onRun = async () => {
    setError(null)
    setRunning(true)
    try {
      const r = await runWalkforward(p)
      const res = await getWalkforwardResult(r.jobId)
      const tr = await getTrades(r.jobId).catch(() => ({ trades: [] as any[] }))
      setFolds((res.folds as any[]) || [])
      setTrades((tr.trades as any[]) || [])
    } catch (e: any) {
      setError(e?.message || String(e))
    } finally {
      setRunning(false)
    }
  }

  return (
    <PageShell title="Walk-Forward 回测" subtitle="K 折滚动 IS/OOS + 交易明细持久化">
      <div className="space-y-6">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
          <Field label="策略">
            <input className="input" value={p.strategy} onChange={(e) => setP({ ...p, strategy: e.target.value })} />
          </Field>
          <Field label="标的（逗号分隔）">
            <input
              className="input"
              value={p.symbols.join(',')}
              onChange={(e) => setP({ ...p, symbols: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) })}
            />
          </Field>
          <Field label="起始日">
            <input className="input" value={p.fromDate} onChange={(e) => setP({ ...p, fromDate: e.target.value })} />
          </Field>
          <Field label="结束日">
            <input className="input" value={p.toDate} onChange={(e) => setP({ ...p, toDate: e.target.value })} />
          </Field>
          <Field label="折数">
            <input
              type="number"
              className="input"
              value={p.folds}
              onChange={(e) => setP({ ...p, folds: parseInt(e.target.value, 10) || 3 })}
            />
          </Field>
          <Field label="训练占比">
            <input
              type="number"
              step="0.05"
              className="input"
              value={p.trainRatio}
              onChange={(e) => setP({ ...p, trainRatio: parseFloat(e.target.value) || 0.7 })}
            />
          </Field>
        </div>
        <button
          onClick={onRun}
          disabled={running}
          className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
        >
          {running ? '运行中...' : '运行 Walk-Forward'}
        </button>
        {error && <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded">{error}</div>}
        <section>
          <h3 className="font-medium mb-2">折结果（{folds.length}）</h3>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 text-gray-600">
              <tr>
                <th className="text-left py-1 px-2">#</th>
                <th className="text-left py-1 px-2">训练区间</th>
                <th className="text-left py-1 px-2">测试区间</th>
                <th className="text-right py-1 px-2">IS Sharpe</th>
                <th className="text-right py-1 px-2">OOS Sharpe</th>
              </tr>
            </thead>
            <tbody>
              {folds.map((f, i) => (
                <tr key={i} className="border-b">
                  <td className="py-1 px-2">{f.foldIdx ?? i + 1}</td>
                  <td className="py-1 px-2">{f.trainFrom} ~ {f.trainTo}</td>
                  <td className="py-1 px-2">{f.testFrom} ~ {f.testTo}</td>
                  <td className="py-1 px-2 text-right">{Number(f.inSampleSharpe ?? 0).toFixed(3)}</td>
                  <td className="py-1 px-2 text-right">{Number(f.oosSharpe ?? 0).toFixed(3)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
        <section>
          <h3 className="font-medium mb-2">交易明细（{trades.length}）</h3>
          {trades.length === 0 ? (
            <div className="text-sm text-gray-400">无</div>
          ) : (
            <div className="text-xs font-mono bg-gray-50 rounded p-3 max-h-64 overflow-auto">
              {trades.slice(0, 50).map((t, i) => (
                <div key={i}>
                  #{t.seq ?? i + 1} {t.side} entry={t.entryPrice} exit={t.exitPrice} pnl={t.pnl}
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </PageShell>
  )
}

function Field(props: { label: string; children: any }) {
  return (
    <label className="block">
      <span className="text-xs text-gray-500">{props.label}</span>
      <div className="mt-1">{props.children}</div>
    </label>
  )
}
