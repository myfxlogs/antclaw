import { createConnectTransport } from '@connectrpc/connect-web'

const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8080'

export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  useBinaryFormat: true,
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem('token')
      if (token) req.header.set('Authorization', `Bearer ${token}`)
      return await next(req)
    },
  ],
})
