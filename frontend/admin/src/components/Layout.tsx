import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { LayoutDashboard, Users, Activity, ClipboardList, Database, KeyRound, Settings, LogOut, TrendingUp, Bot } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { logout } from '../lib/api'

const navItems = [
  { path: '/', key: 'dashboard', icon: LayoutDashboard },
  { path: '/users', key: 'users', icon: Users },
  { path: '/jobs', key: 'jobs', icon: Activity },
  { path: '/audit', key: 'audit', icon: ClipboardList },
  { path: '/data', key: 'data', icon: Database },
  { path: '/datasources', key: 'datasources', icon: KeyRound },
  { path: '/strategies', key: 'strategies', icon: TrendingUp },
  { path: '/system-ai', key: 'systemAI', icon: Bot },
  { path: '/settings', key: 'settings', icon: Settings },
]

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { t } = useTranslation()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-gray-100">
      {/* Sidebar */}
      <aside className="w-64 bg-white shadow-md flex flex-col">
        <div className="p-6">
          <h1 className="text-xl font-bold text-blue-600">AntClaw Admin</h1>
        </div>
        <nav className="px-4 space-y-1 flex-1">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                  isActive ? 'bg-blue-50 text-blue-600' : 'text-gray-600 hover:bg-gray-50'
                }`}
              >
                <Icon className="w-5 h-5" />
                <span className="font-medium">{t(`nav.${item.key}`)}</span>
              </Link>
            )
          })}
        </nav>
        <div className="p-4 border-t">
          <button
            onClick={handleLogout}
            className="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors text-gray-600 hover:bg-gray-50 w-full"
          >
            <LogOut className="w-5 h-5" />
            <span className="font-medium">{t('nav.logout')}</span>
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto p-8">
        <Outlet />
      </main>
    </div>
  )
}
