import { useState } from 'react'
import { AlertTriangle, X, Loader2 } from 'lucide-react'

export interface DangerConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: (reason: string) => Promise<void>
  title: string
  description: string
  targetName: string
  confirmLabel?: string
  requireReason?: boolean
}

export default function DangerConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  description,
  targetName,
  confirmLabel = '确认执行',
  requireReason = true,
}: DangerConfirmDialogProps) {
  const [reason, setReason] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!open) return null

  const reasonValid = !requireReason || reason.trim().length >= 10
  const canSubmit = reasonValid && !loading

  const handleConfirm = async () => {
    if (!canSubmit) return
    setLoading(true)
    setError(null)
    try {
      await onConfirm(reason.trim())
      setReason('')
      setError(null)
      onClose()
    } catch (e: any) {
      setError(e?.message || '操作失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    if (!loading) {
      setReason('')
      setError(null)
      onClose()
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={handleClose}>
      <div className="bg-white rounded-xl p-6 w-[30rem] shadow-xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center gap-3 mb-4">
          <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center flex-shrink-0">
            <AlertTriangle className="w-5 h-5 text-red-600" />
          </div>
          <div className="flex-1">
            <h2 className="text-lg font-bold text-gray-900">{title}</h2>
          </div>
          <button
            onClick={handleClose}
            disabled={loading}
            className="p-1 rounded hover:bg-gray-100 disabled:opacity-50 flex-shrink-0"
          >
            <X className="w-5 h-5 text-gray-400" />
          </button>
        </div>

        {/* Description */}
        <p className="text-sm text-gray-600 mb-3">{description}</p>

        {/* Target */}
        <div className="bg-gray-50 rounded-lg px-3 py-2 mb-4">
          <span className="text-xs text-gray-400">目标</span>
          <p className="text-sm font-mono text-gray-800 break-all">{targetName}</p>
        </div>

        {/* Reason input */}
        {requireReason && (
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              操作原因 <span className="text-red-500">*</span>
            </label>
            <textarea
              value={reason}
              onChange={(e) => { setReason(e.target.value); setError(null) }}
              placeholder="请填写操作原因（至少10个字符）..."
              rows={3}
              maxLength={500}
              className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-red-500 focus:border-red-500 outline-none resize-none"
              disabled={loading}
            />
            <p className={`text-xs mt-1 ${reason.trim().length >= 10 ? 'text-green-600' : 'text-gray-400'}`}>
              {reason.trim().length}/10 最小字符
            </p>
          </div>
        )}

        {/* Error */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg px-3 py-2 mb-4">
            <p className="text-sm text-red-700">{error}</p>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-3 justify-end">
          <button
            onClick={handleClose}
            disabled={loading}
            className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          >
            取消
          </button>
          <button
            onClick={handleConfirm}
            disabled={!canSubmit}
            className={`flex items-center gap-2 px-4 py-2 text-sm rounded-lg font-medium transition-colors ${
              canSubmit
                ? 'bg-red-600 text-white hover:bg-red-700'
                : 'bg-gray-200 text-gray-400 cursor-not-allowed'
            }`}
          >
            {loading && <Loader2 className="w-4 h-4 animate-spin" />}
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
