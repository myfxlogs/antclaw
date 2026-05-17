// Crypto API (A13-P1-01 split)
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { CryptoService, GetCryptoPublicKeyRequestSchema } from '@antclaw/proto/antclaw/v1/crypto_pb'
import { transport } from './transport'

const cryptoClient = createClient(CryptoService, transport)

export async function getCryptoPublicKeyPem(): Promise<string> {
  const res = await cryptoClient.getCryptoPublicKey(create(GetCryptoPublicKeyRequestSchema, {}))
  return res.pem || ''
}
