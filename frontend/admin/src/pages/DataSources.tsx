import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { KeyRound, Save, RefreshCw, Trash2, Lock, Unlock } from 'lucide-react'
import { listDataSources, updateDataSource } from '../lib/api'

interface DataSourceConfig {
  source_id: string
  name: string
  kind: string
  endpoint: string
  has_secret: boolean
  updated_at: string
  updated_by: string
}

export default function DataSources() {
  const { t } = useTranslation()
  const [items, setItems] = useState<DataSourceConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [editing, setEditing] = useState<Record<string, { endpoint?: string; secret?: string }>>({})
  const [saving, setSaving] = useState<string | null>(null)
  const [message, setMessage] = useState<{ kind: 'ok' | 'err'; text: string } | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const data = await listDataSources()
      setItems(data.items || [])
    } catch (e: any) {
      setMessage({ kind: 'err', text: t('datasources.loadError', { message: e.message }) })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const updateEdit = (id: string, patch: { endpoint?: string; secret?: string }) => {
    setEditing((prev) => ({ ...prev, [id]: { ...prev[id], ...patch } }))
  }

  const submit = async (item: DataSourceConfig, action: 'save' | 'clear') => {
    setSaving(item.source_id)
    setMessage(null)
    try {
      const draft = editing[item.source_id] || {}
      const payload: Record<string, unknown> = {}
      if (action === 'save') {
        if (typeof draft.endpoint === 'string' && draft.endpoint !== item.endpoint) {
          payload.endpoint = draft.endpoint
        }
        if (typeof draft.secret === 'string' && draft.secret.length > 0) {
          payload.secret = draft.secret
        }
        if (Object.keys(payload).length === 0) {
          setMessage({ kind: 'err', text: t('datasources.noChanges') })
          return
        }
      } else {
        payload.clear_secret = true
      }

      await updateDataSource(item.source_id, {
        endpoint: payload.endpoint as string | undefined,
        secret: payload.secret as string | undefined,
        clear_secret: payload.clear_secret as boolean | undefined,
      })
      setMessage({ kind: 'ok', text: t('datasources.updated', { name: item.name }) })
      // 清空 secret 输入避免明文残留
      setEditing((prev) => ({ ...prev, [item.source_id]: { ...prev[item.source_id], secret: '' } }))
      await load()
    } catch (e: any) {
      setMessage({ kind: 'err', text: t('datasources.saveError', { message: e.message }) })
    } finally {
      setSaving(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
          <KeyRound className="w-6 h-6 text-blue-600" /> 数据源密钥与端点
        </h1>
        <button
          onClick={load}
          className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} /> 刷新
        </button>
      </div>

      <p className="text-sm text-gray-500">
        密钥通过 Connect-RPC 提交，后端仍使用 Argon2id + AES-256-GCM 加密落库；
        列表接口永不返回明文密钥。
      </p>

      {message && (
        <div
          className={`px-4 py-3 rounded-lg ${
            message.kind === 'ok' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">数据源</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">类型</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">Endpoint</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">密钥</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {items.map((it) => {
              const draft = editing[it.source_id] || {}
              const endpointVal = draft.endpoint ?? it.endpoint
              const secretVal = draft.secret ?? ''
              return (
                <tr key={it.source_id} className="hover:bg-gray-50 align-top">
                  <td className="px-6 py-4">
                    <div className="font-medium">{it.name}</div>
                    <div className="text-xs text-gray-400">{it.source_id}</div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-700">{it.kind}</td>
                  <td className="px-6 py-4">
                    <input
                      type="text"
                      value={endpointVal}
                      onChange={(e) => updateEdit(it.source_id, { endpoint: e.target.value })}
                      className="w-72 px-3 py-2 border rounded-lg text-sm font-mono"
                    />
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      {it.has_secret ? (
                        <Lock className="w-4 h-4 text-green-600" />
                      ) : (
                        <Unlock className="w-4 h-4 text-gray-400" />
                      )}
                      <input
                        type="password"
                        value={secretVal}
                        placeholder={it.has_secret ? '已配置（输入新值替换）' : '未配置'}
                        onChange={(e) => updateEdit(it.source_id, { secret: e.target.value })}
                        className="w-64 px-3 py-2 border rounded-lg text-sm"
                      />
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      <button
                        disabled={saving === it.source_id}
                        onClick={() => submit(it, 'save')}
                        className="px-3 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700 flex items-center gap-1 disabled:opacity-50"
                      >
                        <Save className="w-4 h-4" /> 保存
                      </button>
                      {it.has_secret && (
                        <button
                          disabled={saving === it.source_id}
                          onClick={() => submit(it, 'clear')}
                          className="px-3 py-2 bg-red-50 text-red-600 rounded-lg text-sm hover:bg-red-100 flex items-center gap-1 disabled:opacity-50"
                        >
                          <Trash2 className="w-4 h-4" /> 清除密钥
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
