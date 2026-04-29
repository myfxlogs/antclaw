// /defi —— DefiLlama 协议榜。
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import {
  DeFiService,
  GetTopProtocolsRequestSchema,
} from '@antclaw/proto/antclaw/v1/defi_pb'
import { transport } from '../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'

const client = createClient(DeFiService, transport)

export default function DeFiPage() {
  const { state } = useAsync(() => client.getTopProtocols(create(GetTopProtocolsRequestSchema, { limit: 50 })), [])
  return (
    <PageShell title="DeFi 协议榜" subtitle="DefiLlama TVL 排名">
      <AsyncView
        state={state}
        render={(d: any) => {
          const items: any[] = d.protocols || d.items || []
          return (
            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-1 px-2">#</th>
                  <th className="text-left py-1 px-2">协议</th>
                  <th className="text-left py-1 px-2">类别</th>
                  <th className="text-right py-1 px-2">TVL (USD)</th>
                  <th className="text-right py-1 px-2">24h 变化</th>
                </tr>
              </thead>
              <tbody>
                {items.slice(0, 50).map((p, i) => (
                  <tr key={i} className="border-b">
                    <td className="py-1 px-2">{i + 1}</td>
                    <td className="py-1 px-2 font-medium">{p.name || p.slug}</td>
                    <td className="py-1 px-2 text-gray-500">{p.category || '-'}</td>
                    <td className="py-1 px-2 text-right">{Number(p.tvl ?? 0).toLocaleString()}</td>
                    <td className="py-1 px-2 text-right">{p.change24h != null ? `${(p.change24h * 100).toFixed(2)}%` : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        }}
      />
    </PageShell>
  )
}
