import { useState, useEffect } from 'react'
import { Mail, Shield, Globe, Save } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export default function Settings() {
  const { t } = useTranslation()
  const [emailConfig, setEmailConfig] = useState({ smtp: '', port: '587', user: '' })
  const [securityConfig, setSecurityConfig] = useState({ mfa: true, rateLimit: true })
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    const savedEmail = localStorage.getItem('settings_email')
    const savedSecurity = localStorage.getItem('settings_security')
    if (savedEmail) {
      setEmailConfig(JSON.parse(savedEmail))
    }
    if (savedSecurity) {
      setSecurityConfig(JSON.parse(savedSecurity))
    }
  }, [])

  const handleSave = () => {
    localStorage.setItem('settings_email', JSON.stringify(emailConfig))
    localStorage.setItem('settings_security', JSON.stringify(securityConfig))
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  const handleClearCache = () => {
    localStorage.removeItem('token')
    alert('Cache cleared')
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">{t('settings.title')}</h1>

      <div className="grid gap-6 max-w-2xl">
        <section className="bg-white p-6 rounded-xl shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <Mail className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">{t('settings.notifications')}</h2>
          </div>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">SMTP Server</label>
                <input
                  type="text"
                  value={emailConfig.smtp}
                  onChange={(e) => setEmailConfig({...emailConfig, smtp: e.target.value})}
                  className="w-full px-4 py-2 border rounded-lg"
                  placeholder="smtp.example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Port</label>
                <input
                  type="text"
                  value={emailConfig.port}
                  onChange={(e) => setEmailConfig({...emailConfig, port: e.target.value})}
                  className="w-full px-4 py-2 border rounded-lg"
                />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Username</label>
              <input
                type="text"
                value={emailConfig.user}
                onChange={(e) => setEmailConfig({...emailConfig, user: e.target.value})}
                className="w-full px-4 py-2 border rounded-lg"
              />
            </div>
            <div className="flex gap-2 items-center">
              <button onClick={handleSave} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2">
                <Save className="w-4 h-4" />
                {t('settings.save')}
              </button>
              {saved && <span className="text-green-600 text-sm">Saved!</span>}
            </div>
          </div>
        </section>

        <section className="bg-white p-6 rounded-xl shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">{t('settings.security')}</h2>
          </div>
          <div className="space-y-3">
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                className="w-5 h-5 rounded"
                checked={securityConfig.mfa}
                onChange={(e) => setSecurityConfig({...securityConfig, mfa: e.target.checked})}
              />
              <span className="text-gray-700">要求管理员账号使用 MFA</span>
            </label>
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                className="w-5 h-5 rounded"
                checked={securityConfig.rateLimit}
                onChange={(e) => setSecurityConfig({...securityConfig, rateLimit: e.target.checked})}
              />
              <span className="text-gray-700">启用速率限制</span>
            </label>
          </div>
        </section>

        <section className="bg-white p-6 rounded-xl shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <Globe className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-semibold">{t('settings.general')}</h2>
          </div>
          <div className="space-y-3">
            <button onClick={handleClearCache} className="w-full px-4 py-2 border rounded-lg text-left hover:bg-gray-50">
              清除缓存
            </button>
            <button className="w-full px-4 py-2 border rounded-lg text-left hover:bg-gray-50">
              运行数据库迁移
            </button>
            <button className="w-full px-4 py-2 border border-red-200 text-red-600 rounded-lg text-left hover:bg-red-50">
              重启所有服务
            </button>
          </div>
        </section>
      </div>
    </div>
  )
}
