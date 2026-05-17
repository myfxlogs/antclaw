// Notifications API (A13-P1-01 split)

import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { NotificationService, ListUnreadRequestSchema, ListHistoryRequestSchema, UnreadCountRequestSchema, MarkReadRequestSchema, MarkAllReadRequestSchema, GetPrefsRequestSchema, UpdatePrefsRequestSchema, NotificationPrefsSchema } from '@antclaw/proto/antclaw/v1/notification_pb'
import { transport, API_BASE_URL } from '../../lib/transport'

const notificationClient = createClient(NotificationService, transport)

export interface NotificationItem {
  id: string; userId: string; type: string; category: string; severity: string
  title: string; body: string; data: Record<string, string>; isRead: boolean
  createdAt: number; readAt: number
}

function notifFromProto(n: any): NotificationItem {
  return { id: n.id, userId: n.userId, type: n.type, category: n.category, severity: n.severity, title: n.title, body: n.body, data: n.data || {}, isRead: n.isRead, createdAt: Number(n.createdAt), readAt: Number(n.readAt) }
}

export async function listUnreadNotifications(limit = 50) {
  const r = await notificationClient.listUnread(create(ListUnreadRequestSchema, { limit }))
  return (r.items || []).map(notifFromProto)
}

export async function listNotificationHistory(limit = 50) {
  const r = await notificationClient.listHistory(create(ListHistoryRequestSchema, { limit }))
  return (r.items || []).map(notifFromProto)
}

export async function getUnreadNotificationCount() {
  const r = await notificationClient.unreadCount(create(UnreadCountRequestSchema, {}))
  return Number(r.count || 0)
}

export async function markNotificationRead(id: string) {
  await notificationClient.markRead(create(MarkReadRequestSchema, { id }))
}

export async function markAllNotificationsRead() {
  await notificationClient.markAllRead(create(MarkAllReadRequestSchema, {}))
}

export interface NotificationPrefsItem {
  enabled_types: string[]; min_severity: string; quiet_start: string; quiet_end: string
  timezone: string; push_enabled: boolean; email_enabled: boolean
}

export async function getNotificationPrefs() {
  const resp = await notificationClient.getPrefs(create(GetPrefsRequestSchema, {}))
  const r = resp.prefs
  return r ? { enabled_types: r.enabledTypes || [], min_severity: r.minSeverity || 'low', quiet_start: r.quietStart || '00:00', quiet_end: r.quietEnd || '00:00', timezone: r.timezone || 'UTC', push_enabled: r.pushEnabled, email_enabled: r.emailEnabled } : { enabled_types: [], min_severity: 'low', quiet_start: '00:00', quiet_end: '00:00', timezone: 'UTC', push_enabled: true, email_enabled: false }
}

export async function updateNotificationPrefs(p: NotificationPrefsItem) {
  const prefs = create(NotificationPrefsSchema, { enabledTypes: p.enabled_types, minSeverity: p.min_severity, quietStart: p.quiet_start, quietEnd: p.quiet_end, timezone: p.timezone, pushEnabled: p.push_enabled, emailEnabled: p.email_enabled })
  const resp = await notificationClient.updatePrefs(create(UpdatePrefsRequestSchema, { prefs }))
  const r2 = resp.prefs
  return r2 ? { enabled_types: r2.enabledTypes || [], min_severity: r2.minSeverity || 'low', quiet_start: r2.quietStart || '00:00', quiet_end: r2.quietEnd || '00:00', timezone: r2.timezone || 'UTC', push_enabled: r2.pushEnabled, email_enabled: r2.emailEnabled } : { enabled_types: [], min_severity: 'low', quiet_start: '00:00', quiet_end: '00:00', timezone: 'UTC', push_enabled: true, email_enabled: false }
}

export function openNotificationsSSE(onEvent: (n: NotificationItem) => void): () => void {
  const token = localStorage.getItem('token') || ''
  const base = API_BASE_URL.replace(/\/+$/, '')
  const es = new EventSource(`${base}/sse/notifications?token=${encodeURIComponent(token)}`)
  es.onmessage = (ev) => {
    try {
      const data = JSON.parse(ev.data)
      if (data && data.id) onEvent(notifFromProto(data))
    } catch {}
  }
  es.onerror = () => es.close()
  return () => es.close()
}
