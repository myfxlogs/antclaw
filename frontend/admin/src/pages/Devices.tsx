import { useEffect, useState } from 'react'
import { Smartphone, Search, Trash2 } from 'lucide-react'
import { listDevices, deleteDevice, type DeviceInfo } from '../lib/api'

export default function Devices() {
  const [devices, setDevices] = useState<DeviceInfo[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [osFilter, setOsFilter] = useState('')

  const handleDelete = async (deviceId: string) => {
    if (!confirm('确认删除该设备信息？')) return
    try { await deleteDevice(deviceId); refresh() } catch { alert('删除失败') }
  }
  const refresh = () => {
    listDevices({ osTypeFilter: osFilter }).then(d => { setDevices(d.devices); setTotal(d.total) })
  }

  useEffect(() => {
    let active = true
    const fetch = async () => {
      try {
        const data = await listDevices({ osTypeFilter: osFilter })
        if (active) { setDevices(data.devices); setTotal(data.total) }
      } catch { /* ok */ } finally { if (active) setLoading(false) }
    }
    fetch()
    const timer = setInterval(fetch, 30000)
    return () => { active = false; clearInterval(timer) }
  }, [osFilter])

  const byOS = devices.reduce((acc: Record<string, number>, d) => {
    acc[d.osType] = (acc[d.osType] || 0) + 1
    return acc
  }, {})

  const formatTime = (ts: number) => {
    if (!ts) return '-'
    return new Date(ts * 1000).toLocaleString()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Smartphone className="w-6 h-6 text-gray-500" />
        <h1 className="text-2xl font-bold text-gray-900">设备管理</h1>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl shadow-sm border p-4">
          <div className="text-2xl font-bold text-blue-600">{total}</div>
          <div className="text-xs text-gray-500">总设备数</div>
        </div>
        {Object.entries(byOS).map(([os, count]) => (
          <div key={os} className="bg-white rounded-xl shadow-sm border p-4">
            <div className="text-2xl font-bold text-green-600">{count}</div>
            <div className="text-xs text-gray-500">{os}</div>
          </div>
        ))}
      </div>

      {/* 过滤 + 列表 */}
      <div className="bg-white rounded-xl shadow-sm border">
        <div className="p-4 border-b flex items-center gap-3">
          <Search className="w-4 h-4 text-gray-400" />
          <select
            value={osFilter}
            onChange={e => setOsFilter(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-1.5 text-sm"
          >
            <option value="">全部系统</option>
            <option value="android">Android</option>
            <option value="ios">iOS</option>
          </select>
          <span className="text-xs text-gray-400">{loading ? '加载中...' : `${devices.length} 条`}</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-gray-500">
                <th className="p-3 font-medium">设备 ID</th>
                <th className="p-3 font-medium">用户</th>
                <th className="p-3 font-medium">型号</th>
                <th className="p-3 font-medium">系统</th>
                <th className="p-3 font-medium">App 版本</th>
                <th className="p-3 font-medium">网络</th>
                <th className="p-3 font-medium">时区</th>
                <th className="p-3 font-medium">最后更新</th>
                <th className="p-3 font-medium w-16">操作</th>
              </tr>
            </thead>
            <tbody>
              {devices.map(d => (
                <tr key={d.deviceId} className="border-b border-gray-50 hover:bg-gray-50">
                  <td className="p-3 font-mono text-xs text-gray-600 max-w-[140px] truncate" title={d.deviceId}>{d.deviceId}</td>
                  <td className="p-3 text-xs">
                    {d.displayName || d.username ? (
                      <div>
                        <div className="text-gray-700">{d.displayName || d.username}</div>
                        <div className="text-gray-400">{d.codeId || d.userId?.substring(0,8)}</div>
                      </div>
                    ) : <span className="text-gray-300">-</span>}
                  </td>
                  <td className="p-3 text-gray-700">{d.brand} {d.model}</td>
                  <td className="p-3">
                    <span className="text-xs">{d.osType} {d.osVersion}</span>
                  </td>
                  <td className="p-3 text-xs text-gray-500">{d.appVersion}</td>
                  <td className="p-3 text-xs text-gray-500">{d.networkType || '-'}</td>
                  <td className="p-3 text-xs text-gray-500">{d.timezone || '-'}</td>
                  <td className="p-3 text-xs text-gray-400">{formatTime(d.updatedAt)}</td>
                  <td className="p-3">
                    <button onClick={() => handleDelete(d.deviceId)} className="text-gray-400 hover:text-red-500 transition-colors">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </td>
                </tr>
              ))}
              {devices.length === 0 && !loading && (
                <tr><td colSpan={7} className="p-8 text-center text-gray-400">暂无设备数据</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
