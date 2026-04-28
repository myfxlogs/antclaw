import { useEffect, useState } from 'react'
import { FileText, Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { listAuditLogs } from '../lib/api'

interface AuditEntry {
  log_id: string
  user_id: string
  action: string
  resource: string
  details: string
  created_at: number
  ip_address: string
}

export default function Audit() {
  const { t } = useTranslation()
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadAuditLogs()

    // 订阅审计日志 SSE，实现实时追加
    const evtSource = new EventSource('/sse/audit')
    evtSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as {
          user_id?: string
          action: string
          resource: string
          details: string
          ip_address?: string
          timestamp: number
        }
        setEntries((prev) => [
          {
            log_id: `sse-${Date.now()}`,
            user_id: data.user_id || '',
            action: data.action,
            resource: data.resource,
            details: data.details,
            created_at: data.timestamp,
            ip_address: data.ip_address || '',
          },
          ...prev,
        ])
      } catch {
        // ignore malformed events
      }
    }
    evtSource.onerror = () => {
      evtSource.close()
    }
    return () => {
      evtSource.close()
    }
  }, [])

  const loadAuditLogs = async () => {
    try {
      const response = await listAuditLogs()
      setEntries(response.logs)
    } catch (err) {
      console.error('Failed to load audit logs:', err)
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (timestamp: number) => {
    return new Date(timestamp * 1000).toLocaleString('zh-CN')
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64">{t('audit.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('audit.title')}</h1>
        <button className="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-gray-50">
          <Download className="w-4 h-4" />
          导出
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.action')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.userId')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.resource')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.createdAt')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.ipAddress')}</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {entries.map((entry) => (
              <tr key={entry.log_id} className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <span className="inline-flex items-center gap-2">
                    <FileText className="w-4 h-4 text-gray-400" />
                    <span className="font-medium">{entry.action}</span>
                  </span>
                </td>
                <td className="px-6 py-4 text-sm">{entry.user_id}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{entry.resource}</td>
                <td className="px-6 py-4 text-sm text-gray-500">{formatDate(entry.created_at)}</td>
                <td className="px-6 py-4 text-sm text-gray-500">{entry.ip_address}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
