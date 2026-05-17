// Push API — send, history, stats, online users (A13-P1-01 split)

import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AdminService, SendPushRequestSchema, GetPushHistoryRequestSchema } from '@antclaw/proto/antclaw/v1/admin_pb'
import { SystemService, HealthzRequestSchema, GetOnlineUsersRequestSchema, GetPushStatsRequestSchema } from '@antclaw/proto/antclaw/v1/system_pb'
import { transport } from '../../lib/transport'

const adminClient = createClient(AdminService, transport)
const systemClient = createClient(SystemService, transport)

export async function sendPush(params: { title: string; body: string; severity: string; targetUserIds: string[]; category: string }) {
  const res = await adminClient.sendPush(create(SendPushRequestSchema, {
    title: params.title, body: params.body, severity: params.severity,
    targetUserIds: params.targetUserIds, category: params.category,
  }))
  return { sentCount: res.sentCount, onlineCount: res.onlineCount, pushLogId: res.pushLogId }
}

export async function getPushHistory(params?: { pageSize?: number; cursor?: string }) {
  const res = await adminClient.getPushHistory(create(GetPushHistoryRequestSchema, {
    pageSize: params?.pageSize || 50, cursor: params?.cursor || '',
  }))
  return {
    entries: (res.entries || []).map(e => ({
      id: e.id, title: e.title, body: e.body, severity: e.severity,
      targetCount: e.targetCount, sentCount: e.sentCount, adminUserId: e.adminUserId, createdAt: Number(e.createdAt),
    })),
    nextCursor: res.nextCursor,
  }
}

export async function getPushStats() {
  const res = await systemClient.getPushStats(create(GetPushStatsRequestSchema, {}))
  return { totalNotifications: Number(res.totalNotifications), totalPushStateRecords: Number(res.totalPushStateRecords), byType: (res.byType || []).map((t: any) => ({ pushType: t.pushType, count: t.count })), recent1h: Number(res.recent1h), recent24h: Number(res.recent24h) }
}

export interface OnlineUser { userId: string; codeId: string; displayName: string; email: string; userAgent: string; remoteAddr: string; connectedAt: number }

export async function getOnlineUsers() {
  const res = await systemClient.getOnlineUsers(create(GetOnlineUsersRequestSchema, {}))
  return { count: res.count, users: (res.users || []).map((u: any) => ({ userId: u.userId, codeId: u.codeId || '', displayName: u.displayName || '', email: u.email || '', userAgent: u.userAgent || '', remoteAddr: u.remoteAddr, connectedAt: Number(u.connectedAt) })) }
}

export async function getSystemHealth() {
  const res = await systemClient.healthz(create(HealthzRequestSchema, {}))
  return { status: res.status || 'unknown', version: '0.1.0', uptime: 'unknown', db_status: (res.components as any)?.postgres?.status || 'unknown', redis_status: (res.components as any)?.redis?.status || 'unknown' }
}
