import { useState, useCallback, useEffect } from 'react'
import { create } from '@bufbuild/protobuf'
import { createClient } from '@connectrpc/connect'
import { AdminSocialService, AdminListPostsRequestSchema, AdminListCommentsRequestSchema, AdminListReportsRequestSchema, AdminListModerationCasesRequestSchema, AdminGetPostDetailRequestSchema, ModerateContentRequestSchema } from '@antclaw/proto/antclaw/v1/admin_social_pb'
import type { AdminPostSummary, AdminCommentSummary, AdminReportSummary, ModerationCaseSummary } from '@antclaw/proto/antclaw/v1/admin_social_pb'
import { transport } from '../lib/transport'
import {
  FileText, MessageSquare, AlertTriangle, Shield, Search, Eye,
  Trash2, RotateCcw, Ban,
} from 'lucide-react'
import { LoadingSkeleton, EmptyState, ErrorState, Badge } from '../components/Common'

const client = createClient(AdminSocialService, transport)

type Tab = 'posts' | 'comments' | 'reports' | 'cases'

const impactColors: Record<string, string> = {
  low: 'bg-gray-100 text-gray-600',
  medium: 'bg-orange-100 text-orange-600',
  high: 'bg-red-100 text-red-700',
  critical: 'bg-red-200 text-red-800',
}

function fmtTime(ts: bigint | undefined): string {
  if (!ts) return ''
  return new Date(Number(ts) * 1000).toLocaleDateString()
}

