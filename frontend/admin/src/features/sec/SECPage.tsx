// /sec —— SEC EDGAR 文件列表（按 CIK）。
import { useState } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { SECService, ListFilingsRequestSchema } from '@antclaw/proto/antclaw/v1/sec_pb'
import { transport } from '../_shared/transport'
import { AsyncView, PageShell, useAsync } from '../_shared/AsyncView'
import { formatDate } from '../_shared/format'

const client = createClient(SECService, transport)

export default function SECPage() {
  const [cik, setCik] = useState('320193')   // Apple
  const [submitted, setSubmitted] = useState(cik)
  const { state } = useAsync(
    () => client.listFilings(create(ListFilingsRequestSchema, { cik: submitted } as any)),
    [submitted],
  )
  return (
    <PageShell
      title="SEC EDGAR 文件"
      subtitle="按 CIK 列出最近文件（默认 Apple = 320193）"
      actions={
        <div className="flex gap-2">
          <input className="input w-32" value={cik} onChange={(e) => setCik(e.target.value)} />
          <button onClick={() => setSubmitted(cik)} className="px-3 py-1 bg-blue-600 text-white rounded text-sm">
            查询
          </button>
        </div>
      }
    >
      <AsyncView
        state={state}
        render={(d: any) => {
          const items: any[] = d.filings || d.items || []
          return (
            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left py-1 px-2">日期</th>
                  <th className="text-left py-1 px-2">表单</th>
                  <th className="text-left py-1 px-2">编号</th>
                  <th className="text-left py-1 px-2">链接</th>
                </tr>
              </thead>
              <tbody>
                {items.slice(0, 30).map((f, i) => (
                  <tr key={i} className="border-b">
                    <td className="py-1 px-2">{formatDate(f.filedAt ?? f.filed_at ?? f.date)}</td>
                    <td className="py-1 px-2 font-mono text-xs">{f.formType || f.form_type}</td>
                    <td className="py-1 px-2 font-mono text-xs">{f.accessionNumber || f.accession_number}</td>
                    <td className="py-1 px-2">
                      {f.url ? <a href={f.url} target="_blank" className="text-blue-600 hover:underline">查看</a> : '-'}
                    </td>
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
