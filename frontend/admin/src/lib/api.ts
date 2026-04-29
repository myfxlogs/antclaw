import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { AuthService, LoginRequestSchema, ClientInfoSchema } from '@antclaw/proto/antclaw/v1/auth_pb'
import {
  AdminService,
  ListUsersRequestSchema,
  BanRequestSchema,
  UnbanRequestSchema,
  SetRoleRequestSchema,
  RunJobRequestSchema,
  ListJobsRequestSchema,
  SetJobEnabledRequestSchema,
  ListAuditLogsRequestSchema,
  AdminResetUserPasswordRequestSchema,
  SetUserCodeIDRequestSchema,
} from '@antclaw/proto/antclaw/v1/admin_pb'
import {
  ResetPasswordRequestSchema,
} from '@antclaw/proto/antclaw/v1/auth_pb'
import {
  StrategyService,
  ListStrategiesRequestSchema,
  RunStrategyRequestSchema,
  EnableStrategyRequestSchema,
  DisableStrategyRequestSchema,
  ListStrategyRunsRequestSchema,
} from '@antclaw/proto/antclaw/v1/strategy_pb'
import {
  SystemAIService,
  ListSystemAIConfigsRequestSchema,
  UpdateSystemAIConfigRequestSchema,
  UpdateSystemAISecretRequestSchema,
  DiscoverSystemAIModelsRequestSchema,
  ValidateSystemAIConnectionRequestSchema,
} from '@antclaw/proto/antclaw/v1/system_ai_pb'
import {
  DataSourceService,
  ListDataSourcesRequestSchema,
  UpdateDataSourceRequestSchema,
} from '@antclaw/proto/antclaw/v1/datasource_pb'
import {
  AdminDataService,
  GetDataSummaryRequestSchema,
  GetDataPreviewRequestSchema,
} from '@antclaw/proto/antclaw/v1/admin_data_pb'
import { CryptoService, GetCryptoPublicKeyRequestSchema } from '@antclaw/proto/antclaw/v1/crypto_pb'
import { SystemService, HealthzRequestSchema } from '@antclaw/proto/antclaw/v1/system_pb'
import {
  NotificationService,
  ListUnreadRequestSchema,
  ListHistoryRequestSchema,
  UnreadCountRequestSchema,
  MarkReadRequestSchema,
  MarkAllReadRequestSchema,
  GetPrefsRequestSchema,
  NotificationPrefsSchema,
} from '@antclaw/proto/antclaw/v1/notification_pb'

const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8082'

// Create transport with auth interceptor
const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem('token')
      if (token) {
        req.header.set('Authorization', `Bearer ${token}`)
      }
      return await next(req)
    },
  ],
})

// Service clients (v2: createClient)
const authClient = createClient(AuthService, transport)
const adminClient = createClient(AdminService, transport)
const strategyClient = createClient(StrategyService, transport)
const systemAIClient = createClient(SystemAIService, transport)
const dataSourceClient = createClient(DataSourceService, transport)
const adminDataClient = createClient(AdminDataService, transport)
const cryptoClient = createClient(CryptoService, transport)
const systemClient = createClient(SystemService, transport)
const notificationClient = createClient(NotificationService, transport)

function normalizeMetrics(input: unknown): Record<string, number> {
  if (!input || typeof input !== 'object') return {}
  const out: Record<string, number> = {}
  for (const [key, value] of Object.entries(input as Record<string, unknown>)) {
    const n = Number(value)
    if (!Number.isNaN(n)) {
      out[key] = n
    }
  }
  return out
}

/** RSA 公钥 PEM（Connect，替代旧 GET /crypto/pubkey）。 */
export async function getCryptoPublicKeyPem(): Promise<string> {
  const res = await cryptoClient.getCryptoPublicKey(create(GetCryptoPublicKeyRequestSchema, {}))
  return res.pem || ''
}

// Auth API
export async function login(email: string, password: string) {
  const clientInfo = create(ClientInfoSchema, {
    userAgent: navigator.userAgent,
    ipAddress: '127.0.0.1',
  })
  const request = create(LoginRequestSchema, { email, password, client: clientInfo })
  const response = await authClient.login(request)
  
  // Store token
  if (response.accessToken) {
    localStorage.setItem('token', response.accessToken)
  }
  
  return {
    user_id: response.userId,
    access_token: response.accessToken,
    refresh_token: response.refreshToken,
    expires_at: response.expiresAt,
  }
}

