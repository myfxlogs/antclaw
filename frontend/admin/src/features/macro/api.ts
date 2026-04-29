// macro 系列共享 client：FedWatch / MacroExtras / Treasury / Macro 主体。
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  FedWatchService,
  GetFOMCProbabilitiesRequestSchema,
} from '@antclaw/proto/antclaw/v1/fedwatch_pb'
import {
  MacroExtrasService,
  MacroExtrasServiceGetSeriesRequestSchema,
  ListAvailableSeriesRequestSchema,
} from '@antclaw/proto/antclaw/v1/macro_extras_pb'
import {
  TreasuryService,
  GetCurveRequestSchema,
} from '@antclaw/proto/antclaw/v1/treasury_pb'
import { transport } from '../_shared/transport'

const fed = createClient(FedWatchService, transport)
const xt = createClient(MacroExtrasService, transport)
const tr = createClient(TreasuryService, transport)

export const fetchFOMC = () => fed.getFOMCProbabilities(create(GetFOMCProbabilitiesRequestSchema, {}))
export const fetchMacroSeries = (source: string, seriesId: string) =>
  xt.getSeries(create(MacroExtrasServiceGetSeriesRequestSchema, { source, seriesId } as any))
export const listMacroSeries = (source: string) =>
  xt.listAvailableSeries(create(ListAvailableSeriesRequestSchema, { source } as any))
export const fetchTreasuryCurve = () => tr.getCurve(create(GetCurveRequestSchema, {}))
