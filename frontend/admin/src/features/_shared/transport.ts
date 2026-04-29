// 共享 Connect 传输；所有 features 模块通过本文件构造 service client，避免重复 transport。
import { createConnectTransport } from '@connectrpc/connect-web'

const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8082'

export const transport = createConnectTransport({
  baseUrl: API_BASE_URL,
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem('token')
      if (token) req.header.set('Authorization', `Bearer ${token}`)
      return await next(req)
    },
  ],
})
