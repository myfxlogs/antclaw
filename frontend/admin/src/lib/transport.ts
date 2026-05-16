// 共享 Connect 传输层：认证拦截 + 单飞 token 刷新
import { createConnectTransport } from '@connectrpc/connect-web'
import { createClient } from '@connectrpc/connect'
import { create } from '@bufbuild/protobuf'
import { AuthService, RefreshRequestSchema } from '@antclaw/proto/antclaw/v1/auth_pb'

export const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL ?? 'http://localhost:8082'

// Single-flight refresh state
let refreshPromise: Promise<string | null> | null = null

async function tryRefresh(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refreshToken')
  if (!refreshToken) return null
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const noAuthTransport = createConnectTransport({ baseUrl: API_BASE_URL, useBinaryFormat: true })
      const client = createClient(AuthService, noAuthTransport)
      const res = await client.refresh(create(RefreshRequestSchema, { refreshToken }))
      if (res.accessToken) {
        localStorage.setItem('token', res.accessToken)
        return res.accessToken
      }
    } catch {
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
      window.location.href = '/login'
    } finally {
      refreshPromise = null
    }
    return null
  })()

  return refreshPromise
}

// Create transport with auth + refresh interceptor
export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  useBinaryFormat: true,
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem('token')
      if (token) req.header.set('Authorization', `Bearer ${token}`)
      const res = await next(req)
      // On 401, attempt token refresh
      if ((res as any).code === 'unauthenticated') {
        const newToken = await tryRefresh()
        if (newToken) {
          req.header.set('Authorization', `Bearer ${newToken}`)
          return next(req)
        }
      }
      return res
    },
  ],
})
