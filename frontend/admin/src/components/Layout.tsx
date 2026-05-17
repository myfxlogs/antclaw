import { useEffect, useState } from 'react'
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import {
  LayoutDashboard, Users, Activity, ClipboardList, Database, KeyRound,
  Settings, LogOut, TrendingUp, Bot, BarChart3, LineChart, Globe2, Send,
  Layers, Shield, MessageCircle, Wand2, ChevronDown, ChevronRight, Smartphone,
  Users2,
} from 'lucide-react'
import { useAuth } from '../features/auth/AuthProvider'
import { Permissions } from '../features/auth/permissions'
import NotificationsBell from './NotificationsBell'

interface NavItem { path: string; label: string; icon: any; permission?: string }
interface NavGroup { id: string; label: string; collapsible?: boolean; permission?: string; items: NavItem[] }

const groups: NavGroup[] = [
  {
    id: 'overview',
    label: '概览',
    items: [
      { path: '/', label: '仪表盘', icon: LayoutDashboard },
      { path: '/users', label: '用户', icon: Users, permission: Permissions.USERS_READ },
      { path: '/jobs', label: '作业', icon: Activity },
      { path: '/audit', label: '审计', icon: ClipboardList, permission: Permissions.AUDIT_READ },
      { path: '/data', label: '数据汇总', icon: Database },
      { path: '/datasources', label: '数据源', icon: KeyRound },
      { path: '/strategies', label: '策略', icon: TrendingUp },
      { path: '/system-ai', label: 'AI 配置', icon: Bot, permission: Permissions.AI_MANAGE },
      { path: '/devices', label: '设备管理', icon: Smartphone },
      { path: '/push', label: '推送管理', icon: Send, permission: Permissions.PUSH_SEND },
      { path: '/social', label: '社交管理', icon: Users2, permission: Permissions.SOCIAL_READ },
      { path: '/settings', label: '设置', icon: Settings },
    ],
  },
  {
    id: 'options',
    label: '期权与波动率',
    collapsible: true,
    items: [
      { path: '/options/gex', label: 'GEX', icon: BarChart3 },
      { path: '/options/iv-surface', label: 'IV Surface', icon: BarChart3 },
      { path: '/options/skew', label: 'Skew', icon: BarChart3 },
      { path: '/options/alerts', label: '期权告警', icon: BarChart3 },
      { path: '/vol/move', label: 'MOVE', icon: LineChart },
      { path: '/vol/cross', label: '跨市场波动', icon: LineChart },
      { path: '/vol/term', label: '期限结构', icon: LineChart },
    ],
  },
  {
    id: 'backtest',
    label: '回测与信号',
    collapsible: true,
    items: [
      { path: '/backtest/walkforward', label: 'Walk-Forward', icon: Layers },
      { path: '/signals/regime', label: 'Regime Overlay', icon: Layers },
      { path: '/ta/amt', label: 'AMT', icon: Layers },
      { path: '/microstructure/vp', label: '体积区间', icon: Layers },
    ],
  },
  {
    id: 'macro',
    label: '宏观',
    collapsible: true,
    items: [
      { path: '/macro/fedwatch', label: 'FedWatch', icon: Globe2 },
      { path: '/macro/extras', label: '宏观全谱', icon: Globe2 },
      { path: '/macro/fred-alerts', label: 'FRED 告警', icon: Globe2 },
      { path: '/macro/treasury', label: '美债曲线', icon: Globe2 },
    ],
  },
  {
    id: 'sentiment',
    label: '链上 / SEC / 情绪',
    collapsible: true,
    items: [
      { path: '/onchain', label: '链上分析', icon: Shield },
      { path: '/defi', label: 'DeFi 协议榜', icon: Shield },
      { path: '/sec', label: 'SEC EDGAR', icon: Shield },
      { path: '/sentiment/cboe-pc', label: 'CBOE Put/Call', icon: Shield },
      { path: '/sentiment/myfxbook', label: 'MyFXBook', icon: Shield },
      { path: '/sentiment/insider', label: '内部人交易', icon: Shield },
      { path: '/sentiment/finviz', label: 'Finviz', icon: Shield },
      { path: '/sentiment/crypto-social', label: '加密社交', icon: Shield },
    ],
  },
  {
    id: 'ai',
    label: 'AI',
    collapsible: true,
    items: [
      { path: '/ai/chat', label: 'AI 对话', icon: MessageCircle },
    ],
  },
]

