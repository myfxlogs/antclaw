// features/options 共享 client。
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  OptionsService,
  GetGEXRequestSchema,
  GetIVSurfaceRequestSchema,
  GetOptionsSkewRequestSchema,
  GetIVAlertsRequestSchema,
} from '@antclaw/proto/antclaw/v1/options_pb'
import { transport } from '../_shared/transport'

const client = createClient(OptionsService, transport)

export async function fetchGEX(asset: string) {
  return client.getGEX(create(GetGEXRequestSchema, { asset }))
}
export async function fetchIVSurface(asset: string) {
  return client.getIVSurface(create(GetIVSurfaceRequestSchema, { asset }))
}
export async function fetchSkew(asset: string) {
  return client.getOptionsSkew(create(GetOptionsSkewRequestSchema, { asset }))
}
export async function fetchIVAlerts(asset: string) {
  return client.getIVAlerts(create(GetIVAlertsRequestSchema, { asset }))
}
