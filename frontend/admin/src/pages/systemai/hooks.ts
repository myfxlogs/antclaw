import { useEffect, useMemo, useRef, useState } from 'react'
import {
  clearSystemAISecret,
  discoverSystemAIModels,
  listSystemAIConfigs,
  updateSystemAIConfig,
  updateSystemAISecret,
  validateSystemAI,
} from '../../lib/api'
import { OFFICIAL_PROVIDER_BASE_URLS, toFriendlyDiscoverMessage } from './constants'
import type { AIConfig } from './model'

export function useSystemAIPage() {
  const [configs, setConfigs] = useState<AIConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [savingConfig, setSavingConfig] = useState(false)
  const [savingSecret, setSavingSecret] = useState(false)
  const [selectedProviderId, setSelectedProviderId] = useState('')
  const [draft, setDraft] = useState<AIConfig | null>(null)
  const [secretInput, setSecretInput] = useState('')
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const [validated, setValidated] = useState(false)
  const [validating, setValidating] = useState(false)
  const [discovering, setDiscovering] = useState(false)
  const [lastAutoDiscoverKey, setLastAutoDiscoverKey] = useState('')
  const [lastAutoSavedSecretKey, setLastAutoSavedSecretKey] = useState('')

  const validateBaseURL = (value: string): string | null => {
    const input = value.trim()
    if (!input) return '请先填写 Base URL（模型服务地址）。'
    let parsed: URL
    try {
      parsed = new URL(input)
    } catch {
      return 'base url format invalid'
    }
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
      return 'base url format invalid'
    }
    return null
  }

  const persistDraftConfig = async (cfg: AIConfig) => {
    await updateSystemAIConfig(cfg.provider_id, {
      name: cfg.name,
      base_url: cfg.base_url,
      organization: cfg.organization,
      models: cfg.models,
      default_model: cfg.default_model,
      temperature: cfg.temperature,
      timeout_seconds: cfg.timeout_seconds,
      max_tokens: cfg.max_tokens,
      purposes: cfg.purposes,
      primary_for: cfg.primary_for,
      enabled: cfg.enabled,
    })
  }

  const fetchConfigs = async (): Promise<AIConfig[]> => {
    const json = await listSystemAIConfigs()
    const items = (json.items || []) as AIConfig[]
    setConfigs(items)
    return items
  }

  const load = async () => {
    setLoading(true)
    try {
      await fetchConfigs()
    } catch (err) {
      console.error('failed to load ai configs', err)
      setError('加载配置失败')
    } finally {
      setLoading(false)
    }
  }

  // 静默刷新：不触发全局 loading，仅同步最新 configs，用于密钥保存等背景操作
  const silentReload = async () => {
    try {
      await fetchConfigs()
    } catch (err) {
      console.error('failed to silently reload ai configs', err)
    }
  }

  useEffect(() => { load() }, [])

  const selectedConfig = useMemo(
    () => configs.find((c) => c.provider_id === selectedProviderId) || null,
    [configs, selectedProviderId],
  )

  // 记录上一次渲染所选 provider_id，仅在真正切换 provider 时重置 UI 输入状态；
  // configs 因后台刷新导致 selectedConfig 引用变化时，保留用户正在输入的 secretInput / notice 等。
  const prevProviderIdRef = useRef<string>('')
  useEffect(() => {
    const nextId = selectedConfig?.provider_id || ''
    const providerChanged = nextId !== prevProviderIdRef.current
    prevProviderIdRef.current = nextId

    if (!selectedConfig) {
      setDraft(null)
    } else if (providerChanged) {
      const fixedBase = OFFICIAL_PROVIDER_BASE_URLS[selectedConfig.provider_id]
      const enforcedBase = selectedConfig.provider_id === 'openai_compatible'
        ? (selectedConfig.base_url || '')
        : (fixedBase || '')
      setDraft({
        ...selectedConfig,
        base_url: enforcedBase,
      })
    } else {
      // provider 未变化：只补齐服务端权威字段（has_secret / updated_at / models），避免覆盖用户本地编辑
      setDraft((prev) => (prev ? {
        ...prev,
        has_secret: selectedConfig.has_secret,
        updated_at: selectedConfig.updated_at,
        models: prev.models && prev.models.length > 0 ? prev.models : selectedConfig.models,
      } : prev))
    }

    if (providerChanged) {
      setSecretInput('')
      setNotice('')
      setError('')
      setValidated(false)
      setLastAutoSavedSecretKey('')
    }
  }, [selectedConfig])

  useEffect(() => {
    if (!draft) return
    const secret = secretInput.trim()
    if (!secret) return
    const key = `${draft.provider_id}|${secret}`
    if (key === lastAutoSavedSecretKey) return

    const timer = setTimeout(async () => {
      setSavingSecret(true)
      try {
        await updateSystemAISecret(draft.provider_id, secret)
        setLastAutoSavedSecretKey(key)
        setError('')
        setValidated(false)
        // 本地即时标记密钥已配置，触发模型自动发现；不再 load() 造成全局刷新
        setDraft((prev) => prev ? { ...prev, has_secret: true } : prev)
        setLastAutoDiscoverKey('')
        setNotice('密钥已保存，正在自动发现模型...')
        // 后台静默同步 configs，保持卡片状态一致（不触发 loading 骨架）
        void silentReload()
      } catch (e) {
        const msg = e instanceof Error ? e.message : '密钥自动保存失败'
        setError(msg)
      } finally {
        setSavingSecret(false)
      }
    }, 700)
    return () => clearTimeout(timer)
  }, [draft?.provider_id, secretInput, lastAutoSavedSecretKey])

  useEffect(() => {
    if (!draft) return
    const base = (draft.base_url || '').trim()
    if (!base) return
    // 只有在 key 已经可用时才拉取模型。拉取 /models 基本都需要鉴权，
    // 预取只会返回 401/403/failed_precondition 等无用噪音。
    if (!draft.has_secret && !secretInput.trim()) return
    const key = `${draft.provider_id}|${base}|${draft.has_secret ? 'saved' : 'pending'}`
    if (key === lastAutoDiscoverKey) return

    const timer = setTimeout(async () => {
      setDiscovering(true)
      try {
        const baseError = validateBaseURL(base)
        if (baseError) {
          setError(toFriendlyDiscoverMessage(baseError))
          return
        }
        await persistDraftConfig(draft)
        const body = await discoverSystemAIModels(draft.provider_id)
        const models = (body?.models || []) as string[]
        if (models.length > 0) {
          setDraft((prev) => prev ? {
            ...prev,
            models,
            default_model: body?.default_model || models[0],
          } : prev)
          setNotice(`已自动发现 ${models.length} 个模型`)
          setError('')
          setValidated(false)
          setLastAutoDiscoverKey(key)
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : '拉取模型失败'
        setError(toFriendlyDiscoverMessage(msg))
      } finally {
        setDiscovering(false)
      }
    }, 700)

    return () => clearTimeout(timer)
  }, [draft?.provider_id, draft?.base_url, draft?.has_secret, secretInput, lastAutoDiscoverKey])

  // 选择默认模型后自动验证连接（取消手动「验证」按钮）。
  // 仅在已有可用模型、已配置密钥、且 default_model 变化时触发；
  // 用 key 去重避免重复验证同一模型；失败时仅置错误，不强制重试。
  const [lastAutoValidateKey, setLastAutoValidateKey] = useState('')
  useEffect(() => {
    if (!draft) return
    const model = (draft.default_model || '').trim()
    if (!model) return
    if (!(draft.models && draft.models.length > 0)) return
    if (!(draft.has_secret || secretInput.trim())) return
    const baseError = validateBaseURL(draft.base_url)
    if (baseError) return
    const key = `${draft.provider_id}|${draft.base_url}|${model}`
    if (key === lastAutoValidateKey) return

    const timer = setTimeout(async () => {
      setLastAutoValidateKey(key)
      setValidating(true)
      try {
        await persistDraftConfig(draft)
        if (secretInput.trim()) {
          await updateSystemAISecret(draft.provider_id, secretInput.trim())
        }
        const body = await validateSystemAI(draft.provider_id)
        setValidated(true)
        setNotice(`已自动验证：发现 ${body.model_count ?? 0} 个模型`)
        setError('')
      } catch (e) {
        const msg = e instanceof Error ? e.message : '自动验证失败'
        setValidated(false)
        setError(toFriendlyDiscoverMessage(msg))
      } finally {
        setValidating(false)
      }
    }, 500)
    return () => clearTimeout(timer)
  }, [draft?.provider_id, draft?.base_url, draft?.default_model, draft?.has_secret, secretInput, lastAutoValidateKey])

  const saveConfig = async () => {
    if (!draft) return
    setSavingConfig(true)
    try {
      await persistDraftConfig(draft)
      setNotice('配置已保存')
      setError('')
      void silentReload()
    } finally {
      setSavingConfig(false)
    }
  }

  const clearSecret = async () => {
    if (!draft) return
    setSavingSecret(true)
    try {
      await clearSystemAISecret(draft.provider_id)
      const resetBaseURL = OFFICIAL_PROVIDER_BASE_URLS[draft.provider_id] || ''
      await updateSystemAIConfig(draft.provider_id, {
        name: draft.name,
        base_url: resetBaseURL,
        organization: '',
        models: [],
        default_model: '',
        temperature: 0.2,
        timeout_seconds: 60,
        max_tokens: 4096,
        purposes: draft.purposes || [],
        primary_for: [],
        enabled: false,
      })
      setSecretInput('')
      setLastAutoSavedSecretKey('')
      setLastAutoDiscoverKey('')
      setDraft((prev) => prev ? {
        ...prev,
        base_url: resetBaseURL,
        organization: '',
        models: [],
        default_model: '',
        temperature: 0.2,
        timeout_seconds: 60,
        max_tokens: 4096,
        primary_for: [],
        enabled: false,
        has_secret: false,
      } : prev)
      setNotice('密钥已删除，厂商配置已恢复默认初始化')
      setError('')
      setValidated(false)
      void silentReload()
    } finally {
      setSavingSecret(false)
    }
  }

  const validateConnection = async () => {
    if (!draft) return
    setValidating(true)
    try {
      const baseError = validateBaseURL(draft.base_url)
      if (baseError) {
        setValidated(false)
        setError(toFriendlyDiscoverMessage(baseError))
        return
      }
      await persistDraftConfig(draft)
      if (secretInput.trim()) {
        await updateSystemAISecret(draft.provider_id, secretInput.trim())
      }
      const body = await validateSystemAI(draft.provider_id)
      setValidated(true)
      setNotice(`验证通过：发现 ${body.model_count ?? 0} 个模型`)
      setError('')
    } catch (e) {
      const msg = e instanceof Error ? e.message : '验证失败'
      setValidated(false)
      if (msg.includes('401/403') && !draft.has_secret && !secretInput.trim()) {
        setError('验证失败：当前厂商通常需要 API Key。请先填写并保存密钥，再重试验证连接。')
      } else {
        setError(toFriendlyDiscoverMessage(msg))
      }
    } finally {
      setValidating(false)
    }
  }

  return {
    configs,
    loading,
    savingConfig,
    savingSecret,
    selectedProviderId,
    setSelectedProviderId,
    draft,
    setDraft,
    secretInput,
    setSecretInput,
    notice,
    error,
    validated,
    setValidated,
    validating,
    discovering,
    setLastAutoDiscoverKey,
    load,
    saveConfig,
    clearSecret,
    validateConnection,
  }
}
