// DataTable — generic server-paginated table with loading/error/empty states (A13-P1-03)
import { ReactNode } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { ErrorState, LoadingSkeleton, EmptyState } from './Common'

export interface Column<T> {
  key: string
  header: string
  render?: (row: T) => ReactNode
  className?: string
}

interface DataTableProps<T> {
  columns: Column<T>[]
  rows: T[]
  totalCount: number
  pageSize?: number
  currentPage: number
  onPageChange: (page: number) => void
  loading?: boolean
  error?: string | null
  onRetry?: () => void
  emptyText?: string
  getRowKey?: (row: T) => string
}

export default function DataTable<T extends Record<string, any>>({
  columns, rows, totalCount, pageSize = 20, currentPage, onPageChange,
  loading, error, onRetry, emptyText = 'No data', getRowKey,
}: DataTableProps<T>) {
  if (error) return <ErrorState message={error} onRetry={onRetry} />
  if (loading) return <LoadingSkeleton rows={8} />
  if (rows.length === 0) return <EmptyState title={emptyText} />

  const totalPages = Math.ceil(totalCount / pageSize)
  const from = (currentPage - 1) * pageSize + 1
  const to = Math.min(currentPage * pageSize, totalCount)

  return (
    <div className="bg-white rounded-xl shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-gray-50">
              {columns.map((col) => (
                <th key={col.key} className={`px-4 py-3 text-left font-medium text-gray-600 ${col.className || ''}`}>
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={getRowKey ? getRowKey(row) : i} className="border-b hover:bg-gray-50 transition-colors">
                {columns.map((col) => (
                  <td key={col.key} className={`px-4 py-3 text-gray-700 ${col.className || ''}`}>
                    {col.render ? col.render(row) : row[col.key]}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-4 py-3 border-t bg-gray-50 text-sm text-gray-600">
          <span>
            {from}–{to} of {totalCount}
          </span>
          <div className="flex items-center gap-1">
            <button
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage <= 1}
              className="p-1.5 rounded hover:bg-gray-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              aria-label="Previous page"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
              let pg: number
              if (totalPages <= 7) {
                pg = i + 1
              } else if (currentPage <= 4) {
                pg = i + 1
              } else if (currentPage >= totalPages - 3) {
                pg = totalPages - 6 + i
              } else {
                pg = currentPage - 3 + i
              }
              if (pg < 1 || pg > totalPages) return null
              return (
                <button
                  key={pg}
                  onClick={() => onPageChange(pg)}
                  className={`w-8 h-8 rounded text-xs font-medium transition-colors ${
                    pg === currentPage ? 'bg-blue-600 text-white' : 'hover:bg-gray-200'
                  }`}
                >
                  {pg}
                </button>
              )
            })}
            <button
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage >= totalPages}
              className="p-1.5 rounded hover:bg-gray-200 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
              aria-label="Next page"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
