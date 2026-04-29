// /signals/regime —— Macro / Vol / Liquidity 三联状态条。
import { useState } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { RegimeService, GetOverlayRequestSchema } from '@antclaw/proto/antclaw/v1/regime_pb'
import { transport } from '../../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../../_shared/AsyncView'

const c = createClient(RegimeService, transport)

const SYMBOL_PRESETS = ['BTCUSDT', 'ETHUSDT', 'EURUSD', 'GBPUSD', 'USDJPY', 'XAUUSD', 'SPY', 'QQQ']
const TF_PRESETS = ['D', '4H', '1H']

export default function RegimeOverlayPage() {
  const [symbol, setSymbol] = useState('BTCUSDT')
  const [timeframe, setTimeframe] = useState('D')
  const [submitted, setSubmitted] = useState({ symbol, timeframe })
  const { state } = useAsync(
    () => c.getOverlay(create(GetOverlayRequestSchema, submitted)),
    [submitted],
  )
  return (
    <PageShell
      title="Regime Overlay"
      subtitle="宏观 / 波动 / 流动性 三联状态"
      actions={
        <div className="flex gap-2 items-end">
          <label className="text-sm">
            <span className="text-xs text-gray-500 block">Symbol</span>
            <input
              className="input w-32"
              list="regime-symbols"
              value={symbol}
              onChange={(e) => setSymbol(e.target.value.trim().toUpperCase())}
            />
            <datalist id="regime-symbols">
              {SYMBOL_PRESETS.map((s) => <option key={s} value={s} />)}
            </datalist>
          </label>
          <label className="text-sm">
            <span className="text-xs text-gray-500 block">Timeframe</span>
            <select className="input" value={timeframe} onChange={(e) => setTimeframe(e.target.value)}>
              {TF_PRESETS.map((t) => <option key={t} value={t}>{t}</option>)}
            </select>
          </label>
          <button
            disabled={!symbol}
            onClick={() => setSubmitted({ symbol, timeframe })}
            className="px-3 py-1.5 bg-blue-600 text-white rounded text-sm disabled:bg-gray-300"
          >
            查询
          </button>
        </div>
      }
    >
      <AsyncView
        state={state}
        render={(d: any) => {
          const strips: { name: string; state: string; tone?: string }[] = d.strips ||
            [
              { name: 'Macro', state: d.macro || '-' },
              { name: 'Vol', state: d.vol || '-' },
              { name: 'Liquidity', state: d.liquidity || '-' },
            ]
          return (
            <div className="space-y-3">
              {strips.map((s, i) => (
                <div key={i} className="flex items-center gap-3 p-3 border rounded">
                  <span className="w-24 font-medium">{s.name}</span>
                  <span className={`px-3 py-1 rounded text-sm ${tone(s.state)}`}>{s.state}</span>
                </div>
              ))}
            </div>
          )
        }}
      />
    </PageShell>
  )
}

function tone(state: string): string {
  const v = (state || '').toLowerCase()
  if (v.includes('risk_on') || v.includes('bull') || v.includes('expansion')) return 'bg-green-100 text-green-700'
  if (v.includes('risk_off') || v.includes('bear') || v.includes('contraction')) return 'bg-red-100 text-red-700'
  return 'bg-gray-100 text-gray-700'
}
