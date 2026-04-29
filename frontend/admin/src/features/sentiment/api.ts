import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  SentimentExtrasService,
  GetCBOEPutCallRequestSchema,
  GetMyFXBookPositionsRequestSchema,
  GetInsiderTradesRequestSchema,
  GetCryptoSocialRequestSchema,
  GetFinvizMetricsRequestSchema,
} from '@antclaw/proto/antclaw/v1/sentiment_extras_pb'
import { transport } from '../_shared/transport'

const c = createClient(SentimentExtrasService, transport)

export const fetchCBOE = () => c.getCBOEPutCall(create(GetCBOEPutCallRequestSchema, {}))
export const fetchMyFXBook = (pair: string) =>
  c.getMyFXBookPositions(create(GetMyFXBookPositionsRequestSchema, { pair } as any))
export const fetchInsider = () => c.getInsiderTrades(create(GetInsiderTradesRequestSchema, {}))
export const fetchCryptoSocial = (asset: string) =>
  c.getCryptoSocial(create(GetCryptoSocialRequestSchema, { asset } as any))
export const fetchFinviz = (symbol: string) =>
  c.getFinvizMetrics(create(GetFinvizMetricsRequestSchema, { symbol } as any))
