// Auth API — login, logout, token management (A13-P1-01 split)
// Deprecated: use AuthProvider from src/features/auth/ instead

import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AuthService, LoginRequestSchema, ClientInfoSchema, RefreshRequestSchema } from '@antclaw/proto/antclaw/v1/auth_pb'
import { transport } from '../../lib/transport'

const authClient = createClient(AuthService, transport)

export async function login(email: string, password: string) {
  const clientInfo = create(ClientInfoSchema, { userAgent: navigator.userAgent, ipAddress: '127.0.0.1' })
  const response = await authClient.login(create(LoginRequestSchema, { email, password, client: clientInfo }))
  return { user_id: response.userId, access_token: response.accessToken, refresh_token: response.refreshToken, expires_at: response.expiresAt }
}

export function logout() {
  localStorage.removeItem('token')
  localStorage.removeItem('refreshToken')
}

export async function refreshToken(refreshToken: string) {
  const res = await authClient.refresh(create(RefreshRequestSchema, { refreshToken }))
  return res.accessToken
}
