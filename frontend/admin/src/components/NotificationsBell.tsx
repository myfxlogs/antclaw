import { useEffect, useRef, useState } from 'react'
import { Bell, Check, CheckCheck, X } from 'lucide-react'
import {
  NotificationItem,
  getUnreadNotificationCount,
  listUnreadNotifications,
  listNotificationHistory,
  markAllNotificationsRead,
  markNotificationRead,
  openNotificationsSSE,
} from '../lib/api'

// NotificationsBell —— 顶栏铃铛 + 抽屉。
//
// 数据流：
//   1) 挂载时拉一次未读数 + 未读列表（保证刷新后能看到）
//   2) 通过 SSE 订阅 /sse/notifications，到达即在头部插入并 +1 未读数
//   3) 打开抽屉 → 切换"未读 / 历史"两个 tab
//   4) 单条标记已读：列表中移除并 -1；全部已读：清空未读区，未读数清零
//
// 此组件仅负责"展示与交互"；持久化、过滤、静默时段在后端完成。
export default function NotificationsBell() {
  const [open, setOpen] = useState(false)
  const [tab, setTab] = useState<'unread' | 'history'>('unread')
  const [unread, setUnread] = useState<NotificationItem[]>([])
  const [history, setHistory] = useState<NotificationItem[]>([])
  const [count, setCount] = useState<number>(0)
  const [loading, setLoading] = useState(false)
  const dropRef = useRef<HTMLDivElement | null>(null)

  // 初次拉一次未读
  useEffect(() => {
    let mounted = true
    ;(async () => {
      try {
        const [n, items] = await Promise.all([
          getUnreadNotificationCount(),
          listUnreadNotifications(50),
        ])
        if (!mounted) return
        setCount(n)
        setUnread(items)
      } catch (e) {
        // 401 等错误不打扰用户：保持 0
      }
    })()
    return () => {
      mounted = false
    }
  }, [])

  // 实时订阅
  useEffect(() => {
    const close = openNotificationsSSE((n) => {
      setUnread((prev) => [n, ...prev].slice(0, 100))
      setCount((c) => c + 1)
    })
    return close
  }, [])

  // 点击外部关闭
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (dropRef.current && !dropRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const onOpen = async () => {
    setOpen((v) => !v)
    if (!open && tab === 'history' && history.length === 0) {
      await loadHistory()
    }
  }

  const loadHistory = async () => {
    setLoading(true)
    try {
      setHistory(await listNotificationHistory(50))
    } finally {
      setLoading(false)
    }
  }

  const handleMarkOne = async (id: string) => {
    if (!id) return
    try {
      await markNotificationRead(id)
      setUnread((prev) => prev.filter((x) => x.id !== id))
      setCount((c) => Math.max(0, c - 1))
    } catch {}
  }

  const handleMarkAll = async () => {
    try {
      await markAllNotificationsRead()
      setUnread([])
      setCount(0)
    } catch {}
  }

  return (
    <div ref={dropRef} className="relative">
      <button
        onClick={onOpen}
        className="relative p-2 rounded-lg hover:bg-gray-100 transition-colors"
        title="通知"
      >
        <Bell className="w-5 h-5 text-gray-600" />
        {count > 0 && (
          <span className="absolute -top-0.5 -right-0.5 min-w-[18px] h-[18px] px-1 rounded-full bg-red-500 text-white text-[10px] font-semibold flex items-center justify-center">
            {count > 99 ? '99+' : count}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-96 bg-white rounded-xl shadow-lg border z-50 overflow-hidden">
          <div className="flex items-center justify-between px-4 py-3 border-b">
            <div className="flex gap-3">
              <button
                onClick={() => setTab('unread')}
                className={`text-sm font-medium ${tab === 'unread' ? 'text-blue-600' : 'text-gray-500'}`}
              >
                未读 {count > 0 ? `(${count})` : ''}
              </button>
              <button
                onClick={async () => {
                  setTab('history')
                  if (history.length === 0) await loadHistory()
                }}
                className={`text-sm font-medium ${tab === 'history' ? 'text-blue-600' : 'text-gray-500'}`}
              >
                历史
              </button>
            </div>
            <div className="flex gap-2">
              {tab === 'unread' && unread.length > 0 && (
                <button
                  onClick={handleMarkAll}
                  className="flex items-center gap-1 text-xs text-gray-500 hover:text-blue-600"
                  title="全部标记已读"
                >
                  <CheckCheck className="w-4 h-4" />
                  全部已读
                </button>
              )}
              <button onClick={() => setOpen(false)} className="text-gray-400 hover:text-gray-600">
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>
          <div className="max-h-96 overflow-y-auto">
            {tab === 'unread' ? (
              unread.length === 0 ? (
                <div className="p-6 text-center text-sm text-gray-400">暂无未读通知</div>
              ) : (
                unread.map((n, idx) => (
                  <NotifRow key={n.id || idx} n={n} onMark={() => handleMarkOne(n.id)} />
                ))
              )
            ) : loading ? (
              <div className="p-6 text-center text-sm text-gray-400">加载中…</div>
            ) : history.length === 0 ? (
              <div className="p-6 text-center text-sm text-gray-400">暂无记录</div>
            ) : (
              history.map((n) => <NotifRow key={n.id} n={n} readonly />)
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function severityClass(s: string): string {
  switch (s) {
    case 'critical':
      return 'bg-red-50 text-red-700 border-red-200'
    case 'high':
      return 'bg-orange-50 text-orange-700 border-orange-200'
    case 'low':
      return 'bg-gray-50 text-gray-600 border-gray-200'
    default:
      return 'bg-blue-50 text-blue-700 border-blue-200'
  }
}

function NotifRow({ n, onMark, readonly }: { n: NotificationItem; onMark?: () => void; readonly?: boolean }) {
  const ts = n.created_at ? new Date(n.created_at * 1000).toLocaleString() : ''
  return (
    <div className={`px-4 py-3 border-b last:border-b-0 ${n.is_read ? 'bg-white' : 'bg-blue-50/30'}`}>
      <div className="flex items-start gap-2">
        <span className={`text-[10px] px-1.5 py-0.5 rounded border ${severityClass(n.severity)}`}>
          {n.category}
        </span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-gray-900 truncate">{n.title}</p>
          <p className="text-xs text-gray-600 mt-0.5 line-clamp-2">{n.body}</p>
          <p className="text-[10px] text-gray-400 mt-1">{ts}</p>
        </div>
        {!readonly && onMark && n.id && (
          <button onClick={onMark} className="text-gray-400 hover:text-blue-600" title="标记已读">
            <Check className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  )
}
