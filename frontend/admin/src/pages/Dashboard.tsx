import { useState } from 'react'
import { useTranslation } from 'react-i18next'

export default function Dashboard() {
  const { t } = useTranslation()
  const [loading] = useState(false)

  if (loading) {
    return <div className="flex items-center justify-center h-64">加载中...</div>
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">{t('dashboard.title')}</h1>
      <div className="bg-white p-12 rounded-xl shadow-sm text-center">
        <p className="text-gray-500">系统运行正常</p>
      </div>
    </div>
  )
}
