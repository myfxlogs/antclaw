import { useEffect, useState } from 'react'
import { Database, RefreshCw, Eye, X } from 'lucide-react'
import { getDataSummary, getDataPreview } from '../lib/api'

interface DataSourceSummary {
  job_id: string
  name: string
  table: string
  count: number
  latest_time?: number
  error?: string
}

interface PreviewResponse {
  job_id: string
  table: string
  time_col: string
  columns: string[]
  rows: Record<string, unknown>[]
  total_sampled: number
}

export default function DataSummary() {
  const [items, setItems] = useState<DataSourceSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [updatedAt, setUpdatedAt] = useState<number | null>(null)
  const [previewJobId, setPreviewJobId] = useState<string | null>(null)
  const [previewData, setPreviewData] = useState<PreviewResponse | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const json = await getDataSummary()
      setItems(json.items || [])
      setUpdatedAt(json.updated_at || null)
    } catch (err) {
      console.error('failed to load data summary', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const formatTime = (ts?: number) => {
    if (!ts) return '-'
    return new Date(ts * 1000).toLocaleString('zh-CN')
  }

  const loadPreview = async (jobId: string) => {
    setPreviewJobId(jobId)
    setPreviewLoading(true)
    try {
      const json = await getDataPreview(jobId, 50)
      setPreviewData(json)
    } catch (err) {
      console.error('failed to load preview', err)
    } finally {
      setPreviewLoading(false)
    }
  }

  const closePreview = () => {
    setPreviewJobId(null)
    setPreviewData(null)
  }

  const formatCell = (val: unknown): string => {
    if (val === null || val === undefined) return 'null'
    if (typeof val === 'string') return val
    if (typeof val === 'number') return String(val)
    if (typeof val === 'boolean') return String(val)
    return JSON.stringify(val).slice(0, 100)
  }

  // 表头中文映射
  const columnNameMap: Record<string, string> = {
    'time': '时间',
    'symbol': '标的',
    'open': '开盘价',
    'high': '最高价',
    'low': '最低价',
    'close': '收盘价',
    'volume': '成交量',
    'source': '来源',
    'period': '周期',
    'value': '数值',
    'series_id': '指标ID',
    'created_at': '创建时间',
    'updated_at': '更新时间',
    'id': 'ID',
    'job_id': '任务ID',
    'name': '名称',
    'type': '类型',
    'price': '价格',
    'change': '变动',
    'change_percent': '变动率',
    'market_cap': '市值',
    'fear_greed': '恐惧贪婪指数',
    'timestamp': '时间戳',
    'date': '日期',
    'category': '分类',
  }

  const getColumnName = (col: string): string => {
    return columnNameMap[col] || col
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <Database className="w-6 h-6 text-blue-600" /> 采集数据
        </h1>
        <button
          onClick={load}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} /> 刷新
        </button>
      </div>

      {updatedAt && (
        <p className="text-sm text-gray-500">最后更新: {formatTime(updatedAt)}</p>
      )}

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">任务</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">数据库表</th>
              <th className="px-6 py-3 text-right text-sm font-medium text-gray-500">行数</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">最新时间</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">说明</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {items.map((it) => (
              <tr key={it.job_id} className="hover:bg-gray-50 group">
                <td className="px-6 py-4">
                  <div className="font-medium">{it.name}</div>
                  <div className="text-xs text-gray-400">{it.job_id}</div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-700 font-mono">{it.table}</td>
                <td className="px-6 py-4 text-right tabular-nums">{it.count.toLocaleString()}</td>
                <td className="px-6 py-4 text-sm text-gray-500">{formatTime(it.latest_time)}</td>
                <td className="px-6 py-4 text-sm">
                  <div className="flex items-center gap-3">
                    {it.error ? (
                      <span className="text-red-600">{it.error}</span>
                    ) : it.count === 0 ? (
                      <span className="text-yellow-600">无数据</span>
                    ) : (
                      <span className="text-green-600">已采集</span>
                    )}
                    {it.count > 0 && (
                      <button
                        onClick={() => loadPreview(it.job_id)}
                        className="flex items-center gap-1 px-2 py-1 text-xs bg-blue-50 text-blue-600 rounded hover:bg-blue-100"
                      >
                        <Eye className="w-3 h-3" /> 预览
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Preview Modal */}
      {previewJobId && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-xl max-w-6xl w-full max-h-[90vh] flex flex-col">
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div>
                <h2 className="text-lg font-semibold">数据预览</h2>
                <p className="text-sm text-gray-500">
                  {previewData?.table} • {previewData?.total_sampled} 条记录
                </p>
              </div>
              <button onClick={closePreview} className="p-2 hover:bg-gray-100 rounded-lg">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex-1 overflow-auto p-6">
              {previewLoading ? (
                <div className="flex items-center justify-center py-12">
                  <RefreshCw className="w-8 h-8 animate-spin text-blue-600" />
                </div>
              ) : previewData?.rows && previewData.rows.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 sticky top-0">
                      <tr>
                        {previewData.columns.map((col) => (
                          <th key={col} className="px-3 py-2 text-left font-medium text-gray-600 whitespace-nowrap">
                            {getColumnName(col)}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y">
                      {previewData.rows.map((row, idx) => (
                        <tr key={idx} className="hover:bg-gray-50">
                          {previewData.columns.map((col) => (
                            <td key={col} className="px-3 py-2 text-gray-700 whitespace-nowrap max-w-xs truncate" title={formatCell(row[col])}>
                              {formatCell(row[col])}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">暂无数据</div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
