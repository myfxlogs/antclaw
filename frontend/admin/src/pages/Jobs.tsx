import { Fragment, useEffect, useState } from 'react'
import { RotateCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { listJobs, runJob, setJobEnabled } from '../lib/api'

interface Job {
  job_id: string
  job_name: string
  status: string
  last_run: string
  next_run: string
  enabled: boolean
  last_error: string
}

function formatTime(isoTime: string): string {
  if (!isoTime) return '-'
  try {
    const d = new Date(isoTime)
    if (isNaN(d.getTime())) return '-'
    return d.toLocaleString('zh-CN', { hour12: false })
  } catch {
    return '-'
  }
}

export default function Jobs() {
  const { t } = useTranslation()
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [toggling, setToggling] = useState<string | null>(null)
  const [toast, setToast] = useState<{ msg: string; kind: 'ok' | 'err' } | null>(null)
  const [runningAll, setRunningAll] = useState(false)

  const showToast = (msg: string, kind: 'ok' | 'err' = 'ok') => {
    setToast({ msg, kind })
    setTimeout(() => setToast(null), 2500)
  }

  useEffect(() => {
    loadJobs()
    // 订阅后端 SSE 事件，实现实时 Job 状态更新
    const evtSource = new EventSource('/sse/jobs')
    evtSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as {
          job_id: string
          name?: string
          status?: string
          started_at?: number
          finished_at?: number
          error?: string
        }
        setJobs((prev) => {
          const idx = prev.findIndex((j) => j.job_id === data.job_id)
          if (idx === -1) {
            return prev
          }
          const updated = [...prev]
          const job = updated[idx]
          updated[idx] = {
            ...job,
            job_name: data.name || job.job_name,
            status: data.status || job.status,
            last_run: data.finished_at ? new Date(data.finished_at * 1000).toISOString() : job.last_run,
            last_error: data.status === 'failed' ? (data.error || job.last_error) : (data.status === 'succeeded' ? '' : job.last_error),
          }
          return updated
        })
      } catch {
        // ignore malformed events
      }
    }
    evtSource.onerror = () => {
      evtSource.close()
    }
    return () => {
      evtSource.close()
    }
  }, [])

  const loadJobs = async () => {
    try {
      const response = await listJobs()
      setJobs(response.jobs)
    } catch (err) {
      console.error('Failed to load jobs:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleRunJob = async (jobId: string) => {
    try {
      await runJob(jobId)
      // 立即把该行状态切到 running，等 SSE 推送最终状态（succeeded/failed）
      setJobs((prev) => prev.map((j) => (j.job_id === jobId ? { ...j, status: 'running' } : j)))
      showToast(`已触发：${jobId}`, 'ok')
    } catch (err) {
      console.error('Failed to run job:', err)
      showToast(`触发失败：${jobId}`, 'err')
    }
  }

  const handleRunAll = async () => {
    if (runningAll) return
    setRunningAll(true)
    const targets = jobs.filter((j) => j.enabled)
    let ok = 0
    let fail = 0
    for (const j of targets) {
      try {
        await runJob(j.job_id)
        setJobs((prev) => prev.map((x) => (x.job_id === j.job_id ? { ...x, status: 'running' } : x)))
        ok++
      } catch (e) {
        console.error('runJob failed', j.job_id, e)
        fail++
      }
    }
    setRunningAll(false)
    showToast(`已触发 ${ok} 个任务${fail ? `（失败 ${fail}）` : ''}`, fail ? 'err' : 'ok')
  }

  const handleToggleEnabled = async (jobId: string, currentEnabled: boolean) => {
    setToggling(jobId)
    try {
      await setJobEnabled(jobId, !currentEnabled)
      setJobs((prev) =>
        prev.map((j) => (j.job_id === jobId ? { ...j, enabled: !currentEnabled } : j)),
      )
    } catch (err) {
      console.error('Failed to toggle job:', err)
      showToast(`切换启用状态失败：${jobId}`, 'err')
    } finally {
      setToggling(null)
    }
  }

  const getStatusLabel = (status: string) => {
    if (status === 'running') return t('jobs.running')
    if (status === 'succeeded') return t('jobs.succeeded')
    if (status === 'scheduled' || status === 'pending') return t('jobs.pending')
    if (status === 'failed') return t('jobs.failed')
    return t('jobs.unknown')
  }

  const getStatusClass = (status: string) => {
    if (status === 'running') return 'bg-blue-100 text-blue-700'
    if (status === 'succeeded') return 'bg-green-100 text-green-700'
    if (status === 'scheduled' || status === 'pending') return 'bg-yellow-100 text-yellow-700'
    if (status === 'failed') return 'bg-red-100 text-red-700'
    return 'bg-gray-100 text-gray-700'
  }

  if (loading) {
    return <div className="flex items-center justify-center h-64">{t('jobs.loading')}</div>
  }

  return (
    <div className="space-y-6 relative">
      {toast && (
        <div
          className={`fixed top-6 right-6 z-50 px-4 py-2 rounded-lg shadow-lg text-sm text-white ${
            toast.kind === 'ok' ? 'bg-green-600' : 'bg-red-600'
          }`}
        >
          {toast.msg}
        </div>
      )}

      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{t('jobs.title')}</h1>
        <button
          onClick={handleRunAll}
          disabled={runningAll || jobs.length === 0}
          className={`inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium ${
            runningAll
              ? 'bg-gray-300 text-gray-600 cursor-wait'
              : 'bg-blue-600 text-white hover:bg-blue-700'
          }`}
          title={t('jobs.runAll') || '启动全部任务'}
        >
          <RotateCw className="w-4 h-4" />
          {runningAll ? t('jobs.running') : (t('jobs.runAll') || '启动全部任务')}
        </button>
      </div>

      <div className="bg-white rounded-xl shadow-sm overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.jobName')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.jobId')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.enabled')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.status')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.lastRun')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.nextRun')}</th>
              <th className="px-6 py-3 text-left text-sm font-medium text-gray-500">{t('jobs.actions')}</th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {jobs.map((job) => (
              <Fragment key={job.job_id}>
              <tr className="hover:bg-gray-50">
                <td className="px-6 py-4 font-medium">{job.job_name}</td>
                <td className="px-6 py-4 text-xs font-mono text-gray-500">{job.job_id}</td>
                <td className="px-6 py-4">
                  <button
                    type="button"
                    role="switch"
                    aria-checked={job.enabled}
                    disabled={toggling === job.job_id}
                    onClick={() => handleToggleEnabled(job.job_id, job.enabled)}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition ${
                      job.enabled ? 'bg-green-500' : 'bg-gray-300'
                    } ${toggling === job.job_id ? 'opacity-50 cursor-wait' : 'cursor-pointer'}`}
                    title={job.enabled ? t('jobs.disable') : t('jobs.enable')}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition ${
                        job.enabled ? 'translate-x-6' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </td>
                <td className="px-6 py-4">
                  <span
                    className={`inline-flex px-2 py-1 rounded text-xs font-medium ${getStatusClass(job.status)}`}
                  >
                    {getStatusLabel(job.status)}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{formatTime(job.last_run)}</td>
                <td className="px-6 py-4 text-sm text-gray-600">{formatTime(job.next_run)}</td>
                <td className="px-6 py-4">
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleRunJob(job.job_id)}
                      className="p-2 text-blue-600 hover:bg-blue-50 rounded"
                      title={t('jobs.run')}
                    >
                      <RotateCw className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
              {job.last_error && (
                <tr className="bg-red-50">
                  <td colSpan={7} className="px-6 py-2 text-xs text-red-700">
                    <span className="font-semibold mr-2">最近错误:</span>
                    <span className="break-all">{job.last_error}</span>
                  </td>
                </tr>
              )}
              </Fragment>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}