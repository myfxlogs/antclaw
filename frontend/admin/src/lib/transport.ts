// 共享 Connect 传输层：内存 token 管理 + ConnectError 认证拦截 (A13-P0-02, A13-P0-03)
// 不再依赖 localStorage — token 由 AuthProvider 在内存中维护。

import { createConnectTransport } from '@connectrpc/connect-web'
import { createClient } from '@connectrpc/connect'
import { create } from '@bufbuild/protobuf'
import { ConnectError, Code } from '@connectrpc/connect'
import { AuthService, RefreshRequestSchema } from '@antclaw/proto/antclaw/v1/auth_pb'
import {
  getMemAccessToken,
  getMemRefreshToken,
  clearMemSession,
} from '../features/auth/AuthProvider'

export const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL ?? 'http://localhost:8082'

// Single-flight refresh state
let refreshPromise: Promise<string | null> | null = null

async function tryRefresh(): Promise<string | null> {
  const refreshToken = getMemRefreshToken()
  if (!refreshToken) return null
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const noAuthTransport = createConnectTransport({
        baseUrl: API_BASE_URL,
        useBinaryFormat: true,
      })
      const client = createClient(AuthService, noAuthTransport)
      const res = await client.refresh(
        create(RefreshRequestSchema, { refreshToken }),
      )
      if (res.accessToken) {
        return res.accessToken
      }
    } catch {
      // Refresh failed — clear all in-memory state
      clearMemSession()
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    } finally {
      refreshPromise = null
    }
    return null
  })()

  return refreshPromise
}

// Create transport with auth + refresh interceptor (memory-based tokens)
export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  useBinaryFormat: true,
  interceptors: [
    (next) => async (req) => {
      // Attach access token from memory
      const token = getMemAccessToken()
      if (token) {
        req.header.set('Authorization', `Bearer ${token}`)
      }

      try {
        return await next(req)
      } catch (e: unknown) {
        // A13-P0-03: properly catch ConnectError Unauthenticated
        if (e instanceof ConnectError && e.code === Code.Unauthenticated) {
          const newToken = await tryRefresh()
          if (newToken) {
            // Retry once with fresh token
            req.header.set('Authorization', `Bearer ${newToken}`)
            return next(req)
          }
          // Refresh failed — already redirected by tryRefresh
        }
        // Re-throw all other errors for callers to handle
        throw e
      }
    },
  ],
})