export async function logout() {
  localStorage.removeItem('token')
}

// Dashboard API - Using AdminService
export async function getDashboardStats() {
  const [usersRes, jobsRes] = await Promise.all([
    adminClient.listUsers(create(ListUsersRequestSchema, { pageSize: 1 })),
    adminClient.listJobs(create(ListJobsRequestSchema, {})),
  ])

  return {
    total_users: (usersRes as any).total || 0,
    active_users: Math.floor(((usersRes as any).total || 0) * 0.7),
    premium_users: Math.floor(((usersRes as any).total || 0) * 0.12),
    signals_generated: 0,
    jobs_run_today: (jobsRes.jobs || []).filter((j: any) =>
      j.status === 'succeeded' || j.status === 'running'
    ).length,
  }
}

export async function getRecentActivity() {
  const response = await adminClient.listAuditLogs(create(ListAuditLogsRequestSchema, { pageSize: 10 }))
  return {
    activities: (response.entries || []).map((e: any) => ({
      id: e.logId,
      type: e.action,
      user: e.userId,
      time: Number(e.createdAt),
      details: e.details,
    })),
  }
}

export async function getSystemHealth() {
  const res = await systemClient.healthz(create(HealthzRequestSchema, {}))
  const pg = (res.components as any)?.postgres?.status || 'unknown'
  const rd = (res.components as any)?.redis?.status || 'unknown'
  return {
    status: res.status || 'unknown',
    version: '0.1.0',
    uptime: 'unknown',
    db_status: pg,
    redis_status: rd,
  }
}

// Users API
export async function listUsers(params?: {
  cursor?: string
  page_size?: number
  email_filter?: string
  role_filter?: string
  banned_only?: boolean
}) {
  const request = create(ListUsersRequestSchema, {
    cursor: params?.cursor || '',
    pageSize: params?.page_size || 20,
    emailFilter: params?.email_filter || '',
    roleFilter: params?.role_filter || '',
    bannedOnly: params?.banned_only || false,
  })
  const response = await adminClient.listUsers(request)
  
  return {
    users: response.users.map((u: any) => ({
      user_id: u.userId,
      email: u.email,
      username: u.username,
      display_name: u.displayName,
      roles: u.roles || [],
      created_at: Number(u.createdAt),
      code_id: u.codeId || '',
    })),
    next_cursor: response.nextCursor,
  }
}

// 管理员设置/重置用户的数字 ID。codeId 留空则后端自动重新随机分配。
export async function setUserCodeID(userId: string, codeId: string) {
  const response = await adminClient.setUserCodeID(
    create(SetUserCodeIDRequestSchema, { userId, codeId }),
  )
  return { code_id: response.codeId }
}

export async function banUser(userId: string, reason: string, expiresAt?: number) {
  const request = create(BanRequestSchema, {
    userId,
    reason,
    expiresAt: expiresAt ? BigInt(expiresAt) : BigInt(0),
  })
  await adminClient.ban(request)
  return { success: true }
}

export async function unbanUser(userId: string) {
  const request = create(UnbanRequestSchema, { userId })
  await adminClient.unban(request)
  return { success: true }
}

export async function setUserRole(userId: string, roles: string[]) {
  const request = create(SetRoleRequestSchema, { userId, roles })
  await adminClient.setRole(request)
  return { success: true }
}

// Jobs API
export async function listJobs(params?: { status_filter?: string }) {
  const request = create(ListJobsRequestSchema, {
    statusFilter: params?.status_filter || '',
  })
  const response = await adminClient.listJobs(request)
  return {
    jobs: response.jobs.map((j: any) => ({
      job_id: j.jobId,
      job_name: j.jobName,
      status: j.status,
      last_run: j.lastRun,
      next_run: j.nextRun,
      enabled: j.enabled,
      last_error: j.lastError || '',
    })),
  }
}

export async function runJob(jobName: string, params?: Record<string, string>) {
  const request = create(RunJobRequestSchema, {
    jobName,
    params: params || {},
  })
  const response = await adminClient.runJob(request)
  return { success: true, job_id: response.jobId }
}

