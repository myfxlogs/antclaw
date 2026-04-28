import { useState } from 'react'
import { Key, Globe, Bell, Shield } from 'lucide-react'

export default function Settings() {
  const [aiKey, setAiKey] = useState('')
  const [locale, setLocale] = useState('zh-CN')
  const [timezone, setTimezone] = useState('Asia/Shanghai')
  const [notifications, setNotifications] = useState(true)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Settings</h1>

      <div className="grid gap-6 max-w-2xl">
        {/* AI Key Section */}
        <section className="bg-white p-6 rounded-xl shadow-sm border">
          <div className="flex items-center gap-3 mb-4">
            <Key className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">AI API Key (BYOK)</h2>
          </div>
          <p className="text-sm text-gray-600 mb-4">
            Bring your own API key for AI features. Your key is encrypted and never shared.
          </p>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Provider</label>
              <select className="w-full px-4 py-2 border rounded-lg">
                <option value="gemini">Google Gemini</option>
                <option value="claude">Anthropic Claude</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">API Key</label>
              <input
                type="text"
                value={aiKey}
                onChange={(e) => setAiKey(e.target.value)}
                className="w-full px-4 py-2 border rounded-lg"
                placeholder="Enter your API key"
              />
            </div>
            <button className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
              Save Key
            </button>
          </div>
        </section>

        {/* Locale & Timezone */}
        <section className="bg-white p-6 rounded-xl shadow-sm border">
          <div className="flex items-center gap-3 mb-4">
            <Globe className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">Region & Language</h2>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Language</label>
              <select 
                value={locale} 
                onChange={(e) => setLocale(e.target.value)}
                className="w-full px-4 py-2 border rounded-lg"
              >
                <option value="zh-CN">简体中文</option>
                <option value="en-US">English</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Timezone</label>
              <select 
                value={timezone} 
                onChange={(e) => setTimezone(e.target.value)}
                className="w-full px-4 py-2 border rounded-lg"
              >
                <option value="Asia/Shanghai">Asia/Shanghai</option>
                <option value="America/New_York">America/New_York</option>
                <option value="Europe/London">Europe/London</option>
              </select>
            </div>
          </div>
        </section>

        {/* Notifications */}
        <section className="bg-white p-6 rounded-xl shadow-sm border">
          <div className="flex items-center gap-3 mb-4">
            <Bell className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">Notifications</h2>
          </div>
          <label className="flex items-center gap-3">
            <input 
              type="checkbox" 
              checked={notifications}
              onChange={(e) => setNotifications(e.target.checked)}
              className="w-5 h-5 rounded"
            />
            <span className="text-gray-700">Enable push notifications for new signals</span>
          </label>
        </section>

        {/* Security */}
        <section className="bg-white p-6 rounded-xl shadow-sm border">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">Security</h2>
          </div>
          <div className="space-y-3">
            <button className="w-full px-4 py-2 border rounded-lg text-sm hover:bg-gray-50 text-left">
              Change Password
            </button>
            <button className="w-full px-4 py-2 border rounded-lg text-sm hover:bg-gray-50 text-left">
              View Active Sessions
            </button>
            <button className="w-full px-4 py-2 border border-red-200 text-red-600 rounded-lg text-sm hover:bg-red-50 text-left">
              Delete Account
            </button>
          </div>
        </section>
      </div>
    </div>
  )
}
