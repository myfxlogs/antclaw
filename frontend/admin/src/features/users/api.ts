// Users API — list, ban, unban, set role, reset password, set codeID (A13-P1-01 split)

import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AdminService, ListUsersRequestSchema, BanRequestSchema, UnbanRequestSchema, SetRoleRequestSchema, SetUserCodeIDRequestSchema } from '@antclaw/proto/antclaw/v1/admin_pb'
import { transport } from '../../lib/transport'

const adminClient = createClient(AdminService, transport)

export async function listUsers(params?: { cursor?: string; page_size?: number; email_filter?: string; role_filter?: string; banned_only?: boolean }) {
  const response = await adminClient.listUsers(create(ListUsersRequestSchema, {
    cursor: params?.cursor || '', pageSize: params?.page_size || 20,
    emailFilter: params?.email_filter || '', roleFilter: params?.role_filter || '',
    bannedOnly: params?.banned_only || false,
  }))
  return {
    users: response.users.map((u: any) => ({
      user_id: u.userId, email: u.email, username: u.username, display_name: u.displayName,
      roles: u.roles || [], created_at: Number(u.createdAt), code_id: u.codeId || '',
    })),
    next_cursor: response.nextCursor,
  }
}

export async function banUser(userId: string, reason: string, expiresAt?: number) {
  await adminClient.ban(create(BanRequestSchema, { userId, reason, expiresAt: expiresAt ? BigInt(expiresAt) : BigInt(0) }))
  return { success: true }
}

export async function unbanUser(userId: string) {
  await adminClient.unban(create(UnbanRequestSchema, { userId }))
  return { success: true }
}

export async function setUserRole(userId: string, roles: string[]) {
  await adminClient.setRole(create(SetRoleRequestSchema, { userId, roles }))
  return { success: true }
}

export async function adminResetPassword(userId: string, newPassword: string) {
  const response: any = await (adminClient as any).adminResetUserPassword({ userId, newPassword })
  return { success: response.success, temporary_password: response.temporaryPassword }
}

export async function setUserCodeID(userId: string, codeId: string) {
  const response = await adminClient.setUserCodeID(create(SetUserCodeIDRequestSchema, { userId, codeId }))
  return { code_id: response.codeId }
}