export async function setJobEnabled(jobId: string, enabled: boolean) {
  const request = create(SetJobEnabledRequestSchema, { jobId, enabled })
  const response = await adminClient.setJobEnabled(request)
  return { job_id: response.jobId, enabled: response.enabled }
}

// Audit API
export async function listAuditLogs(params?: {
  cursor?: string
  page_size?: number
  user_id_filter?: string
  action_filter?: string
}) {
  const request = create(ListAuditLogsRequestSchema, {
    cursor: params?.cursor || '',
    pageSize: params?.page_size || 20,
    userIdFilter: params?.user_id_filter || '',
    actionFilter: params?.action_filter || '',
  })
  const response = await adminClient.listAuditLogs(request)
  return {
    logs: response.entries.map((e: any) => ({
      log_id: e.logId,
      user_id: e.userId,
      action: e.action,
      resource: e.resource,
      details: e.details,
      created_at: Number(e.createdAt),
      ip_address: e.ipAddress,
    })),
    next_cursor: response.nextCursor,
  }
}

// Reset Password API (admin can reset user password)
export async function resetPassword(token: string, newPassword: string) {
  const request = create(ResetPasswordRequestSchema, {
    token,
    newPassword,
  })
  const response = await authClient.resetPassword(request)
  return {
    user_id: response.userId,
  }
}

// Admin Reset Password - directly set password for a user
export async function adminResetPassword(userId: string, newPassword: string) {
  await adminClient.adminResetUserPassword(
    create(AdminResetUserPasswordRequestSchema, { userId, newPassword }),
  )
  return { success: true }
}

export async function listStrategies() {
  const response = await strategyClient.listStrategies(create(ListStrategiesRequestSchema, {}))
  return {
    items: response.items.map((it: any) => ({
      id: it.id,
      name: it.name,
      kind: it.kind,
      symbol: it.symbol,
      timeframe: it.timeframe,
      enabled: it.enabled,
      status: it.status,
      schedule_cron: it.scheduleCron,
      last_run_at: it.lastRunAt,
      last_run_status: it.lastRunStatus,
      updated_at: it.updatedAt,
    })),
    total: response.total,
  }
}

export async function runStrategy(id: string) {
  const response = await strategyClient.runStrategy(create(RunStrategyRequestSchema, { id }))
  const run = response.item
  return {
    run_id: run?.runId,
    strategy_id: run?.strategyId,
    started_at: run?.startedAt,
    finished_at: run?.finishedAt,
    status: run?.status,
    metrics: normalizeMetrics(run?.metrics),
    mock: run?.mock || false,
    error_message: run?.errorMessage || '',
  }
}

export async function enableStrategy(id: string) {
  await strategyClient.enableStrategy(create(EnableStrategyRequestSchema, { id }))
  return { success: true }
}

export async function disableStrategy(id: string) {
  await strategyClient.disableStrategy(create(DisableStrategyRequestSchema, { id }))
  return { success: true }
}

export async function listStrategyRuns(id: string, limit = 20) {
  const response = await strategyClient.listStrategyRuns(create(ListStrategyRunsRequestSchema, { id, limit }))
  return {
    items: response.items.map((run: any) => ({
      run_id: run.runId,
      strategy_id: run.strategyId,
      started_at: run.startedAt,
      finished_at: run.finishedAt,
      status: run.status,
      metrics: normalizeMetrics(run.metrics),
      mock: run.mock,
      error_message: run.errorMessage,
    })),
  }
}

export async function listSystemAIConfigs() {
  const response = await systemAIClient.listConfigs(create(ListSystemAIConfigsRequestSchema, {}))
  return {
    items: response.items.map((it: any) => ({
      provider_id: it.providerId,
      name: it.name,
      base_url: it.baseUrl,
      organization: it.organization,
      models: it.models || [],
      default_model: it.defaultModel,
      temperature: it.temperature,
      timeout_seconds: it.timeoutSeconds,
      max_tokens: it.maxTokens,
      purposes: it.purposes || [],
      primary_for: it.primaryFor || [],
      enabled: it.enabled,
      has_secret: it.hasSecret,
      updated_at: it.updatedAt,
    })),
  }
}

