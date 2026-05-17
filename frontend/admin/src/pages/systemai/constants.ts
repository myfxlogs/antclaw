import type { TFunction } from 'i18next'

export const PROVIDER_LINKS: Record<string, string> = {
  openai: 'https://platform.openai.com/api-keys',
  openai_compatible: 'https://platform.openai.com/docs/api-reference/introduction',
  anthropic: 'https://console.anthropic.com/settings/keys',
  deepseek: 'https://platform.deepseek.com/api_keys',
  moonshot: 'https://platform.moonshot.cn/console/api-keys',
  qwen: 'https://bailian.console.aliyun.com/?apiKey=1',
  zhipu: 'https://open.bigmodel.cn/usercenter/apikeys',
  gemini: 'https://aistudio.google.com/apikey',
}

export const ALL_PURPOSES = ['chat', 'embedding', 'summarizer', 'reasoning']

export const OFFICIAL_PROVIDER_BASE_URLS: Record<string, string> = {
  openai: 'https://api.openai.com/v1',
  anthropic: 'https://api.anthropic.com/v1',
  deepseek: 'https://api.deepseek.com/v1',
  moonshot: 'https://api.moonshot.cn/v1',
  qwen: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  zhipu: 'https://open.bigmodel.cn/api/paas/v4',
  // Gemini 通过 OpenAI 兼容端点暴露：https://ai.google.dev/gemini-api/docs/openai
  gemini: 'https://generativelanguage.googleapis.com/v1beta/openai',
}

export function toFriendlyDiscoverMessage(msg: string, t: TFunction) {
  const lower = msg.toLowerCase()
  if (msg.includes('base_url')) return t('systemai.baseUrlRequired')
  if (msg.includes('base url format invalid')) return t('systemai.baseUrlInvalid')
  // 借鉴 anttrader：根据厂商真实错误识别免费档耗尽 / 配额限流
  if (lower.includes('free-tier exhausted') || lower.includes('freetieronly') || lower.includes('free tier') || lower.includes('free-tier only')) {
    return t('systemai.freeTierExhausted')
  }
  if (lower.includes('quota exhausted') || lower.includes('[resource_exhausted]') || lower.includes('status 429') || lower.includes('too many requests') || lower.includes('rate limit')) {
    return t('systemai.quotaExhausted')
  }
  if (lower.includes('status 403') && (lower.includes('quota') || lower.includes('exhaust') || lower.includes('allocation'))) {
    return t('systemai.quota403')
  }
  if (msg.includes('unauthorized')) return t('systemai.unauthorized')
  if (msg.includes('endpoint')) return t('systemai.endpointNotFound')
  if (msg.includes('timeout')) return t('systemai.timeout')
  if (msg.includes('unreachable')) return t('systemai.unreachable')
  if (msg.includes('invalid /models')) return t('systemai.invalidModels')
  if (msg.includes('no models returned')) return t('systemai.noModelsReturned')
  return msg || t('systemai.defaultFetchError')
}
