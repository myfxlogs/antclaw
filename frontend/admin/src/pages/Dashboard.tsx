import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bell, BarChart3, Clock, Activity } from 'lucide-react'
import { getOnlineUsers } from '../lib/api'

interface PushStats {
  totalNotifications: number
  totalPushStateRecords: number
  byType: { pushType: string; count: number }[]
  recent1h: number
  recent24h: number
}

interface OnlineState {
  count: number
  users: { userId: string; codeId: string; remoteAddr: string; connectedAt: number }[]
}

const PUSH_TYPE_LABELS: Record<string, string> = {
  calendar_pre: '日历预提醒',
  calendar_actual: '日历公布',
  surprise: '意外数据',
  digest: '日报摘要',
  weekly_digest: '周报摘要',
  macro_regime: '宏观 Regime',
  macro_fred: 'FRED 异动',
  options_risk: '期权风险',
  onchain_correlation: '链上关联',
  carry: '套息交易',
  regime_transition: 'Regime 转换',
  risk_confluence: '风险共振',
  cot_release: 'COT 发布',
  cot_signal: 'COT 信号',
  calibration: '信号校准',
}

function timeAgo(ts: number): string {
  const diff = Math.floor(Date.now() / 1000 - ts)
  if (diff < 60) return `${diff}s`
  if (diff < 3600) return `${Math.floor(diff / 60)}m`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h`
  return `${Math.floor(diff / 86400)}d`
}

export default function Dashboard() {
  const { t } = useTranslation()
  const [online, setOnline] = useState<OnlineState>({ count: 0, users: [] })
  const [pushStats, setPushStats] = useState<PushStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let active = true
    const fetchData = async () => {
      try {
        const onlineData = await getOnlineUsers()
        if (!active) return
        setOnline(onlineData)

        // 获取推送统计
        try {
          const { getPushStats } = await import('../lib/api')
          setPushStats(await getPushStats())
        } catch { /* push stats optional */ }
      } catch { /* ignore */ } finally {
        if (active) setLoading(false)
      }
    }
    fetchData()
    const timer = setInterval(fetchData, 15000)
    return () => { active = false; clearInterval(timer) }
  }, [])

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">{t('dashboard.title')}</h1>
        <div className="text-gray-400">加载中...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">{t('dashboard.title')}</h1>

      {/* 在线用户 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="p-6 border-b border-gray-100">
          <h2 className="text-lg font-semibold text-gray-800">在线用户</h2>
        </div>
        <div className="p-6">
          <div className={`inline-flex items-center gap-2 px-4 py-3 rounded-lg ${online.count > 0 ? 'bg-green-50' : 'bg-gray-50'}`}>
            <div className={`w-2.5 h-2.5 rounded-full ${online.count > 0 ? 'bg-green-500' : 'bg-gray-300'}`} />
            <span className="text-2xl font-bold text-gray-900">{online.count}</span>
            <span className="text-sm text-gray-500">当前在线</span>
          </div>
          {online.users.length > 0 && (
            <table className="w-full mt-4 text-sm">
              <thead>
                <tr className="border-b text-left text-gray-500">
                  <th className="pb-2">ID</th>
                  <th className="pb-2">来源</th>
                  <th className="pb-2">连接时长</th>
                </tr>
              </thead>
              <tbody>
                {online.users.map(u => (
                  <tr key={u.userId} className="border-b border-gray-50">
                    <td className="py-2 font-mono text-xs">{u.codeId || u.userId.substring(0, 12)}...</td>
                    <td className="py-2 text-gray-500">{u.remoteAddr}</td>
                    <td className="py-2 text-gray-500">{timeAgo(u.connectedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      {/* 推送统计 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        <div className="p-6 border-b border-gray-100">
          <div className="flex items-center gap-3">
            <Bell className="w-5 h-5 text-gray-400" />
            <h2 className="text-lg font-semibold text-gray-800">推送管理</h2>
          </div>
        </div>
        <div className="p-6">
          {pushStats ? (
            <div className="space-y-6">
              {/* 总览卡片 */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="bg-blue-50 rounded-lg p-4">
                  <div className="text-2xl font-bold text-blue-700">{pushStats.totalNotifications}</div>
                  <div className="text-xs text-blue-600 mt-1">通知总数</div>
                </div>
                <div className="bg-purple-50 rounded-lg p-4">
                  <div className="text-2xl font-bold text-purple-700">{pushStats.totalPushStateRecords}</div>
                  <div className="text-xs text-purple-600 mt-1">推送记录</div>
                </div>
                <div className="bg-green-50 rounded-lg p-4">
                  <div className="flex items-center gap-1">
                    <Clock className="w-4 h-4 text-green-500" />
                    <span className="text-2xl font-bold text-green-700">{pushStats.recent1h}</span>
                  </div>
                  <div className="text-xs text-green-600 mt-1">最近 1 小时</div>
                </div>
                <div className="bg-amber-50 rounded-lg p-4">
                  <div className="flex items-center gap-1">
                    <Activity className="w-4 h-4 text-amber-500" />
                    <span className="text-2xl font-bold text-amber-700">{pushStats.recent24h}</span>
                  </div>
                  <div className="text-xs text-amber-600 mt-1">最近 24 小时</div>
                </div>
              </div>

              {/* 按类型统计 */}
              {pushStats.byType.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <BarChart3 className="w-4 h-4 text-gray-400" />
                    <span className="text-sm font-medium text-gray-600">按推送类型</span>
                  </div>
                  <div className="space-y-2">
                    {pushStats.byType.map(item => (
                      <div key={item.pushType} className="flex items-center gap-3">
                        <span className="text-xs text-gray-600 w-32 truncate">
                          {PUSH_TYPE_LABELS[item.pushType] || item.pushType}
                        </span>
                        <div className="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
                          <div
                            className="bg-blue-500 h-4 rounded-full transition-all"
                            style={{ width: `${Math.min(100, (item.count / Math.max(...pushStats.byType.map(t => t.count))) * 100)}%` }}
                          />
                        </div>
                        <span className="text-xs font-mono text-gray-500 w-8 text-right">{item.count}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="text-gray-400 text-sm py-4 text-center">
              推送统计暂不可用（请确保 Worker 正在运行）
            </div>
          )}
          <div className="mt-4 text-xs text-gray-400">每 15 秒自动刷新</div>
        </div>
      </div>
    </div>
  )
}
