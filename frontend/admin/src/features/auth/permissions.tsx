// permissions.ts — 管理端权限枚举与 RequirePermission 组件 (A13-P0-01)

import { type ReactNode } from 'react'
import { useAuth } from './AuthProvider'

// ── Permission constants ──

export const Permissions = {
  USERS_READ: 'admin.users.read',
  USERS_BAN: 'admin.users.ban',
  USERS_RESET_PASSWORD: 'admin.users.reset_password',
  USERS_SET_ROLE: 'admin.users.set_role',
  SOCIAL_READ: 'admin.social.read',
  SOCIAL_MODERATE: 'admin.social.moderate',
  SOCIAL_DELETE: 'admin.social.delete',
  PUSH_SEND: 'admin.push.send',
  AUDIT_READ: 'admin.audit.read',
  SYSTEM_MANAGE: 'admin.system.manage',
  AI_MANAGE: 'admin.ai.manage',
} as const

export type Permission = (typeof Permissions)[keyof typeof Permissions]

// ── RequirePermission component ──

interface RequirePermissionProps {
  permission: string
  /** Content to show when user lacks permission (default: null) */
  fallback?: ReactNode
  children: ReactNode
}

/** Wraps children; renders nothing (or a fallback) if the user lacks the given permission. */
export function RequirePermission({
  permission,
  fallback = null,
  children,
}: RequirePermissionProps) {
  const { hasPermission } = useAuth()
  if (!hasPermission(permission)) return <>{fallback}</>
  return <>{children}</>
}

// ── RequireAnyPermission ──

interface RequireAnyPermissionProps {
  permissions: string[]
  fallback?: ReactNode
  children: ReactNode
}

/** Wraps children; renders nothing (or fallback) if the user lacks ALL specified permissions. */
export function RequireAnyPermission({
  permissions,
  fallback = null,
  children,
}: RequireAnyPermissionProps) {
  const { hasPermission } = useAuth()
  const ok = permissions.some((p) => hasPermission(p))
  if (!ok) return <>{fallback}</>
  return <>{children}</>
}