export default function SocialManagement() {
  const [tab, setTab] = useState<Tab>('posts')
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const [posts, setPosts] = useState<AdminPostSummary[]>([])
  const [postsCursor, setPostsCursor] = useState('')
  const [postsTotal, setPostsTotal] = useState(0)

  const [comments, setComments] = useState<AdminCommentSummary[]>([])
  const [commentsCursor, setCommentsCursor] = useState('')

  const [reports, setReports] = useState<AdminReportSummary[]>([])
  const [reportsCursor, setReportsCursor] = useState('')

  const [cases, setCases] = useState<ModerationCaseSummary[]>([])
  const [casesCursor, setCasesCursor] = useState('')

  const loadPosts = useCallback(async (cursor?: string) => {
    setLoading(true); setError(null)
    try {
      const res = await client.listPosts(create(AdminListPostsRequestSchema, {
        pageSize: 20, cursor: cursor || '', keyword: search || undefined,
      }))
      setPosts(cursor ? prev => [...prev, ...(res.posts as AdminPostSummary[])] : (res.posts as AdminPostSummary[]))
      setPostsCursor(res.nextCursor || '')
      setPostsTotal(res.totalCount || 0)
    } catch (e: any) { setError(e?.message || '加载帖子失败') }
    finally { setLoading(false) }
  }, [search])

  const loadComments = useCallback(async (cursor?: string) => {
    setLoading(true); setError(null)
    try {
      const res = await client.listComments(create(AdminListCommentsRequestSchema, {
        pageSize: 20, cursor: cursor || '',
      }))
      setComments(cursor ? prev => [...prev, ...(res.comments as AdminCommentSummary[])] : (res.comments as AdminCommentSummary[]))
      setCommentsCursor(res.nextCursor || '')
    } catch (e: any) { setError(e?.message || '加载评论失败') }
    finally { setLoading(false) }
  }, [])

  const loadReports = useCallback(async (cursor?: string) => {
    setLoading(true); setError(null)
    try {
      const res = await client.listReports(create(AdminListReportsRequestSchema, {
        pageSize: 20, cursor: cursor || '',
      }))
      setReports(cursor ? prev => [...prev, ...(res.reports as AdminReportSummary[])] : (res.reports as AdminReportSummary[]))
      setReportsCursor(res.nextCursor || '')
    } catch (e: any) { setError(e?.message || '加载举报失败') }
    finally { setLoading(false) }
  }, [])

  const loadCases = useCallback(async (cursor?: string) => {
    setLoading(true); setError(null)
    try {
      const res = await client.listModerationCases(create(AdminListModerationCasesRequestSchema, {
        pageSize: 20, cursor: cursor || '',
      }))
      setCases(cursor ? prev => [...prev, ...(res.cases as ModerationCaseSummary[])] : (res.cases as ModerationCaseSummary[]))
      setCasesCursor(res.nextCursor || '')
    } catch (e: any) { setError(e?.message || '加载审核 case 失败') }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { setError(null)
    switch (tab) {
      case 'posts': if (posts.length === 0) loadPosts(); break
      case 'comments': if (comments.length === 0) loadComments(); break
      case 'reports': if (reports.length === 0) loadReports(); break
      case 'cases': if (cases.length === 0) loadCases(); break
    }
  }, [tab])

  const moderateAction = async (targetType: string, targetId: string, action: string) => {
    try {
      await client.moderateContent(create(ModerateContentRequestSchema, {
        targetType, targetId, action, reason: `Admin ${action}`,
      }))
      if (tab === 'posts') loadPosts()
      else if (tab === 'comments') loadComments()
    } catch (e: any) { setError(e?.message || `操作 ${action} 失败`) }
  }

  const viewPostDetail = async (postId: string) => {
    try { await client.getPostDetail(create(AdminGetPostDetailRequestSchema, { postId })) }
    catch { /* detail view is optional */ }
  }

  const handleSearch = () => { setPosts([]); loadPosts() }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Shield className="w-6 h-6 text-gray-500" />
        <h1 className="text-2xl font-bold text-gray-900">社交管理</h1>
      </div>

      <div className="flex gap-1 bg-gray-100 rounded-lg p-1">
        {([['posts', FileText, '帖子'], ['comments', MessageSquare, '评论'], ['reports', AlertTriangle, '举报'], ['cases', Shield, '审核Case']] as const).map(([id, Icon, label]) => (
          <button key={id} onClick={() => setTab(id)}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              tab === id ? 'bg-white shadow text-blue-600' : 'text-gray-500 hover:text-gray-700'
            }`}><Icon className="w-4 h-4" />{label}</button>
        ))}
      </div>

      {error && <ErrorState message={error} onRetry={() => { setError(null); loadPosts() }} />}

      {tab === 'posts' && (
        <div className="flex gap-2">
          <div className="flex-1 relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input type="text" value={search} onChange={e => setSearch(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleSearch()}
              placeholder="搜索帖子内容或作者..."
              className="w-full pl-9 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 outline-none" />
          </div>
          <button onClick={handleSearch} className="px-4 py-2 bg-blue-500 text-white rounded-lg text-sm hover:bg-blue-600">搜索</button>
        </div>
      )}

      {loading && <LoadingSkeleton rows={5} />}
      {!loading && (
        <>
          {tab === 'posts' && (
            <>
              {postsTotal > 0 && <p className="text-sm text-gray-500">共 {postsTotal} 条帖子</p>}
              {posts.length === 0 ? <EmptyState title="暂无帖子" /> : (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                      <tr>
                        <th className="text-left p-3">帖子</th><th className="text-left p-3 w-20">类型</th><th className="text-left p-3 w-16">状态</th>
                        <th className="text-center p-3 w-16">赞</th><th className="text-center p-3 w-16">评</th><th className="text-center p-3 w-16">举报</th>
                        <th className="text-right p-3 w-32">操作</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {posts.map(p => (
                        <tr key={p.postId} className="hover:bg-gray-50">
                          <td className="p-3">
                            <div className="font-medium text-gray-800 truncate max-w-xs">{p.contentPreview}</div>
                            <div className="text-xs text-gray-400 mt-0.5">
                              {p.authorName || p.authorId?.substring(0, 8)} · {fmtTime(p.createdAt)}
                              {p.visibility !== 'public' && <span className="ml-2 text-gray-300">({p.visibility})</span>}
                            </div>
                          </td>
                          <td className="p-3 text-gray-500">{p.postType}</td>
                          <td className="p-3"><Badge variant={p.status === 'active' ? 'success' : p.status === 'hidden' ? 'warning' : 'danger'}>{p.status}</Badge></td>
                          <td className="p-3 text-center text-gray-500">{String(p.likeCount)}</td>
                          <td className="p-3 text-center text-gray-500">{String(p.commentCount)}</td>
                          <td className="p-3 text-center">
                            {Number(p.reportCount) > 0 ? <span className="text-red-500 font-medium">{String(p.reportCount)}</span> : <span className="text-gray-300">0</span>}
                          </td>
                          <td className="p-3 text-right">
                            <div className="flex items-center justify-end gap-1">
                              <button onClick={() => viewPostDetail(p.postId)} title="查看详情" className="p-1.5 text-gray-400 hover:text-blue-500 rounded"><Eye className="w-4 h-4" /></button>
                              {p.status !== 'hidden' && <button onClick={() => moderateAction('post', p.postId, 'hide')} title="隐藏" className="p-1.5 text-gray-400 hover:text-orange-500 rounded"><Ban className="w-4 h-4" /></button>}
                              {p.status === 'hidden' && <button onClick={() => moderateAction('post', p.postId, 'restore')} title="恢复" className="p-1.5 text-gray-400 hover:text-green-500 rounded"><RotateCcw className="w-4 h-4" /></button>}
                              {p.status !== 'deleted' && <button onClick={() => moderateAction('post', p.postId, 'delete')} title="删除" className="p-1.5 text-gray-400 hover:text-red-500 rounded"><Trash2 className="w-4 h-4" /></button>}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {postsCursor && <div className="p-3 text-center"><button onClick={() => loadPosts(postsCursor)} className="text-sm text-blue-500 hover:text-blue-600">加载更多</button></div>}
                </div>
              )}
            </>
          )}

          {tab === 'comments' && (
            <>
              {comments.length === 0 ? <EmptyState title="暂无评论" /> : (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                      <tr><th className="text-left p-3">评论内容</th><th className="text-left p-3 w-24">作者</th><th className="text-left p-3 w-16">状态</th><th className="text-right p-3 w-32">操作</th></tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {comments.map(c => (
                        <tr key={c.commentId} className="hover:bg-gray-50">
                          <td className="p-3">
                            <div className="text-gray-800 truncate max-w-md">{c.content}</div>
                            <div className="text-xs text-gray-400 mt-0.5">帖子 {c.postId?.substring(0, 8)} · {fmtTime(c.createdAt)}</div>
                          </td>
                          <td className="p-3 text-gray-500">{c.authorName || c.authorId?.substring(0, 8)}</td>
                          <td className="p-3"><Badge variant={c.status === 'active' ? 'success' : 'warning'}>{c.status}</Badge></td>
                          <td className="p-3 text-right">
                            <div className="flex items-center justify-end gap-1">
                              {c.status !== 'hidden' && <button onClick={() => moderateAction('comment', c.commentId, 'hide')} title="隐藏" className="p-1.5 text-gray-400 hover:text-orange-500 rounded"><Ban className="w-4 h-4" /></button>}
                              {c.status !== 'deleted' && <button onClick={() => moderateAction('comment', c.commentId, 'delete')} title="删除" className="p-1.5 text-gray-400 hover:text-red-500 rounded"><Trash2 className="w-4 h-4" /></button>}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {commentsCursor && <div className="p-3 text-center"><button onClick={() => loadComments(commentsCursor)} className="text-sm text-blue-500 hover:text-blue-600">加载更多</button></div>}
                </div>
              )}
            </>
          )}

          {tab === 'reports' && (
            <>
              {reports.length === 0 ? <EmptyState title="暂无举报" /> : (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                      <tr><th className="text-left p-3">举报原因</th><th className="text-left p-3 w-16">类型</th><th className="text-left p-3 w-16">优先级</th><th className="text-left p-3 w-20">状态</th><th className="text-left p-3 w-24">日期</th></tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {reports.map(r => (
                        <tr key={r.reportId} className="hover:bg-gray-50">
                          <td className="p-3">
                            <div className="text-gray-800">{r.reason}</div>
                            <div className="text-xs text-gray-400 mt-0.5">目标: {r.targetType}/{r.targetId?.substring(0, 8)} · 举报者: {r.reporterId?.substring(0, 8)}</div>
                          </td>
                          <td className="p-3 text-gray-500">{r.targetType}</td>
                          <td className="p-3"><span className={`px-2 py-0.5 rounded-full text-xs ${impactColors[r.priority || ''] || 'bg-gray-100 text-gray-600'}`}>{r.priority}</span></td>
                          <td className="p-3"><Badge variant={r.status === 'pending' ? 'warning' : r.status === 'actioned' ? 'success' : 'info'}>{r.status}</Badge></td>
                          <td className="p-3 text-gray-400 text-xs">{fmtTime(r.createdAt)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {reportsCursor && <div className="p-3 text-center"><button onClick={() => loadReports(reportsCursor)} className="text-sm text-blue-500 hover:text-blue-600">加载更多</button></div>}
                </div>
              )}
            </>
          )}

          {tab === 'cases' && (
            <>
              {cases.length === 0 ? <EmptyState title="暂无审核 case" /> : (
                <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-gray-50 text-gray-500 text-xs uppercase">
                      <tr><th className="text-left p-3">Case ID</th><th className="text-left p-3">来源/目标</th><th className="text-left p-3 w-16">优先级</th><th className="text-left p-3 w-20">状态</th><th className="text-left p-3 w-24">日期</th></tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {cases.map(c => (
                        <tr key={c.caseId} className="hover:bg-gray-50">
                          <td className="p-3 font-mono text-xs text-gray-500">{c.caseId?.substring(0, 12)}...</td>
                          <td className="p-3"><div className="text-gray-800 text-xs">{c.source}</div><div className="text-xs text-gray-400">{c.targetType}/{c.targetId?.substring(0, 8)}</div></td>
                          <td className="p-3"><span className={`px-2 py-0.5 rounded-full text-xs ${impactColors[c.priority || ''] || 'bg-gray-100 text-gray-600'}`}>{c.priority}</span></td>
                          <td className="p-3"><Badge variant={c.status === 'open' ? 'warning' : 'success'}>{c.status}</Badge></td>
                          <td className="p-3 text-gray-400 text-xs">{fmtTime(c.createdAt)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  {casesCursor && <div className="p-3 text-center"><button onClick={() => loadCases(casesCursor)} className="text-sm text-blue-500 hover:text-blue-600">加载更多</button></div>}
                </div>
              )}
            </>
          )}
        </>
      )}
    </div>
  )
}
