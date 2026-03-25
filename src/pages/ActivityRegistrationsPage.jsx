import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import {
  apiExportActivityRegistrationsCsv,
  apiGetActivity,
  apiGetActivityRegistrations,
} from '../services/publicApi.js'

function ActivityRegistrationsPage() {
  const { id } = useParams()
  const { addToast } = useToast()

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [activity, setActivity] = useState(null)
  const [items, setItems] = useState([])
  const [exporting, setExporting] = useState(false)

  const ownerId = useMemo(() => {
    try {
      return JSON.parse(localStorage.getItem('auth_user') || '{}')?.id
    } catch {
      return undefined
    }
  }, [])

  const isOwner = useMemo(() => {
    if (!activity) return false
    if (!ownerId) return false
    return String(ownerId) === String(activity.userId)
  }, [activity, ownerId])

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const [a, regs] = await Promise.all([
        apiGetActivity(id),
        apiGetActivityRegistrations(id),
      ])
      setActivity(a)
      setItems(Array.isArray(regs) ? regs : [])
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '加载失败，请稍后重试'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [id])

  const onExport = async () => {
    setExporting(true)
    try {
      const blob = await apiExportActivityRegistrationsCsv(id)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `activity_${id}_registrations.csv`
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
      addToast('已开始下载', 'success')
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '导出失败，请稍后重试'
      addToast(msg, 'danger')
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <div className="row-between">
          <div>
            <h1 className="page-title">报名管理</h1>
            <p className="muted">活动 ID：{id}</p>
          </div>
          <div className="actions">
            <Link className="btn btn-secondary" to={`/activities/${id}`}>
              返回详情
            </Link>
            <button className="btn" type="button" disabled={exporting || !isOwner} onClick={onExport}>
              {exporting ? '导出中…' : '导出 CSV'}
            </button>
          </div>
        </div>
      </div>

      {loading ? <LoadingCard title="正在加载报名信息…" lines={5} /> : null}
      {error ? <Alert variant="danger">{error}</Alert> : null}
      {!loading && activity && !isOwner ? (
        <Alert variant="warning">仅活动发布者可查看报名管理。</Alert>
      ) : null}

      {!loading && isOwner ? (
        <Card>
          <div className="row-between">
            <div>
              <h2 className="card-title">{activity?.title || '未命名活动'}</h2>
              <p className="muted">
                共 {Array.isArray(items) ? items.length : 0} 人报名
              </p>
            </div>
            <button className="btn btn-secondary" type="button" onClick={load} disabled={loading}>
              刷新
            </button>
          </div>

          <div className="mt-4">
            {items.length ? (
              <div className="stack" style={{ gap: '10px' }}>
                {items.map((it) => (
                  <div
                    key={it.id}
                    className="row-between"
                    style={{
                      padding: '12px',
                      border: '1px solid var(--border-color)',
                      borderRadius: '8px',
                    }}
                  >
                    <div>
                      <div style={{ fontSize: '16px', fontWeight: 600 }}>
                        {it.username || `用户 ${it.userId}`}
                      </div>
                      <div className="muted" style={{ fontSize: '14px', marginTop: '4px' }}>
                        报名时间：{it.createdAt || '-'}
                      </div>
                    </div>
                    <div className="muted" style={{ fontSize: '14px' }}>
                      {it.status || 'pending'}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <EmptyState description="暂无报名记录" />
            )}
          </div>
        </Card>
      ) : null}
    </div>
  )
}

export default ActivityRegistrationsPage

