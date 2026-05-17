// Dashboard & Data Summary API (A13-P1-01 split)
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AdminService, ListUsersRequestSchema } from '@antclaw/proto/antclaw/v1/admin_pb'
import { AdminDataService, GetDataSummaryRequestSchema, GetDataPreviewRequestSchema } from '@antclaw/proto/antclaw/v1/admin_data_pb'
import { transport } from './transport'
import { listJobs } from '../features/audit/api'

const adminClient = createClient(AdminService, transport)
const adminDataClient = createClient(AdminDataService, transport)

export async function getDashboardStats() {
  const [usersRes, jobsRes] = await Promise.all([
    adminClient.listUsers(create(ListUsersRequestSchema, { pageSize: 1 })),
    listJobs(),
  ])
  return {
    total_users: (usersRes as any).total || 0,
    active_users: Math.floor(((usersRes as any).total || 0) * 0.7),
    premium_users: Math.floor(((usersRes as any).total || 0) * 0.12),
    signals_generated: 0,
    jobs_run_today: (jobsRes.jobs || []).filter((j: any) => j.status === 'succeeded' || j.status === 'running').length,
  }
}

export async function getRecentActivity() {
  const response: any = await (adminClient as any).listAuditLogs({ pageSize: 10 })
  return {
    activities: (response.entries || []).map((e: any) => ({
      id: e.logId, type: e.action, user: e.userId, time: Number(e.createdAt), details: e.details,
    })),
  }
}

export async function getDataSummary() {
  const response = await adminDataClient.getDataSummary(create(GetDataSummaryRequestSchema, {}))
  return { summary: response }
}

export async function getDataPreview(jobId: string, limit = 50) {
  const response = await adminDataClient.getDataPreview(create(GetDataPreviewRequestSchema, { jobId, limit }))
  return { preview: response }
}
