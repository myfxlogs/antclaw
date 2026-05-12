import { AlertCircle } from 'lucide-react'
import { Alert, Tooltip } from 'antd'

export function Section({
  step,
  title,
  subtitle,
  children,
}: {
  step?: number
  title: string
  subtitle?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="bg-white rounded-xl shadow-sm border border-gray-100">
      <header className="flex items-start justify-between gap-4 px-6 py-4">
        <div className="flex items-start gap-3">
          {typeof step === 'number' && (
            <span className="w-7 h-7 rounded-full bg-blue-600 text-white text-sm font-semibold flex items-center justify-center shrink-0">
              {step}
            </span>
          )}
          <div>
            <h2 className="text-base font-semibold text-gray-900">{title}</h2>
            {subtitle && <div className="text-xs text-gray-500 mt-0.5">{subtitle}</div>}
          </div>
        </div>
      </header>
      <div className="px-6 pb-6 pt-2">{children}</div>
    </section>
  )
}

export function Label({
  text,
  hint,
  badge,
}: {
  text: string
  hint?: string
  badge?: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between mb-1.5">
      <div className="flex items-center gap-1.5">
        <span className="text-sm font-medium text-gray-700">{text}</span>
        {hint && (
          <Tooltip title={hint}>
            <AlertCircle className="w-3.5 h-3.5 text-gray-400 cursor-help" />
          </Tooltip>
        )}
      </div>
      {badge}
    </div>
  )
}

export function StatusBanner({
  tone,
  title,
  description,
  notice,
}: {
  tone: 'success' | 'warning' | 'error' | 'info'
  title: string
  description: string
  notice?: string
}) {
  return (
    <Alert
      type={tone}
      showIcon
      message={<span className="font-semibold">{title}</span>}
      description={
        <div className="space-y-1">
          <div>{description}</div>
          {notice && <div className="text-xs text-green-700">{notice}</div>}
        </div>
      }
    />
  )
}
