// AsyncView 统一渲染 RPC 调用的 idle/loading/error/success 四态。
// 所有 features 页面用 useAsync(loader) + <AsyncView state={s} render={...} /> 包裹。
import { useEffect, useState, useCallback, ReactNode } from 'react'

export type AsyncState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: string }

export function useAsync<T>(loader: () => Promise<T>, deps: unknown[] = []) {
  const [state, setState] = useState<AsyncState<T>>({ status: 'idle' })
  const reload = useCallback(async () => {
    setState({ status: 'loading' })
    try {
      const data = await loader()
      setState({ status: 'success', data })
    } catch (e: any) {
      setState({ status: 'error', error: e?.message || String(e) })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)
  useEffect(() => {
    reload()
  }, [reload])
  return { state, reload }
}

export function AsyncView<T>(props: {
  state: AsyncState<T>
  render: (data: T) => ReactNode
  emptyText?: string
}) {
  const { state, render, emptyText = '暂无数据' } = props
  if (state.status === 'idle' || state.status === 'loading') {
    return <div className="p-8 text-gray-400 animate-pulse">加载中...</div>
  }
  if (state.status === 'error') {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded text-red-700">
        加载失败：{state.error}
      </div>
    )
  }
  const data = state.data
  if (
    data == null ||
    (Array.isArray(data) && data.length === 0)
  ) {
    return <div className="p-8 text-gray-400">{emptyText}</div>
  }
  return <>{render(data)}</>
}

// PageShell 统一页面容器（标题 + 副标题 + 内容卡片）。
export function PageShell(props: { title: string; subtitle?: ReactNode; actions?: ReactNode; children: ReactNode }) {
  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{props.title}</h1>
          {props.subtitle && <p className="text-sm text-gray-500 mt-1">{props.subtitle}</p>}
        </div>
        {props.actions}
      </div>
      <div className="bg-white rounded-xl shadow-sm p-6">{props.children}</div>
    </div>
  )
}
