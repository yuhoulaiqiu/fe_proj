import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { apiGetActivity } from '../services/publicApi.js'

function ActivityDetailPage() {
  const { id } = useParams()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [data, setData] = useState(null)

  useEffect(() => {
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const res = await apiGetActivity(id)
        if (cancelled) return
        setData(res)
      } catch (err) {
        if (cancelled) return
        const msg =
          err?.response?.data?.message || err?.message || '加载失败，请稍后重试'
        setError(msg)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    run()
    return () => {
      cancelled = true
    }
  }, [id])

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">活动详情</h1>
        <p className="muted">活动 ID：{id}</p>
      </div>

      {error ? <Alert variant="danger">{error}</Alert> : null}

      {loading ? (
        <LoadingCard title="正在加载活动详情…" lines={3} />
      ) : data ? (
        <Card>
          <h2 className="card-title">{data.title || '未命名活动'}</h2>
          <div className="chips">
            {data.location ? <Badge variant="neutral">{`地点：${data.location}`}</Badge> : null}
            {data.startTime || data.endTime ? (
              <Badge variant="neutral">
                {`时间：${data.startTime || '-'} ~ ${data.endTime || '-'}`}
              </Badge>
            ) : null}
          </div>
          <div className="mt-3">
            <p className="muted">{data.content || data.summary || '暂无内容'}</p>
          </div>
        </Card>
      ) : (
        <EmptyState description="未找到该活动，可能已被删除或 ID 不存在。" />
      )}

      <Link className="btn btn-secondary" to="/activities">
        返回列表
      </Link>
    </div>
  )
}

export default ActivityDetailPage
