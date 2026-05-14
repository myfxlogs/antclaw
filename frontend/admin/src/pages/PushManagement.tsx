import { useEffect, useState } from 'react'
import { Send, History, Users, AlertTriangle, CheckCircle } from 'lucide-react'
import { getOnlineUsers, sendPush, getPushHistory } from '../lib/api'

interface OnlineUser {
  userId: string
  remoteAddr: string
  connectedAt: number
}

interface PushRecord {
  id: string
  title: string
  body: string
  severity: string
  targetCount: number
  sentCount: number
  adminUserId: string
  createdAt: number
}

export default function PushManagement() {
  const [onlineUsers, setOnlineUsers] = useState<OnlineUser[]>([])
  const [selectedUsers, setSelectedUsers] = useState<string[]>([])
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [severity, setSeverity] = useState('normal')
  const [sending, setSending] = useState(false)
  const [sendResult, setSendResult] = useState<{ ok: boolean; msg: string } | null>(null)
  const [history, setHistory] = useState<PushRecord[]>([])

  useEffect(() => {
    refreshUsers()
    refreshHistory()
    const timer = setInterval(refreshUsers, 15000)
    return () => clearInterval(timer)
  }, [])

  const refreshUsers = async () => {
    try {
      const data = await getOnlineUsers()
      setOnlineUsers(data.users)
    } catch { /* ok */ }
  }

  const refreshHistory = async () => {
    try {
      const { entries } = await getPushHistory({ pageSize: 20 })
      setHistory(entries)
    } catch { /* ok */ }
  }

  const toggleUser = (id: string) => {
    setSelectedUsers(prev =>
      prev.includes(id) ? prev.filter(u => u !== id) : [...prev, id]
    )
  }

  const selectAll = () => setSelectedUsers(onlineUsers.map(u => u.userId))
  const clearSelection = () => setSelectedUsers([])

  const handleSend = async () => {
    if (!title.trim()) return
    setSending(true)
    setSendResult(null)
    try {
      const res = await sendPush({
        title: title.trim(),
        body: body.trim(),
        severity,
        targetUserIds: selectedUsers,
        category: 'system',
      })
      setSendResult({ ok: true, msg: `推送成功：${res.sentCount}/${res.onlineCount} 个用户已收到` })
      setTitle(''); setBody(''); setSelectedUsers([]); setSeverity('normal')
      refreshHistory()
    } catch (e: any) {
      setSendResult({ ok: false, msg: e?.message || '推送失败' })
    } finally {
      setSending(false)
    }
  }

  const formatTime = (ts: number) => {
    const d = new Date(ts * 1000)
    return d.toLocaleString()
  }

  const sevColors: Record<string, string> = {
    low: 'bg-gray-100 text-gray-600',
    normal: 'bg-blue-100 text-blue-700',
    high: 'bg-orange-100 text-orange-700',
    critical: 'bg-red-100 text-red-700',
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Send className="w-6 h-6 text-gray-500" />
        <h1 className="text-2xl font-bold text-gray-900">推送管理</h1>
      </div>

      {/* 编辑区 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">编辑推送消息</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">标题 *</label>
            <input
              type="text"
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="输入推送标题..."
              className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">正文</label>
            <textarea
              value={body}
              onChange={e => setBody(e.target.value)}
              placeholder="输入推送正文内容..."
              rows={4}
              className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none resize-y"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">严重级别</label>
            <select
              value={severity}
              onChange={e => setSeverity(e.target.value)}
              className="border border-gray-300 rounded-lg px-4 py-2 text-sm"
            >
              <option value="low">低</option>
              <option value="normal">普通</option>
              <option value="high">高</option>
              <option value="critical">紧急</option>
            </select>
          </div>
        </div>
      </div>

      {/* 目标用户选择 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-gray-400" />
            <h2 className="text-lg font-semibold text-gray-800">
              目标用户（{selectedUsers.length > 0 ? selectedUsers.length : '全部在线'}）
            </h2>
          </div>
          <div className="flex gap-2">
            <button onClick={selectAll} className="text-xs text-blue-600 hover:underline">全选</button>
            <button onClick={clearSelection} className="text-xs text-gray-500 hover:underline">清除</button>
          </div>
        </div>
        {onlineUsers.length === 0 ? (
          <div className="text-gray-400 text-sm">暂无在线用户（将推送给全部在线用户）</div>
        ) : (
          <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto">
            {onlineUsers.map(u => (
              <button
                key={u.userId}
                onClick={() => toggleUser(u.userId)}
                className={`px-3 py-1.5 rounded-full text-xs transition-colors ${
                  selectedUsers.length === 0 || selectedUsers.includes(u.userId)
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-500 hover:bg-gray-200'
                }`}
              >
                {u.userId.substring(0, 8)}...
              </button>
            ))}
          </div>
        )}
        <div className="text-xs text-gray-400 mt-2">
          不选择任何用户 = 推送给全部在线用户
        </div>
      </div>

      {/* 发送按钮 + 结果 */}
      <div className="flex items-center gap-4">
        <button
          onClick={handleSend}
          disabled={sending || !title.trim()}
          className={`flex items-center gap-2 px-6 py-3 rounded-lg text-sm font-medium transition-colors ${
            sending || !title.trim()
              ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
              : 'bg-blue-500 text-white hover:bg-blue-600'
          }`}
        >
          <Send className="w-4 h-4" />
          {sending ? '发送中...' : '发送推送'}
        </button>
        {sendResult && (
          <div className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm ${
            sendResult.ok ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
          }`}>
            {sendResult.ok ? <CheckCircle className="w-4 h-4" /> : <AlertTriangle className="w-4 h-4" />}
            {sendResult.msg}
          </div>
        )}
      </div>

      {/* 推送历史 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center gap-2 mb-4">
          <History className="w-5 h-5 text-gray-400" />
          <h2 className="text-lg font-semibold text-gray-800">推送历史</h2>
          <button onClick={refreshHistory} className="ml-auto text-xs text-blue-600 hover:underline">刷新</button>
        </div>
        {history.length === 0 ? (
          <div className="text-gray-400 text-sm text-center py-8">暂无推送记录</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100 text-left text-gray-500">
                  <th className="pb-2 font-medium w-8">#</th>
                  <th className="pb-2 font-medium">标题</th>
                  <th className="pb-2 font-medium">级别</th>
                  <th className="pb-2 font-medium">目标/发送</th>
                  <th className="pb-2 font-medium">时间</th>
                </tr>
              </thead>
              <tbody>
                {history.map((r, i) => (
                  <tr key={r.id} className="border-b border-gray-50 hover:bg-gray-50">
                    <td className="py-2 text-gray-400">{history.length - i}</td>
                    <td className="py-2">
                      <div className="font-medium text-gray-800 max-w-xs truncate">{r.title}</div>
                      <div className="text-xs text-gray-400 max-w-xs truncate">{r.body}</div>
                    </td>
                    <td className="py-2">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${sevColors[r.severity] || ''}`}>
                        {r.severity}
                      </span>
                    </td>
                    <td className="py-2 text-gray-600">
                      {r.sentCount}/{r.targetCount}
                    </td>
                    <td className="py-2 text-gray-400 text-xs">{formatTime(r.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
