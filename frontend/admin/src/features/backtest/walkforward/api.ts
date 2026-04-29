import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  BacktestService,
  RunWalkforwardRequestSchema,
  GetWalkforwardResultRequestSchema,
  GetTradesRequestSchema,
} from '@antclaw/proto/antclaw/v1/backtest_pb'
import { transport } from '../../_shared/transport'

const c = createClient(BacktestService, transport)

export interface WFParams {
  strategy: string
  symbols: string[]
  fromDate: string
  toDate: string
  folds: number
  trainRatio: number
}

export async function runWalkforward(p: WFParams) {
  return c.runWalkforward(create(RunWalkforwardRequestSchema, p as any))
}
export async function getWalkforwardResult(jobId: string) {
  return c.getWalkforwardResult(create(GetWalkforwardResultRequestSchema, { jobId }))
}
export async function getTrades(jobId: string) {
  return c.getTrades(create(GetTradesRequestSchema, { jobId }))
}