export async function updateSystemAIConfig(providerId: string, payload: Record<string, unknown>) {
  await systemAIClient.updateConfig(create(UpdateSystemAIConfigRequestSchema, {
    providerId,
    name: String(payload.name || ''),
    baseUrl: String(payload.base_url || ''),
    organization: String(payload.organization || ''),
    models: (payload.models as string[]) || [],
    defaultModel: String(payload.default_model || ''),
    temperature: Number(payload.temperature || 0),
    timeoutSeconds: Number(payload.timeout_seconds || 0),
    maxTokens: Number(payload.max_tokens || 0),
    purposes: (payload.purposes as string[]) || [],
    primaryFor: (payload.primary_for as string[]) || [],
    enabled: Boolean(payload.enabled),
  }))
  return { provider_id: providerId }
}

export async function updateSystemAISecret(providerId: string, secret: string) {
  await systemAIClient.updateSecret(create(UpdateSystemAISecretRequestSchema, { providerId, secret }))
  return { provider_id: providerId, secret_updated: true }
}

export async function clearSystemAISecret(providerId: string) {
  await systemAIClient.updateSecret(create(UpdateSystemAISecretRequestSchema, { providerId, secret: '' }))
  return { provider_id: providerId, secret_updated: false }
}

export async function discoverSystemAIModels(providerId: string) {
  const response = await systemAIClient.discoverModels(create(DiscoverSystemAIModelsRequestSchema, {
    providerId,
  }))
  return {
    provider_id: response.providerId,
    models: response.models,
    default_model: response.defaultModel,
  }
}

export async function validateSystemAI(providerId: string) {
  const response = await systemAIClient.validateConnection(create(ValidateSystemAIConnectionRequestSchema, {
    providerId,
  }))
  return {
    provider_id: response.providerId,
    ok: response.ok,
    model_count: response.modelCount,
  }
}

export async function listDataSources() {
  const response = await dataSourceClient.listDataSources(create(ListDataSourcesRequestSchema, {}))
  return {
    items: response.items.map((it: any) => ({
      source_id: it.sourceId,
      name: it.name,
      kind: it.kind,
      endpoint: it.endpoint,
      has_secret: it.hasSecret,
      updated_at: it.updatedAt,
      updated_by: it.updatedBy,
    })),
  }
}

export async function updateDataSource(sourceId: string, payload: { endpoint?: string; secret?: string; clear_secret?: boolean }) {
  const request = create(UpdateDataSourceRequestSchema, {
    sourceId,
    endpoint: payload.endpoint,
    secret: payload.secret,
    clearSecret: Boolean(payload.clear_secret),
  })
  const response = await dataSourceClient.updateDataSource(request)
  return {
    item: {
      source_id: response.item?.sourceId,
      name: response.item?.name,
      kind: response.item?.kind,
      endpoint: response.item?.endpoint,
      has_secret: response.item?.hasSecret,
      updated_at: response.item?.updatedAt,
      updated_by: response.item?.updatedBy,
    },
  }
}

/** 采集数据汇总（原 GET /admin/data/summary）。 */
export async function getDataSummary() {
  const response = await adminDataClient.getDataSummary(create(GetDataSummaryRequestSchema, {}))
  return {
    items: response.items.map((it: any) => ({
      job_id: it.jobId,
      name: it.name,
      table: it.table,
      count: Number(it.count),
      latest_time: it.latestTime ? Number(it.latestTime) : 0,
      error: it.error || undefined,
    })),
    updated_at: Number(response.updatedAt),
  }
}

/** 采集数据预览（原 GET /admin/data/preview）。 */
export async function getDataPreview(jobId: string, limit = 50) {
  const response = await adminDataClient.getDataPreview(
    create(GetDataPreviewRequestSchema, { jobId, limit }),
  )
  let rows: Record<string, unknown>[] = []
  if (response.rowsJson) {
    try {
      rows = JSON.parse(response.rowsJson) as Record<string, unknown>[]
    } catch {
      rows = []
    }
  }
  return {
    job_id: response.jobId,
    table: response.table,
    time_col: response.timeCol,
    columns: [...response.columns],
    rows,
    total_sampled: response.totalSampled,
  }
}

// ============================================================================
// 通知推送（NotificationService）
// 实时通道：GET /sse/notifications?access_token=...
// ============================================================================

