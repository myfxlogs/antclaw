import { useState } from 'react'
import { Send, History, Users, AlertTriangle, CheckCircle, Eye, TestTube } from 'lucide-react'
import { getOnlineUsers, sendPush, getPushHistory } from '../lib/api'
import { usePolling } from '../hooks/usePolling'
import DangerConfirmDialog from '../components/DangerConfirmDialog'
import { Badge, ErrorState, EmptyState, LoadingSkeleton } from '../components/Common'

export default function PushManagement() {
  const [selectedUsers, setSelectedUsers] = useState<string[]>([])
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [severity, setSeverity] = useState('normal')
  const [sending, setSending] = useState(false)
  const [sendResult, setSendResult] = useState<{ ok: boolean; msg: string } | null>(null)
  const [showConfirm, setShowConfirm] = useState(false)

  // Online users polling with backoff + pause-when-hidden
  const { data: onlineData, loading: usersLoading, error: usersError, lastUpdated } = usePolling({
    fetcher: async () => {
      const data = await getOnlineUsers()
      return data.users
    },
    intervalMs: 15000,
    pauseWhenHidden: true,
  })

  // Push history polling
  const { data: history, loading: historyLoading } = usePolling({
    fetcher: async () => {
      const { entries } = await getPushHistory({ pageSize: 20 })
      return entries
    },
    intervalMs: 30000,
    pauseWhenHidden: true,
  })

  const onlineUsers = onlineData || []
  const pushHistory = history || []
  const targetLabel = selectedUsers.length > 0 ? `${selectedUsers.length} 个用户` : '全部在线用户'
  const isAllUsers = selectedUsers.length === 0

  const toggleUser = (id: string) => {
    setSelectedUsers(prev => prev.includes(id) ? prev.filter(u => u !== id) : [...prev, id])
  }

  const selectAll = () => setSelectedUsers(onlineUsers.map(u => u.userId))
  const clearSelection = () => setSelectedUsers([])

  const handleSend = async (_reason: string) => {
    setSending(true)
    setSendResult(null)
    try {
      const res = await sendPush({ title: title.trim(), body: body.trim(), severity, targetUserIds: selectedUsers, category: 'system' })
      setSendResult({ ok: true, msg: `推送成功：${res.sentCount}/${res.onlineCount} 个用户已收到` })
      setTitle(''); setBody(''); setSelectedUsers([]); setSeverity('normal')
    } catch (e: any) {
      setSendResult({ ok: false, msg: e?.message || '推送失败' })
    } finally {
      setSending(false)
      setShowConfirm(false)
    }
  }

  const handleTestSend = async () => {
    if (!title.trim()) return
    setSending(true)
    setSendResult(null)
    try {
      const res = await sendPush({ title: `[测试] ${title.trim()}`, body: body.trim(), severity, targetUserIds: [], category: 'system' })
      setSendResult({ ok: true, msg: `测试发送成功：${res.sentCount} 个用户收到` })
    } catch (e: any) {
      setSendResult({ ok: false, msg: e?.message || '测试发送失败' })
    } finally { setSending(false) }
  }

  const sevColors: Record<string, string> = {
    low: 'bg-gray-100 text-gray-600', normal: 'bg-blue-100 text-blue-700',
    high: 'bg-orange-100 text-orange-700', critical: 'bg-red-100 text-red-700',
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Send className="w-6 h-6 text-gray-500" />
        <h1 className="text-2xl font-bold text-gray-900">推送管理</h1>
        {lastUpdated && <span className="text-xs text-gray-400 ml-auto">最后更新: {lastUpdated.toLocaleTimeString()}</span>}
      </div>

      {/* 编辑区 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">编辑推送消息</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">标题 *</label>
            <input type="text" value={title} onChange={e => setTitle(e.target.value)} placeholder="输入推送标题..." className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">正文</label>
            <textarea value={body} onChange={e => setBody(e.target.value)} placeholder="输入推送正文内容..." rows={4} className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-y" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">严重级别</label>
            <select value={severity} onChange={e => setSeverity(e.target.value)} className="border border-gray-300 rounded-lg px-4 py-2 text-sm">
              <option value="low">低</option><option value="normal">普通</option><option value="high">高</option><option value="critical">紧急</option>
            </select>
          </div>
        </div>

        {/* 预览卡片 */}
        {(title || body) && (
          <div className="mt-4 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <div className="flex items-center gap-2 mb-2"><Eye className="w-4 h-4 text-gray-400" /><span className="text-xs text-gray-500 font-medium">推送预览</span></div>
            <div className="space-y-1">
              <div className="font-medium text-gray-900">{title || '(无标题)'}</div>
              {body && <div className="text-sm text-gray-600">{body}</div>}
              <Badge variant={severity === 'critical' ? 'danger' : severity === 'high' ? 'warning' : 'info'}>{severity}</Badge>
            </div>
            <div className="mt-2 text-xs text-gray-400">目标：{targetLabel}</div>
          </div>
        )}
      </div>

      {/* 目标用户 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Users className="w-5 h-5 text-gray-400" />
            <h2 className="text-lg font-semibold text-gray-800">目标用户（{targetLabel}）</h2>
          </div>
          <div className="flex gap-2">
            <button onClick={selectAll} className="text-xs text-blue-600 hover:underline">全选</button>
            <button onClick={clearSelection} className="text-xs text-gray-500 hover:underline">清除</button>
          </div>
        </div>
        {usersLoading ? <LoadingSkeleton rows={3} /> :
         usersError ? <ErrorState message={usersError} /> :
         onlineUsers.length === 0 ? <EmptyState title="暂无在线用户" description="将推送给全部在线用户" /> :
         <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto">
          {onlineUsers.map(u => (
            <button key={u.userId} onClick={() => toggleUser(u.userId)}
              className={`px-3 py-1.5 rounded-full text-xs transition-colors ${selectedUsers.length === 0 || selectedUsers.includes(u.userId) ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'}`}>
              {u.displayName || u.userId.substring(0, 8)}...
            </button>
          ))}
        </div>}

        {/* 全员推送警告 */}
        {isAllUsers && !usersLoading && onlineUsers.length > 0 && (
          <div className="mt-3 flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg text-sm text-yellow-800">
            <AlertTriangle className="w-4 h-4 flex-shrink-0" />
            此推送将发送给全部 {onlineUsers.length} 个在线用户，请确认内容后再发送
          </div>
        )}
      </div>

      {/* 发送按钮 */}
      <div className="flex items-center gap-4">
        <button onClick={() => setShowConfirm(true)} disabled={sending || !title.trim()}
          className={`flex items-center gap-2 px-6 py-3 rounded-lg text-sm font-medium transition-colors ${sending || !title.trim() ? 'bg-gray-200 text-gray-400 cursor-not-allowed' : 'bg-blue-500 text-white hover:bg-blue-600'}`}>
          <Send className="w-4 h-4" />{sending ? '发送中...' : '发送推送'}
        </button>
        <button onClick={handleTestSend} disabled={sending || !title.trim()}
          className="flex items-center gap-2 px-4 py-3 rounded-lg text-sm border border-blue-200 text-blue-600 hover:bg-blue-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
          <TestTube className="w-4 h-4" />测试发送
        </button>
        {sendResult && (
          <div className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm ${sendResult.ok ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
            {sendResult.ok ? <CheckCircle className="w-4 h-4" /> : <AlertTriangle className="w-4 h-4" />}{sendResult.msg}
          </div>
        )}
      </div>

      {/* 发送确认对话框 */}
      <DangerConfirmDialog
        open={showConfirm}
        onClose={() => setShowConfirm(false)}
        onConfirm={handleSend}
        title="确认发送推送"
        description={isAllUsers ? `即将向全部 ${onlineUsers.length} 个在线用户发送推送` : `即将向 ${selectedUsers.length} 个选定用户发送推送`}
        targetName={title || '(无标题)'}
        confirmLabel="确认发送"
        requireReason={false}
      />

      {/* 推送历史 */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div className="flex items-center gap-2 mb-4">
          <History className="w-5 h-5 text-gray-400" />
          <h2 className="text-lg font-semibold text-gray-800">推送历史</h2>
        </div>
        {historyLoading ? <LoadingSkeleton rows={3} /> :
         pushHistory.length === 0 ? <EmptyState title="暂无推送记录" /> :
         <div className="space-y-2">
          {pushHistory.map(r => (
            <div key={r.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg text-sm">
              <div>
                <div className="font-medium text-gray-800">{r.title}</div>
                <div className="text-xs text-gray-500">{r.body.substring(0, 80)}{r.body.length > 80 ? '...' : ''}</div>
              </div>
              <div className="flex items-center gap-3 text-xs text-gray-500">
                <Badge>{r.sentCount}/{r.targetCount}</Badge>
                <span>{new Date(r.createdAt * 1000).toLocaleString()}</span>
                <span className={sevColors[r.severity] || ''}>{r.severity}</span>
              </div>
            </div>
          ))}
        </div>}
      </div>
    </div>
  )
}
