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

export function toFriendlyDiscoverMessage(msg: string) {
  const lower = msg.toLowerCase()
  if (msg.includes('base_url')) return '请先填写 Base URL（模型服务地址）。'
  if (msg.includes('base url format invalid')) return 'Base URL 格式无效：请填写完整地址，例如 https://model.myfxlogs.org 或 https://model.myfxlogs.org/v1'
  // 借鉴 anttrader：根据厂商真实错误识别免费档耗尽 / 配额限流
  if (lower.includes('free-tier exhausted') || lower.includes('freetieronly') || lower.includes('free tier') || lower.includes('free-tier only')) {
    return '免费额度已耗尽：请在厂商控制台关闭「仅使用免费档」或更换付费 Key。'
  }
  if (lower.includes('quota exhausted') || lower.includes('[resource_exhausted]') || lower.includes('status 429') || lower.includes('too many requests') || lower.includes('rate limit')) {
    return '配额受限或被限流：厂商已拒绝调用，请检查计费/速率限制或稍后重试。'
  }
  if (lower.includes('status 403') && (lower.includes('quota') || lower.includes('exhaust') || lower.includes('allocation'))) {
    return '调用被拒（配额受限）：请检查厂商控制台的计费/配额状态。'
  }
  if (msg.includes('unauthorized')) return '鉴权失败：请检查 API Key/Secret 是否正确。'
  if (msg.includes('endpoint')) return '模型端点不存在：请确认 Base URL 与服务协议匹配（部分服务需要 /v1）。'
  if (msg.includes('timeout')) return '请求超时：请检查网络连通性或稍后重试。'
  if (msg.includes('unreachable')) return '无法连接到模型服务：请检查 Base URL、网络或网关。'
  if (msg.includes('invalid /models')) return '模型服务返回格式不兼容 /models 协议。'
  if (msg.includes('no models returned')) return '模型服务未返回可用模型，请检查账号权限或服务配置。'
  return msg || '拉取模型失败，请检查 Base URL 与密钥配置。'
}
