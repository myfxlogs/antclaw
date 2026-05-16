import { useEffect, useState } from 'react'
import { Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { createClient } from '@connectrpc/connect'
import { StreamService } from '@antclaw/proto/antclaw/v1/stream_pb'
import { transport } from '../lib/transport'
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

function structToObject(s: any): Record<string, unknown> {
  const obj: Record<string, unknown> = {}
  if (s?.fields) {
    for (const [k, v] of Object.entries(s.fields)) {
      obj[k] = (v as any).kind?.value ?? (v as any).stringValue ?? null
    }
  }
  return obj
}

const streamClient = createClient(StreamService, transport)

export default function Audit() {
  const { t } = useTranslation()
  const [entries, setEntries] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadAuditLogs()

    // Subscribe to Connect server-streaming (protobuf binary)
    const ac = new AbortController()
    ;(async () => {
      try {
        const stream = await streamClient.subscribeEvents({ channel: 'audit' }, { signal: ac.signal })
        for await (const event of stream) {
          if (event.type === 'system.heartbeat' || !event.payload) continue
          const data = structToObject(event.payload)
          setEntries((prev) => [
            {
              log_id: `stream-${Date.now()}`,
              user_id: String(data.user_id || ''),
              action: String(data.action),
              resource: String(data.resource),
              details: String(data.details),
              created_at: Number(data.timestamp),
              ip_address: String(data.ip_address || ''),
            },
            ...prev,
          ])
        }
      } catch {
        // stream closed
      }
    })()
    return () => ac.abort()
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

  const handleExport = () => {
    const json = JSON.stringify(entries, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit_${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64">{t('audit.loading')}</div>
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('audit.title')}</h1>
        <button
          onClick={handleExport}
          className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg text-sm font-medium text-gray-700"
          disabled={entries.length === 0}
        >
          <Download className="w-4 h-4" />
          {t('audit.export') || '导出'}
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.time')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.action')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.resource')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.details')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('audit.user')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">IP</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {entries.map((entry) => (
              <tr key={entry.log_id} className="hover:bg-gray-50">
                <td className="px-6 py-4 text-sm text-gray-500">
                  {new Date(entry.created_at * 1000).toLocaleString('zh-CN', { hour12: false })}
                </td>
                <td className="px-6 py-4 text-sm">
                  <span className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs font-medium">
                    {entry.action}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600 font-mono">{entry.resource}</td>
                <td className="px-6 py-4 text-sm text-gray-600 max-w-xs truncate">{entry.details}</td>
                <td className="px-6 py-4 text-sm text-gray-600 font-mono">{entry.user_id || '-'}</td>
                <td className="px-6 py-4 text-sm text-gray-500 font-mono">{entry.ip_address || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
