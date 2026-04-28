export interface StrategyItem {
  id: string
  name: string
  kind: string
  symbol: string
  timeframe: string
  enabled: boolean
  status: string
  schedule_cron: string
  last_run_at?: string
  last_run_status?: string
  updated_at: string
}

export interface RunResult {
  run_id: string
  strategy_id: string
  started_at: string
  finished_at: string
  status: string
  metrics: Record<string, number>
  mock: boolean
}
