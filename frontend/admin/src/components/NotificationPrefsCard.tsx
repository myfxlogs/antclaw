import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bell, Save, Loader2 } from 'lucide-react'
import {
  NotificationPrefsItem,
  getNotificationPrefs,
  updateNotificationPrefs,
} from '../lib/api'

// NotificationPrefsCard —— 用户通知偏好。
//
// 落地的用户控制项：
//   - 类别白名单（alert / signal / system / digest）
//   - 最低严重度（低于此级别的"实时推送"会被静默；历史记录依然可见）
//   - 静默时段（quiet_start == quiet_end 表示不静默；critical 永远穿透）
//   - 实时推送开关
const ALL_CATEGORIES: { key: string; label: string }[] = [
  { key: 'alert', label: '告警' },
  { key: 'signal', label: '信号' },
  { key: 'system', label: '系统' },
  { key: 'digest', label: '摘要 / 早报' },
]

const SEVERITIES: { key: string; label: string }[] = [
  { key: 'low', label: '低（全部）' },
  { key: 'normal', label: '普通及以上' },
  { key: 'high', label: '仅高 / 严重' },
  { key: 'critical', label: '仅严重' },
]

const HHMM = /^([01]\d|2[0-3]):[0-5]\d$/

export default function NotificationPrefsCard() {
  const { t } = useTranslation()
  const [prefs, setPrefs] = useState<NotificationPrefsItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState<{ kind: 'ok' | 'err'; text: string } | null>(null)

  useEffect(() => {
    ;(async () => {
      try {
        setPrefs(await getNotificationPrefs())
      } catch (e: any) {
        setMsg({ kind: 'err', text: t('notifications.loadPrefsError', { message: e?.message || String(e) }) })
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  if (loading) {
    return (
      <section className="bg-white p-6 rounded-xl shadow-sm">
        <Loader2 className="w-5 h-5 animate-spin text-gray-400" />
      </section>
    )
  }
  if (!prefs) return null

  const toggleType = (key: string) => {
    const has = prefs.enabled_types.includes(key)
    setPrefs({
      ...prefs,
      enabled_types: has
        ? prefs.enabled_types.filter((x) => x !== key)
        : [...prefs.enabled_types, key],
    })
  }

  const onSave = async () => {
    setMsg(null)
    if (!HHMM.test(prefs.quiet_start) || !HHMM.test(prefs.quiet_end)) {
      setMsg({ kind: 'err', text: t('notifications.quietTimeFormatError') })
      return
    }
    setSaving(true)
    try {
      const out = await updateNotificationPrefs(prefs)
      setPrefs(out)
      setMsg({ kind: 'ok', text: t('notifications.saved') })
    } catch (e: any) {
      setMsg({ kind: 'err', text: t('notifications.saveError', { message: e?.message || String(e) }) })
    } finally {
      setSaving(false)
      setTimeout(() => setMsg(null), 2500)
    }
  }

  return (
    <section className="bg-white p-6 rounded-xl shadow-sm">
      <div className="flex items-center gap-3 mb-4">
        <Bell className="w-5 h-5 text-blue-600" />
        <h2 className="text-lg font-semibold">通知偏好</h2>
      </div>

      <div className="space-y-5">
        <div>
          <p className="text-sm font-medium text-gray-700 mb-2">接收的通知类别</p>
          <div className="flex flex-wrap gap-2">
            {ALL_CATEGORIES.map((c) => {
              const on = prefs.enabled_types.includes(c.key)
              return (
                <button
                  key={c.key}
                  type="button"
                  onClick={() => toggleType(c.key)}
                  className={`px-3 py-1.5 text-sm rounded-full border transition-colors ${
                    on
                      ? 'bg-blue-50 border-blue-300 text-blue-700'
                      : 'bg-white border-gray-200 text-gray-500 hover:border-gray-300'
                  }`}
                >
                  {c.label}
                </button>
              )
            })}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">最低推送等级</label>
            <select
              value={prefs.min_severity}
              onChange={(e) => setPrefs({ ...prefs, min_severity: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg"
            >
              {SEVERITIES.map((s) => (
                <option key={s.key} value={s.key}>
                  {s.label}
                </option>
              ))}
            </select>
            <p className="text-xs text-gray-400 mt-1">低于此级别的实时推送会被静默；历史记录里仍可查到。</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">时区</label>
            <input
              type="text"
              value={prefs.timezone}
              onChange={(e) => setPrefs({ ...prefs, timezone: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg"
              placeholder="例如 Asia/Shanghai"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">静默开始 (HH:MM)</label>
            <input
              type="text"
              value={prefs.quiet_start}
              onChange={(e) => setPrefs({ ...prefs, quiet_start: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg font-mono"
              placeholder="22:00"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">静默结束 (HH:MM)</label>
            <input
              type="text"
              value={prefs.quiet_end}
              onChange={(e) => setPrefs({ ...prefs, quiet_end: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg font-mono"
              placeholder="07:00"
            />
          </div>
        </div>
        <p className="text-xs text-gray-400 -mt-3">
          开始时间与结束时间相同表示不开启静默；严重 (critical) 级别会无视静默直接推送。
        </p>

        <div className="flex gap-6">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={prefs.push_enabled}
              onChange={(e) => setPrefs({ ...prefs, push_enabled: e.target.checked })}
              className="w-4 h-4"
            />
            <span className="text-sm text-gray-700">启用站内实时推送</span>
          </label>
          <label className="flex items-center gap-2 opacity-60">
            <input
              type="checkbox"
              checked={prefs.email_enabled}
              onChange={(e) => setPrefs({ ...prefs, email_enabled: e.target.checked })}
              className="w-4 h-4"
            />
            <span className="text-sm text-gray-700">邮件通知（即将开放）</span>
          </label>
        </div>

        <div className="flex items-center gap-3 pt-2">
          <button
            onClick={onSave}
            disabled={saving}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2"
          >
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            保存
          </button>
          {msg && (
            <span className={`text-sm ${msg.kind === 'ok' ? 'text-green-600' : 'text-red-600'}`}>
              {msg.text}
            </span>
          )}
        </div>
      </div>
    </section>
  )
}
