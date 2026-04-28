import { History, Pencil, Play, Power, PowerOff, RefreshCw, TrendingUp, X } from 'lucide-react'
import type { RunResult, StrategyItem } from './model'

const TH_BASE = 'px-4 py-3 text-xs font-semibold tracking-wide uppercase text-gray-500'
const TD_BASE = 'px-4 py-3 text-sm text-gray-700 align-middle'
const BTN_BASE = 'inline-flex items-center gap-1 px-2.5 py-1 text-xs rounded-md border whitespace-nowrap transition-colors disabled:opacity-60 disabled:cursor-not-allowed'

interface StrategiesViewProps {
  items: StrategyItem[]
  loading: boolean
  runningId: string | null
  previewRuns: RunResult[] | null
  previewStrategyId: string | null
  onRefresh: () => void
  onEnable: (id: string) => void
  onDisable: (id: string) => void
  onRun: (id: string) => void
  onLoadRuns: (id: string) => void
  onClosePreview: () => void
}

export function StrategiesView(props: StrategiesViewProps) {
  const {
    items, loading, runningId, previewRuns, previewStrategyId,
    onRefresh, onEnable, onDisable, onRun, onLoadRuns, onClosePreview,
  } = props

  return (
    <div className="max-w-screen-2xl mx-auto space-y-6">
      {/* 页头 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <TrendingUp className="w-6 h-6 text-blue-600" /> 策略管理
          </h1>
          <p className="text-sm text-gray-500 mt-1">查看、运行、启停所有交易策略</p>
        </div>
        <button
          onClick={onRefresh}
          className="px-4 py-2 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 flex items-center gap-2 shadow-sm"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} /> 刷新
        </button>
      </div>

      {/* 表格容器：横向溢出滚动；表格使用自然布局，min-width 保证大屏不挤窄 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full min-w-[960px] border-collapse">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className={`${TH_BASE} text-left`}>策略</th>
                <th className={`${TH_BASE} text-left w-32`}>标的</th>
                <th className={`${TH_BASE} text-left w-24`}>周期</th>
                <th className={`${TH_BASE} text-left w-24`}>状态</th>
                <th className={`${TH_BASE} text-left w-56`}>最后运行</th>
                <th className={`${TH_BASE} text-right w-72`}>操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {items.map((it) => (
                <tr key={it.id} className="hover:bg-gray-50/70 transition-colors">
                  <td className={TD_BASE}>
                    <div className="font-medium text-gray-900">{it.name}</div>
                    <div className="text-xs text-gray-400 mt-0.5">{it.kind} · {it.schedule_cron}</div>
                  </td>
                  <td className={`${TD_BASE} font-mono whitespace-nowrap`}>{it.symbol}</td>
                  <td className={`${TD_BASE} whitespace-nowrap`}>{it.timeframe}</td>
                  <td className={TD_BASE}>
                    <span
                      className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                        it.enabled
                          ? 'bg-green-50 text-green-700 border border-green-200'
                          : 'bg-gray-50 text-gray-600 border border-gray-200'
                      }`}
                    >
                      {it.enabled ? '启用' : '停用'}
                    </span>
                  </td>
                  <td className={`${TD_BASE} text-gray-500 whitespace-nowrap`}>
                    {it.last_run_at ? new Date(it.last_run_at).toLocaleString('zh-CN') : '—'}
                    {it.last_run_status && (
                      <span
                        className={`ml-2 text-[11px] ${
                          it.last_run_status === 'success' ? 'text-green-600' : 'text-red-600'
                        }`}
                      >
                        {it.last_run_status}
                      </span>
                    )}
                  </td>
                  <td className={`${TD_BASE} text-right`}>
                    <div className="flex items-center justify-end gap-1.5">
                      <button
                        onClick={() => onRun(it.id)}
                        disabled={runningId === it.id}
                        className={`${BTN_BASE} bg-blue-50 text-blue-700 border-blue-200 hover:bg-blue-100`}
                        title="运行一次"
                      >
                        <Play className="w-3 h-3" /> {runningId === it.id ? '运行中…' : '运行'}
                      </button>
                      <button
                        onClick={() => onLoadRuns(it.id)}
                        className={`${BTN_BASE} bg-white text-gray-700 border-gray-200 hover:bg-gray-50`}
                        title="历史记录"
                      >
                        <History className="w-3 h-3" /> 历史
                      </button>
                      {it.enabled ? (
                        <button
                          onClick={() => onDisable(it.id)}
                          className={`${BTN_BASE} bg-orange-50 text-orange-700 border-orange-200 hover:bg-orange-100`}
                          title="停用"
                        >
                          <PowerOff className="w-3 h-3" /> 停用
                        </button>
                      ) : (
                        <button
                          onClick={() => onEnable(it.id)}
                          className={`${BTN_BASE} bg-green-50 text-green-700 border-green-200 hover:bg-green-100`}
                          title="启用"
                        >
                          <Power className="w-3 h-3" /> 启用
                        </button>
                      )}
                      <button
                        onClick={() => alert('编辑功能将在 Connect 全量改造后开放')}
                        className={`${BTN_BASE} bg-white text-gray-700 border-gray-200 hover:bg-gray-50`}
                        title="编辑"
                      >
                        <Pencil className="w-3 h-3" /> 编辑
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!loading && items.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <TrendingUp className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <p>暂无策略</p>
            <p className="text-sm text-gray-400 mt-2">点击"新建策略"创建第一个策略</p>
          </div>
        )}
        {loading && (
          <div className="text-center py-12">
            <RefreshCw className="w-8 h-8 animate-spin mx-auto text-blue-600" />
          </div>
        )}
      </div>

      {previewStrategyId && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-4xl w-full max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold">历史运行记录</h2>
              <button onClick={onClosePreview} className="p-2 hover:bg-gray-100 rounded-lg" aria-label="关闭">
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="flex-1 overflow-auto p-6">
              {previewRuns && previewRuns.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full min-w-[720px] text-sm">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className={`${TH_BASE} text-left`}>时间</th>
                        <th className={`${TH_BASE} text-left w-24`}>状态</th>
                        <th className={`${TH_BASE} text-right w-28`}>总收益</th>
                        <th className={`${TH_BASE} text-right w-24`}>夏普</th>
                        <th className={`${TH_BASE} text-right w-28`}>最大回撤</th>
                        <th className={`${TH_BASE} text-center w-20`}>模拟</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {previewRuns.map((run) => (
                        <tr key={run.run_id} className="hover:bg-gray-50">
                          <td className={`${TD_BASE} whitespace-nowrap`}>{new Date(run.finished_at).toLocaleString('zh-CN')}</td>
                          <td className={TD_BASE}>
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                                run.status === 'success'
                                  ? 'bg-green-50 text-green-700 border border-green-200'
                                  : 'bg-red-50 text-red-700 border border-red-200'
                              }`}
                            >
                              {run.status}
                            </span>
                          </td>
                          <td className={`${TD_BASE} text-right font-mono`}>{(run.metrics?.total_return * 100).toFixed(2)}%</td>
                          <td className={`${TD_BASE} text-right font-mono`}>{run.metrics?.sharpe?.toFixed(2)}</td>
                          <td className={`${TD_BASE} text-right font-mono`}>{(run.metrics?.max_drawdown * 100).toFixed(2)}%</td>
                          <td className={`${TD_BASE} text-center`}>{run.mock ? '✓' : ''}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">暂无运行记录</div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
