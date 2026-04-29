import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  VolService,
  GetVixRequestSchema,
  GetMoveRequestSchema,
  GetDvolRequestSchema,
} from '@antclaw/proto/antclaw/v1/vol_pb'
import { transport } from '../_shared/transport'

const c = createClient(VolService, transport)
export const fetchVix = () => c.getVix(create(GetVixRequestSchema, {}))
export const fetchMove = () => c.getMove(create(GetMoveRequestSchema, {}))
export const fetchDvol = (asset: string) => c.getDvol(create(GetDvolRequestSchema, { asset }))
