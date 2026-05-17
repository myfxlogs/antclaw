// /ai/chat —— 通过 AIService.RunWithTools 与配置好的 LLM 提供商对话。
//
// 功能：
//   - 启动时拉取 SystemAI 配置列表，构建 provider/model 选择器（仅展示已配密钥的）
//   - 顶部展示当前选中的 provider/model
//   - 每条 AI 回复显示本轮 prompt/completion/total token 与累计消耗
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AIService, RunWithToolsRequestSchema } from '@antclaw/proto/antclaw/v1/ai_pb'
import {
  SystemAIService,
  ListSystemAIConfigsRequestSchema,
} from '@antclaw/proto/antclaw/v1/system_ai_pb'
import { transport } from '../../_shared/transport'
import { PageShell } from '../../_shared/AsyncView'

const aiClient = createClient(AIService, transport)
const sysAIClient = createClient(SystemAIService, transport)

interface Turn {
  role: 'user' | 'assistant'
  text: string
  tools?: { name: string; result?: string }[]
  model?: string
  providerId?: string
  promptTokens?: number
  completionTokens?: number
  totalTokens?: number
}

interface ProviderOption {
  providerId: string
  name: string
  defaultModel: string
  models: string[]
}

export default function ChatPage() {
  const { t } = useTranslation()
  const [userId] = useState<string>(() => localStorage.getItem('user_id') || 'admin')
  const [threadId, setThreadId] = useState<string>('')
  const [input, setInput] = useState('')
  const [turns, setTurns] = useState<Turn[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  // Provider / model 选择
  const [providers, setProviders] = useState<ProviderOption[]>([])
  const [providerId, setProviderId] = useState<string>('')
  const [model, setModel] = useState<string>('')

  useEffect(() => {
    let alive = true
    sysAIClient
      .listConfigs(create(ListSystemAIConfigsRequestSchema, {}))
      .then((res) => {
        if (!alive) return
        const opts: ProviderOption[] = (res.items || [])
          .filter((c: any) => c.enabled && c.hasSecret)
          .map((c: any) => ({
            providerId: c.providerId,
            name: c.name || c.providerId,
            defaultModel: c.defaultModel || '',
            models: c.models || [],
          }))
        setProviders(opts)
        // 优先选 default for chat（primary_for 含 chat），其次第一项
        const pickFirst =
          (res.items || []).find(
            (c: any) =>
              c.enabled && c.hasSecret && (c.primaryFor || []).includes('chat'),
          ) || (res.items || []).find((c: any) => c.enabled && c.hasSecret)
        if (pickFirst) {
          setProviderId(pickFirst.providerId)
          setModel(pickFirst.defaultModel || (pickFirst.models?.[0] ?? ''))
        }
      })
      .catch((e) => setErr(t('ai.loadProviderError', { message: e?.message || String(e) })))
    return () => {
      alive = false
    }
  }, [])

  // 切换 provider 时刷新 model
  useEffect(() => {
    const p = providers.find((o) => o.providerId === providerId)
    if (!p) return
    if (!p.models.includes(model)) {
      setModel(p.defaultModel || p.models[0] || '')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [providerId])

  const currentProvider = providers.find((o) => o.providerId === providerId)
  const cumulativeTokens = useMemo(
    () => turns.reduce((s, t) => s + (t.totalTokens || 0), 0),
    [turns],
  )

  const send = async () => {
    if (!input.trim() || busy) return
    const text = input.trim()
    setInput('')
    setBusy(true)
    setErr(null)
    setTurns((t) => [...t, { role: 'user', text }])
    try {
      const r = await aiClient.runWithTools(
        create(RunWithToolsRequestSchema, {
          userId,
          threadId,
          message: text,
          maxHops: 5,
          model,
          providerId,
        } as any),
      )
      if (r.threadId && !threadId) setThreadId(r.threadId)
      setTurns((t) => [
        ...t,
        {
          role: 'assistant',
          text: r.answer || '(空)',
          tools: (r.calls || []).map((c) => ({
            name: c.name,
            result: c.resultJson || c.error,
          })),
          model: r.model,
          providerId: r.providerId,
          promptTokens: r.promptTokens,
          completionTokens: r.completionTokens,
          totalTokens: r.totalTokens,
        },
      ])
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <PageShell
      title="AI 对话"
      subtitle={
        <span className="text-xs text-gray-500">
          thread=<span className="font-mono">{threadId || '(未创建)'}</span> · user=
          <span className="font-mono">{userId}</span> · 累计 tokens=
          <span className="font-mono">{cumulativeTokens}</span>
        </span>
      }
      actions={
        <div className="flex flex-wrap gap-2 items-end">
          <label className="text-sm">
            <span className="text-xs text-gray-500 block">Provider</span>
            <select
              className="input"
              value={providerId}
              onChange={(e) => setProviderId(e.target.value)}
              disabled={providers.length === 0}
            >
              {providers.length === 0 && <option value="">（未配置）</option>}
              {providers.map((p) => (
                <option key={p.providerId} value={p.providerId}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <label className="text-sm">
            <span className="text-xs text-gray-500 block">Model</span>
            <input
              className="input w-56"
              list="chat-models"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              placeholder="model-id"
            />
            <datalist id="chat-models">
              {(currentProvider?.models || []).map((m) => (
                <option key={m} value={m} />
              ))}
            </datalist>
          </label>
        </div>
      }
    >
      <div className="flex flex-col h-[60vh]">
        <div className="flex-1 overflow-auto space-y-3 mb-3">
          {turns.length === 0 && (
            <div className="text-sm text-gray-400">输入消息开始对话</div>
          )}
          {turns.map((t, i) => (
            <div
              key={i}
              className={`p-3 rounded ${t.role === 'user' ? 'bg-blue-50' : 'bg-gray-50'}`}
            >
              <div className="text-xs text-gray-500 mb-1 flex items-center gap-2">
                <span>{t.role === 'user' ? '我' : 'AI'}</span>
                {t.role === 'assistant' && t.model && (
                  <span className="px-1.5 py-0.5 rounded bg-gray-200 text-gray-700 font-mono text-[10px]">
                    {t.providerId ? `${t.providerId}/` : ''}
                    {t.model}
                  </span>
                )}
                {t.role === 'assistant' && (t.totalTokens || 0) > 0 && (
                  <span className="text-[10px] text-gray-500 font-mono">
                    tokens: {t.promptTokens || 0}+{t.completionTokens || 0}={t.totalTokens || 0}
                  </span>
                )}
              </div>
              <div className="whitespace-pre-wrap text-sm">{t.text}</div>
              {t.tools && t.tools.length > 0 && (
                <div className="mt-2 text-xs text-gray-500">
                  工具调用：
                  {t.tools.map((tc, j) => (
                    <span
                      key={j}
                      className="ml-1 px-2 py-0.5 bg-purple-100 text-purple-700 rounded"
                    >
                      {tc.name}
                    </span>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
        {err && (
          <div className="p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700 mb-2">
            {err}
          </div>
        )}
        <div className="flex gap-2">
          <input
            className="input flex-1"
            placeholder={providerId ? '问点什么...' : '请先在 SystemAI 配置 Provider 和 API Key'}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                send()
              }
            }}
            disabled={!providerId}
          />
          <button
            onClick={send}
            disabled={busy || !providerId || !input.trim()}
            className="px-4 py-2 bg-blue-600 text-white rounded disabled:opacity-50"
          >
            发送
          </button>
        </div>
      </div>
    </PageShell>
  )
}
