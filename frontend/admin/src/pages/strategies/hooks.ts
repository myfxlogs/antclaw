import { useEffect, useState } from 'react'
import { disableStrategy, enableStrategy, listStrategies, listStrategyRuns, runStrategy } from '../../lib/api'
import type { RunResult, StrategyItem } from './model'

export function useStrategies() {
  const [items, setItems] = useState<StrategyItem[]>([])
  const [loading, setLoading] = useState(true)
  const [runningId, setRunningId] = useState<string | null>(null)
  const [previewRuns, setPreviewRuns] = useState<RunResult[] | null>(null)
  const [previewStrategyId, setPreviewStrategyId] = useState<string | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const data = await listStrategies()
      setItems(data.items || [])
    } catch (err) {
      console.error('failed to load strategies', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const handleEnable = async (id: string) => {
    await enableStrategy(id)
    await load()
  }

  const handleDisable = async (id: string) => {
    await disableStrategy(id)
    await load()
  }

  const handleRun = async (id: string) => {
    setRunningId(id)
    try {
      const result = await runStrategy(id)
      alert(`回测完成! 总收益: ${(result.metrics?.total_return * 100).toFixed(2)}%`)
      await load()
    } finally {
      setRunningId(null)
    }
  }

  const loadRuns = async (id: string) => {
    setPreviewStrategyId(id)
    try {
      const data = await listStrategyRuns(id, 20)
      setPreviewRuns(data.items || [])
    } catch (err) {
      console.error('failed to load runs', err)
    }
  }

  const closePreview = () => {
    setPreviewStrategyId(null)
    setPreviewRuns(null)
  }

  return {
    items,
    loading,
    runningId,
    previewRuns,
    previewStrategyId,
    load,
    handleEnable,
    handleDisable,
    handleRun,
    loadRuns,
    closePreview,
  }
}
