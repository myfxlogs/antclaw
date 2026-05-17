// Common UI components for admin panel (A13-P1-06)

import { ReactNode } from 'react'
import { AlertTriangle, RefreshCw, AlertCircle } from 'lucide-react'

/* ── ErrorState ── */
export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center p-8 bg-red-50 border border-red-200 rounded-xl text-center space-y-3">
      <AlertCircle className="w-8 h-8 text-red-500" />
      <p className="text-red-700 font-medium">{message}</p>
      {onRetry && (
        <button onClick={onRetry} className="flex items-center gap-1.5 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors text-sm">
          <RefreshCw className="w-4 h-4" />
          Retry
        </button>
      )}
    </div>
  )
}

/* ── EmptyState ── */
export function EmptyState({ icon, title, description }: { icon?: ReactNode; title: string; description?: string }) {
  return (
    <div className="flex flex-col items-center justify-center p-12 text-center space-y-3">
      {icon && <div className="text-gray-300">{icon}</div>}
      <p className="text-gray-500 font-medium">{title}</p>
      {description && <p className="text-sm text-gray-400">{description}</p>}
    </div>
  )
}

/* ── ConfirmDialog ── */
interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
}

export function ConfirmDialog({ open, onClose, onConfirm, title, message, confirmLabel = 'Confirm', danger }: ConfirmDialogProps) {
  if (!open) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={onClose}>
      <div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md space-y-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-start gap-3">
          {danger && <AlertTriangle className="w-6 h-6 text-red-500 flex-shrink-0 mt-0.5" />}
          <div>
            <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
            <p className="text-sm text-gray-600 mt-1">{message}</p>
          </div>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
            Cancel
          </button>
          <button
            onClick={() => { onConfirm(); onClose() }}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
              danger ? 'bg-red-600 text-white hover:bg-red-700' : 'bg-blue-600 text-white hover:bg-blue-700'
            }`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}

// DangerConfirmDialog moved to components/DangerConfirmDialog.tsx
export { default as DangerConfirmDialog } from './DangerConfirmDialog'
export type { DangerConfirmDialogProps } from './DangerConfirmDialog'

/* ── Badge ── */
export function Badge({ children, variant = 'default' }: { children: ReactNode; variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' }) {
  const colors: Record<string, string> = {
    default: 'bg-gray-100 text-gray-700',
    success: 'bg-green-100 text-green-700',
    warning: 'bg-yellow-100 text-yellow-700',
    danger: 'bg-red-100 text-red-700',
    info: 'bg-blue-100 text-blue-700',
  }
  return <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${colors[variant]}`}>{children}</span>
}

/* ── PageHeader ── */
export function PageHeader({ title, subtitle, actions }: { title: string; subtitle?: string; actions?: ReactNode }) {
  return (
    <div className="flex items-start justify-between mb-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}

/* ── LoadingSkeleton ── */
export function LoadingSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3 p-4">
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="h-10 bg-gray-100 rounded animate-pulse" style={{ width: `${70 + Math.random() * 30}%` }} />
      ))}
    </div>
  )
}

/* ── StreamingIndicator ── */
export function StreamingIndicator({ active }: { active: boolean }) {
  if (!active) return null
  return (
    <div className="flex items-center gap-1.5 px-2 py-1 bg-green-50 text-green-700 rounded text-xs">
      <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
      Live
    </div>
  )
}