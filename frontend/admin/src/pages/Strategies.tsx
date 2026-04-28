import { useStrategies } from './strategies/hooks'
import { StrategiesView } from './strategies/view'

export default function Strategies() {
  const state = useStrategies()
  return <StrategiesView
    items={state.items}
    loading={state.loading}
    runningId={state.runningId}
    previewRuns={state.previewRuns}
    previewStrategyId={state.previewStrategyId}
    onRefresh={state.load}
    onEnable={state.handleEnable}
    onDisable={state.handleDisable}
    onRun={state.handleRun}
    onLoadRuns={state.loadRuns}
    onClosePreview={state.closePreview}
  />
}
