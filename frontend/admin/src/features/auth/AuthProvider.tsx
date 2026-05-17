// AuthProvider — 管理端认证与权限上下文 (A13-P0-01, A13-P0-02)
// 内存态 token 管理，不使用 localStorage 存储敏感凭据。

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useMemo,
  useEffect,
  type ReactNode,
} from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import {
  AuthService,
  LoginRequestSchema,
  RefreshRequestSchema,
  ClientInfoSchema,
} from '@antclaw/proto/antclaw/v1/auth_pb'

export const API_BASE_URL =
  (import.meta as any).env?.VITE_API_BASE_URL ?? 'http://localhost:8082'

// ── Types ──

export interface CurrentUser {
  userId: string
  email: string
  roles: string[]
  permissions: string[]
}

interface AuthState {
  currentUser: CurrentUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  hasPermission: (perm: string) => boolean
  requirePermission: (perm: string) => boolean
  refreshToken: () => Promise<void>
  getAccessToken: () => string | null
  getRefreshToken: () => string | null
}

const AuthContext = createContext<AuthState | null>(null)

// ── In-memory token store ──

let memAccessToken: string | null = null
let memRefreshToken: string | null = null
let memUser: CurrentUser | null = null

// ── Helper: create a transport WITHOUT auth interceptor (for refresh) ──

function noAuthTransport() {
  return createConnectTransport({
    baseUrl: API_BASE_URL,
    useBinaryFormat: true,
  })
}

// ── Single-flight refresh state ──

let refreshPromise: Promise<string | null> | null = null

async function doRefresh(): Promise<string | null> {
  const rt = memRefreshToken
  if (!rt) return null
  if (refreshPromise) return refreshPromise

  refreshPromise = (async () => {
    try {
      const client = createClient(AuthService, noAuthTransport())
      const res = await client.refresh(create(RefreshRequestSchema, { refreshToken: rt }))
      if (res.accessToken) {
        memAccessToken = res.accessToken
        if (res.refreshToken) memRefreshToken = res.refreshToken
        return res.accessToken
      }
    } catch {
      // refresh failed — clear in-memory state
      memAccessToken = null
      memRefreshToken = null
      memUser = null
    } finally {
      refreshPromise = null
    }
    return null
  })()

  return refreshPromise
}

// ── Provider ──

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(memUser)
  const [isLoading, setIsLoading] = useState(false)

  const login = useCallback(async (email: string, password: string) => {
    setIsLoading(true)
    try {
      const clientInfo = create(ClientInfoSchema, {
        userAgent: navigator.userAgent,
        ipAddress: '127.0.0.1',
      })
      const transport = noAuthTransport()
      const client = createClient(AuthService, transport)
      const res = await client.login(
        create(LoginRequestSchema, { email, password, client: clientInfo }),
      )

      if (!res.accessToken) throw new Error('No access token in login response')

      // Store in memory (NOT localStorage)
      memAccessToken = res.accessToken
      memRefreshToken = res.refreshToken ?? null

      // Parse user info from response
      const user: CurrentUser = {
        userId: res.userId ?? '',
        email,
        roles: (res as any).roles ?? ['admin'],
        permissions: (res as any).permissions ?? [],
      }
      memUser = user
      setCurrentUser(user)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const logout = useCallback(() => {
    memAccessToken = null
    memRefreshToken = null
    memUser = null
    setCurrentUser(null)
  }, [])

  const hasPermission = useCallback(
    (perm: string): boolean => {
      if (!currentUser) return false
      return currentUser.permissions.includes(perm)
    },
    [currentUser],
  )

  const requirePermission = useCallback(
    (perm: string): boolean => {
      return hasPermission(perm)
    },
    [hasPermission],
  )

  const doRefreshToken = useCallback(async () => {
    const tok = await doRefresh()
    if (tok && memUser) {
      setCurrentUser({ ...memUser })
    }
  }, [])

  const getAccessToken = useCallback(() => memAccessToken, [])
  const getRefreshToken = useCallback(() => memRefreshToken, [])

  // On mount, if we have a stored refresh token, attempt a silent refresh
  useEffect(() => {
    const rt = memRefreshToken
    if (rt && !memAccessToken) {
      doRefresh().then((tok) => {
        if (tok && memUser) setCurrentUser({ ...memUser })
      })
    }
  }, [])

  const value = useMemo<AuthState>(
    () => ({
      currentUser,
      isAuthenticated: !!currentUser && !!memAccessToken,
      isLoading,
      login,
      logout,
      hasPermission,
      requirePermission,
      refreshToken: doRefreshToken,
      getAccessToken,
      getRefreshToken,
    }),
    [
      currentUser,
      isLoading,
      login,
      logout,
      hasPermission,
      requirePermission,
      doRefreshToken,
      getAccessToken,
      getRefreshToken,
    ],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// ── Hook ──

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

// ── Exported for transport interceptor ──

export function getMemAccessToken(): string | null {
  return memAccessToken
}

export function getMemRefreshToken(): string | null {
  return memRefreshToken
}

export function clearMemSession() {
  memAccessToken = null
  memRefreshToken = null
  memUser = null
}

export { doRefresh as refreshTokenFromMemory }
