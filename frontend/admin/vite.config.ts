import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), 'VITE_')
  const isProduction = mode === 'production'

  // P2-04: production build must have API base URL
  if (isProduction && !env.VITE_API_BASE_URL) {
    throw new Error(
      'VITE_API_BASE_URL is required for production builds. ' +
      'Set it in .env.production or via environment variable. ' +
      'Example: VITE_API_BASE_URL=https://api.antclaw.example.com'
    )
  }

  return {
    plugins: [react()],
    server: { port: 3001 },
    build: {
      outDir: 'dist',
      sourcemap: isProduction ? false : true,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes('@connectrpc') || id.includes('@bufbuild')) return 'connect'
            if (id.includes('lucide-react')) return 'icons'
          },
        },
      },
    },
    define: {
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
      __API_BASE__: JSON.stringify(env.VITE_API_BASE_URL || ''),
      __APP_VERSION__: JSON.stringify(process.env.npm_package_version || '0.1.0'),
    },
  }
})
