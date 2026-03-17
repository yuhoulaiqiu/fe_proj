import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { apiGetLostItem } from '../services/publicApi.js'

const TYPE_LABEL = { lost: '失物', found: '招领' }
const STATUS_LABEL = { open: '未处理', claimed: '已认领', returned: '已归还' }
const STATUS_BADGE = { open: 'warning', claimed: 'neutral', returned: 'success' }

function LostItemDetailPage() {
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
        const res = await apiGetLostItem(id)
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
        <h1 className="page-title">失物招领详情</h1>
        <p className="muted">记录 ID：{id}</p>
      </div>

      {error ? <Alert variant="danger">{error}</Alert> : null}

      {loading ? (
        <LoadingCard title="正在加载详情…" lines={3} />
      ) : data ? (
        <Card>
          <h2 className="card-title">{data.title || '未命名记录'}</h2>
          <div className="chips">
            {data.itemType ? (
              <Badge variant="neutral">
                {`类型：${TYPE_LABEL[data.itemType] || data.itemType}`}
              </Badge>
            ) : null}
            {data.status ? (
              <Badge variant={STATUS_BADGE[data.status] || 'neutral'}>
                {`状态：${STATUS_LABEL[data.status] || data.status}`}
              </Badge>
            ) : null}
            {data.location ? <Badge variant="neutral">{`地点：${data.location}`}</Badge> : null}
            {data.occurredAt ? (
              <Badge variant="neutral">{`时间：${data.occurredAt}`}</Badge>
            ) : null}
            {data.contact ? <Badge variant="neutral">{`联系方式：${data.contact}`}</Badge> : null}
          </div>
          <div className="mt-3">
            <p className="muted">{data.description || '暂无描述'}</p>
          </div>
        </Card>
      ) : (
        <EmptyState description="未找到该记录，可能已被删除或 ID 不存在。" />
      )}

      <Link className="btn btn-secondary" to="/lost-found">
        返回列表
      </Link>
    </div>
  )
}

export default LostItemDetailPage
