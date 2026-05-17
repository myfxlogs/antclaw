// Audit API — audit logs, devices, jobs (A13-P1-01 split)

import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AdminService, ListJobsRequestSchema, SetJobEnabledRequestSchema } from '@antclaw/proto/antclaw/v1/admin_pb'
import { DeviceService, ListDevicesRequestSchema, DeleteDeviceInfoRequestSchema } from '@antclaw/proto/antclaw/v1/device_pb'
import { transport } from '../../lib/transport'

const adminClient = createClient(AdminService, transport)

// ── Audit ──

export async function listAuditLogs(params?: { pageSize?: number; cursor?: string; userId?: string; action?: string }) {
  const response: any = await (adminClient as any).listAuditLogs({
    pageSize: params?.pageSize || 50,
    userId: params?.userId || '', action: params?.action || '',
  })
  return { logs: response.entries || [], nextCursor: response.nextCursor }
}

// ── Jobs ──

export async function listJobs() {
  const response = await adminClient.listJobs(create(ListJobsRequestSchema, {}))
  return { jobs: response.jobs || [] }
}

export async function runJob(jobId: string) {
  await (adminClient as any).runJob({ jobId })
  return { success: true }
}

export async function setJobEnabled(jobId: string, enabled: boolean) {
  await adminClient.setJobEnabled(create(SetJobEnabledRequestSchema, { jobId, enabled }))
  return { success: true }
}

export async function runAllJobs() {
  const { jobs } = await listJobs()
  let ok = 0, fail = 0
  for (const j of jobs) {
    try { await runJob(j.jobId); ok++ } catch { fail++ }
  }
  return { ok, fail }
}

// ── Devices ──

export interface DeviceInfo {
  deviceId: string; model: string; brand: string; osVersion: string; osType: string
  appVersion: string; buildNumber: string; screenWidth: number; screenHeight: number
  networkType: string; timezone: string; locale: string; manufacturer: string
  fingerprint: string; userId: string; displayName: string; username: string; codeId: string
  createdAt: number; updatedAt: number
}

export async function listDevices(params?: { osTypeFilter?: string }) {
  const client = createClient(DeviceService, transport)
  const res = await client.listDevices(create(ListDevicesRequestSchema, { osTypeFilter: params?.osTypeFilter || '' }))
  return {
    devices: (res.devices || []).map((d: any) => ({
      deviceId: d.deviceId, model: d.model, brand: d.brand, osVersion: d.osVersion, osType: d.osType,
      appVersion: d.appVersion, buildNumber: d.buildNumber, screenWidth: d.screenWidth, screenHeight: d.screenHeight,
      networkType: d.networkType, timezone: d.timezone, locale: d.locale, manufacturer: d.manufacturer,
      fingerprint: d.fingerprint, userId: d.userId, displayName: d.displayName, username: d.username, codeId: d.codeId,
      createdAt: Number(d.createdAt), updatedAt: Number(d.updatedAt),
    })),
    total: res.total,
  }
}

export async function deleteDevice(deviceId: string) {
  const client = createClient(DeviceService, transport)
  const res = await client.deleteDeviceInfo(create(DeleteDeviceInfoRequestSchema, { deviceId }))
  return res.success
}
