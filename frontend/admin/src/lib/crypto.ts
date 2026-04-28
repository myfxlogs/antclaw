// 浏览器端加密工具：RSA-OAEP（SHA-256）+ AES-256-GCM + HMAC-SHA256。
//
// 与后端 internal/crypto 包对接的协议：
//   - 公钥：通过 Connect CryptoService.GetCryptoPublicKey（api.getCryptoPublicKeyPem）拉到 PEM；
//   - 加密：随机 32B aesKey + 12B iv → AES-GCM 加密 payload；
//     aesKey 用 RSA-OAEP(SHA-256) 加 server pubkey 加密；
//   - 签名：sessionKey = SHA-256(jwt_token)；
//     sig = HMAC-SHA256(sessionKey, `${ts}\n${nonce}\n${body}`)；
//   - body 即 envelope JSON 字符串。
//
// 全部使用 WebCrypto（SubtleCrypto），无第三方依赖。

import { getCryptoPublicKeyPem } from './api'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { CryptoService, PostEnvelopeRequestSchema } from '@antclaw/proto/antclaw/v1/crypto_pb'

const enc = new TextEncoder()

// 把 Uint8Array 转换为纯 ArrayBuffer，规避 SharedArrayBuffer 类型分歧。
// 显式 .slice() 拷贝产出非 shared buffer，TS5 严格模式可接受为 BufferSource。
function toAB(u8: Uint8Array): ArrayBuffer {
  return u8.buffer.slice(u8.byteOffset, u8.byteOffset + u8.byteLength) as ArrayBuffer
}

// ============ base64 helpers ============
const b64 = {
  encode(buf: ArrayBuffer | Uint8Array): string {
    const u8 = buf instanceof Uint8Array ? buf : new Uint8Array(buf)
    let s = ''
    for (let i = 0; i < u8.length; i++) s += String.fromCharCode(u8[i])
    return btoa(s)
  },
  decode(s: string): Uint8Array {
    const bin = atob(s)
    const u8 = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; i++) u8[i] = bin.charCodeAt(i)
    return u8
  },
}

// ============ PEM → CryptoKey ============
async function importRsaPublicKeyFromPEM(pem: string): Promise<CryptoKey> {
  const body = pem
    .replace(/-----BEGIN PUBLIC KEY-----/, '')
    .replace(/-----END PUBLIC KEY-----/, '')
    .replace(/\s+/g, '')
  const der = b64.decode(body)
  return crypto.subtle.importKey(
    'spki',
    toAB(der),
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
}

let cachedPubKey: { pem: string; key: CryptoKey } | null = null

/** 获取并缓存服务端 RSA 公钥。 */
export async function getServerPublicKey(): Promise<CryptoKey> {
  if (cachedPubKey) return cachedPubKey.key
  const pem = (await getCryptoPublicKeyPem()).trim()
  if (!pem) throw new Error('empty server public key')
  const key = await importRsaPublicKeyFromPEM(pem)
  cachedPubKey = { pem, key }
  return key
}

// ============ Hybrid 加密 ============
export interface HybridEnvelope {
  key_enc: string
  iv: string
  ciphertext: string
}

/** 用服务端公钥执行混合加密，返回 envelope。 */
export async function hybridEncryptJSON(payload: unknown): Promise<HybridEnvelope> {
  const pubKey = await getServerPublicKey()

  const aesKeyRaw = crypto.getRandomValues(new Uint8Array(32))
  const iv = crypto.getRandomValues(new Uint8Array(12))

  const aesKey = await crypto.subtle.importKey(
    'raw',
    toAB(aesKeyRaw),
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt'],
  )
  const plaintext = enc.encode(JSON.stringify(payload))
  const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv: toAB(iv) }, aesKey, toAB(plaintext))
  const keyEnc = await crypto.subtle.encrypt({ name: 'RSA-OAEP' }, pubKey, toAB(aesKeyRaw))

  return {
    key_enc: b64.encode(keyEnc),
    iv: b64.encode(iv),
    ciphertext: b64.encode(ct),
  }
}

// ============ 接口签名 ============
async function deriveSessionKey(token: string): Promise<CryptoKey> {
  const hash = await crypto.subtle.digest('SHA-256', toAB(enc.encode(token)))
  return crypto.subtle.importKey(
    'raw',
    hash,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign'],
  )
}

function randomNonce(): string {
  const u8 = crypto.getRandomValues(new Uint8Array(16))
  let s = ''
  for (let i = 0; i < u8.length; i++) s += u8[i].toString(16).padStart(2, '0')
  return s
}

/** 签名后端要求的 (timestamp, nonce, body)；返回 hex 形式的 HMAC。 */
export async function signRequest(token: string, body: string): Promise<{
  timestamp: string
  nonce: string
  signature: string
}> {
  const timestamp = Math.floor(Date.now() / 1000).toString()
  const nonce = randomNonce()
  const sessionKey = await deriveSessionKey(token)
  const data = enc.encode(`${timestamp}\n${nonce}\n${body}`)
  const sig = await crypto.subtle.sign('HMAC', sessionKey, toAB(data))
  // hex 输出，与后端 hex.EncodeToString 保持一致
  const u8 = new Uint8Array(sig)
  let hex = ''
  for (let i = 0; i < u8.length; i++) hex += u8[i].toString(16).padStart(2, '0')
  return { timestamp, nonce, signature: hex }
}

/** 高层封装：发送一次需要"加密 + 签名"的 PUT 请求。 */
export async function sendSecurePut(
  url: string,
  payload: unknown,
): Promise<{ body_b64: string }> {
  const token = localStorage.getItem('token') || ''
  if (!token) throw new Error('not authenticated')
  // 1) 生成 envelope 并 base64 编码
  const envelope = await hybridEncryptJSON(payload)
  const bodyJson = JSON.stringify(envelope)
  const bodyB64 = b64.encode(enc.encode(bodyJson))
  // 2) 计算签名（签名对象为 base64 字符串）
  const sig = await signRequest(token, bodyB64)
  // 3) 通过 Connect 调用 CryptoService.PostEnvelope
  const API_BASE_URL = (import.meta as any).env?.VITE_API_BASE_URL || 'http://localhost:8082'
  const transport = createConnectTransport({ baseUrl: API_BASE_URL })
  const client = createClient(CryptoService, transport)
  const res = await client.postEnvelope({
    bodyB64,
    ts: sig.timestamp,
    nonce: sig.nonce,
    sig: sig.signature,
    targetPath: url,
    targetBodyB64: '',
  } as any /* PostEnvelopeRequest */)
  return { body_b64: (res as any).bodyB64 || '' }
}