const COLLAPSE_STORAGE_KEY = 'antclaw.admin.sidebar.collapsed'

// 默认收拢这 4 个二级分组（用户偏好）。
const DEFAULT_COLLAPSED: Record<string, boolean> = {
  options: true,
  backtest: true,
  sentiment: true,
  ai: true,
}

function loadCollapsedState(): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(COLLAPSE_STORAGE_KEY)
    if (raw) return { ...DEFAULT_COLLAPSED, ...JSON.parse(raw) }
  } catch {}
  return { ...DEFAULT_COLLAPSED }
}

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { logout, hasPermission } = useAuth()
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>(() => loadCollapsedState())

  // 当用户访问的路由属于某个折叠组时，自动展开该组，避免活跃项被隐藏。
  useEffect(() => {
    const active = groups.find((g) => g.items.some((it) => it.path === location.pathname))
    if (active && active.collapsible && collapsed[active.id]) {
      setCollapsed((prev) => ({ ...prev, [active.id]: false }))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname])

  useEffect(() => {
    try {
      localStorage.setItem(COLLAPSE_STORAGE_KEY, JSON.stringify(collapsed))
    } catch {}
  }, [collapsed])

  const toggleGroup = (id: string) =>
    setCollapsed((prev) => ({ ...prev, [id]: !prev[id] }))

  const handleLogout = () => {
    logout() // AuthProvider.logout() — clears memory state
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-gray-100">
      <aside className="w-60 bg-white shadow-md flex flex-col">
        <div className="p-5 border-b">
          <h1 className="text-xl font-bold text-blue-600 flex items-center gap-2">
            <Wand2 className="w-5 h-5" /> AntClaw Admin
          </h1>
        </div>
        <nav className="px-2 py-3 space-y-3 flex-1 overflow-y-auto">
          {groups.map((g) => {
            // Permission filtering (A13-P0-01)
            const visibleItems = g.items.filter(
              (it) => !it.permission || hasPermission(it.permission),
            )
            if (visibleItems.length === 0) return null

            const isCollapsed = !!(g.collapsible && collapsed[g.id])
            const header = g.collapsible ? (
              <button
                type="button"
                onClick={() => toggleGroup(g.id)}
                className="w-full flex items-center justify-between px-3 mb-1 text-xs font-medium text-gray-400 uppercase tracking-wide hover:text-gray-600 transition-colors"
              >
                <span>{g.label}</span>
                {isCollapsed ? (
                  <ChevronRight className="w-3.5 h-3.5" />
                ) : (
                  <ChevronDown className="w-3.5 h-3.5" />
                )}
              </button>
            ) : (
              <div className="px-3 mb-1 text-xs font-medium text-gray-400 uppercase tracking-wide">{g.label}</div>
            )
            return (
              <div key={g.id}>
                {header}
                {!isCollapsed && (
                  <div className="space-y-0.5">
                    {visibleItems.map((it) => {
                      const Icon = it.icon
                      const isActive = location.pathname === it.path
                      return (
                        <Link
                          key={it.path}
                          to={it.path}
                          className={`flex items-center gap-2 px-3 py-1.5 rounded text-sm transition-colors ${
                            isActive ? 'bg-blue-50 text-blue-600' : 'text-gray-600 hover:bg-gray-50'
                          }`}
                        >
                          <Icon className="w-4 h-4" />
                          <span>{it.label}</span>
                        </Link>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </nav>
        <div className="p-3 border-t">
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-3 py-2 rounded text-sm text-gray-600 hover:bg-gray-50 w-full"
          >
            <LogOut className="w-4 h-4" />
            <span>登出</span>
          </button>
        </div>
      </aside>
      <main className="flex-1 overflow-auto">
        <div className="sticky top-0 z-10 bg-gray-100/90 backdrop-blur px-8 py-3 flex justify-end border-b border-gray-200">
          <NotificationsBell />
        </div>
        <div className="p-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