export interface NotificationItem {
  id: string
  type: string
  category: string
  severity: string
  title: string
  body: string
  data: Record<string, string>
  is_read: boolean
  created_at: number
  read_at: number
}

function notifFromProto(n: any): NotificationItem {
  return {
    id: String(n.id || ''),
    type: String(n.type || 'in_app'),
    category: String(n.category || 'system'),
    severity: String(n.severity || 'normal'),
    title: String(n.title || ''),
    body: String(n.body || ''),
    data: (n.data && typeof n.data === 'object') ? n.data : {},
    is_read: Boolean(n.isRead),
    created_at: Number(n.createdAt || 0),
    read_at: Number(n.readAt || 0),
  }
}

export async function listUnreadNotifications(limit = 50): Promise<NotificationItem[]> {
  const r = await notificationClient.listUnread(create(ListUnreadRequestSchema, { limit }))
  return (r.items || []).map(notifFromProto)
}

export async function listNotificationHistory(limit = 50): Promise<NotificationItem[]> {
  const r = await notificationClient.listHistory(create(ListHistoryRequestSchema, { limit }))
  return (r.items || []).map(notifFromProto)
}

export async function getUnreadNotificationCount(): Promise<number> {
  const r = await notificationClient.unreadCount(create(UnreadCountRequestSchema, {}))
  return Number(r.count || 0)
}

export async function markNotificationRead(id: string): Promise<void> {
  await notificationClient.markRead(create(MarkReadRequestSchema, { id }))
}

export async function markAllNotificationsRead(): Promise<void> {
  await notificationClient.markAllRead(create(MarkAllReadRequestSchema, {}))
}

export interface NotificationPrefsItem {
  enabled_types: string[]
  min_severity: string
  quiet_start: string
  quiet_end: string
  timezone: string
  push_enabled: boolean
  email_enabled: boolean
}

export async function getNotificationPrefs(): Promise<NotificationPrefsItem> {
  const r = await notificationClient.getPrefs(create(GetPrefsRequestSchema, {}))
  return {
    enabled_types: r.enabledTypes || [],
    min_severity: r.minSeverity || 'low',
    quiet_start: r.quietStart || '00:00',
    quiet_end: r.quietEnd || '00:00',
    timezone: r.timezone || 'UTC',
    push_enabled: Boolean(r.pushEnabled),
    email_enabled: Boolean(r.emailEnabled),
  }
}

export async function updateNotificationPrefs(p: NotificationPrefsItem): Promise<NotificationPrefsItem> {
  const r = await notificationClient.updatePrefs(create(NotificationPrefsSchema, {
    enabledTypes: p.enabled_types,
    minSeverity: p.min_severity,
    quietStart: p.quiet_start,
    quietEnd: p.quiet_end,
    timezone: p.timezone,
    pushEnabled: p.push_enabled,
    emailEnabled: p.email_enabled,
  }))
  return {
    enabled_types: r.enabledTypes || [],
    min_severity: r.minSeverity || 'low',
    quiet_start: r.quietStart || '00:00',
    quiet_end: r.quietEnd || '00:00',
    timezone: r.timezone || 'UTC',
    push_enabled: Boolean(r.pushEnabled),
    email_enabled: Boolean(r.emailEnabled),
  }
}

// 个人通知 SSE：返回 EventSource，调用方负责绑定 onmessage / 关闭。
export function openNotificationsSSE(onEvent: (n: NotificationItem) => void): () => void {
  const token = localStorage.getItem('token') || ''
  const url = `${API_BASE_URL}/sse/notifications?access_token=${encodeURIComponent(token)}`
  const es = new EventSource(url)
  es.addEventListener('notification', (ev: MessageEvent) => {
    try {
      const raw = JSON.parse(ev.data)
      onEvent({
        id: '',
        type: String(raw.type || 'in_app'),
        category: String(raw.category || 'system'),
        severity: String(raw.severity || 'normal'),
        title: String(raw.title || ''),
        body: String(raw.body || ''),
        data: raw.data || {},
        is_read: false,
        created_at: Math.floor(Date.now() / 1000),
        read_at: 0,
      })
    } catch {}
  })
  return () => es.close()
}
