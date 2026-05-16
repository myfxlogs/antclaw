import { useCallback, useEffect, useState } from 'react'
import { createClient } from '@connectrpc/connect'
import { SignalsService } from '@antclaw/proto/antclaw/v1/signals_pb'
import { transport } from '../_shared/transport'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

const client = createClient(SignalsService, transport)

type SignalView = {
  pair: string
  direction: string
  confidence: number
}

export default function SignalsPage() {
  const [signals, setSignals] = useState<SignalView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const pairs = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'BTCUSD']
      const results: SignalView[] = []
      for (const pair of pairs) {
        try {
          const resp = await client.getBias({ pair, timeframe: '1D' })
          const bias = resp.biases?.[0]
          results.push({
            pair,
            direction: bias?.direction || 'neutral',
            confidence: Math.round((bias?.confidence || 0) * 100),
          })
        } catch {
          results.push({ pair, direction: 'neutral', confidence: 0 })
        }
      }
      setSignals(results)
    } catch (e: any) {
      setError(e?.message || '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const getIcon = (d: string) => {
    if (d === 'bullish') return <TrendingUp className="w-5 h-5 text-green-600" />
    if (d === 'bearish') return <TrendingDown className="w-5 h-5 text-red-600" />
    return <Minus className="w-5 h-5 text-gray-500" />
  }

  const getCls = (d: string) => {
    if (d === 'bullish') return 'bg-green-50 text-green-700 border-green-200'
    if (d === 'bearish') return 'bg-red-50 text-red-700 border-red-200'
    return 'bg-gray-50 text-gray-700 border-gray-200'
  }

  if (loading) return <div className="p-8 animate-pulse text-gray-400">加载信号中...</div>
  if (error) return <div className="p-4 bg-red-50 border border-red-200 rounded text-red-700">{error}</div>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">交易信号</h1>
        <button onClick={load} className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
          刷新
        </button>
      </div>
      <div className="grid gap-4">
        {signals.map((s) => (
          <div key={s.pair} className={`p-6 border rounded-xl ${getCls(s.direction)}`}>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-4">
                <div className="p-2 bg-white rounded-lg shadow-sm">{getIcon(s.direction)}</div>
                <div>
                  <h3 className="text-lg font-semibold">{s.pair}</h3>
                  <p className="text-sm opacity-75">{s.direction === 'bullish' ? '看多' : s.direction === 'bearish' ? '看空' : '中性'}</p>
                </div>
              </div>
              <div className="text-right">
                <p className="text-sm opacity-75">置信度</p>
                <p className="text-xl font-bold">{s.confidence}%</p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
