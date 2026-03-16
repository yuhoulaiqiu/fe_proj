import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { apiGetService } from '../services/publicApi.js'

function ServiceDetailPage() {
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
        const res = await apiGetService(id)
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
        <h1 className="page-title">服务详情</h1>
        <p className="muted">服务 ID：{id}</p>
      </div>

      {error ? <Alert variant="danger">{error}</Alert> : null}

      {loading ? (
        <LoadingCard title="正在加载服务详情…" lines={3} />
      ) : data ? (
        <Card>
          <h2 className="card-title">{data.name || '未命名服务'}</h2>
          <div className="chips">
            <Badge variant="neutral">{`类别：${data.category || '-'}`}</Badge>
            {data.phone ? <Badge variant="neutral">{`电话：${data.phone}`}</Badge> : null}
            {data.address ? <Badge variant="neutral">{`地址：${data.address}`}</Badge> : null}
          </div>
          <div className="mt-3">
            <p className="muted">{data.description || '暂无说明'}</p>
          </div>
        </Card>
      ) : (
        <EmptyState description="未找到该服务，可能已下线或 ID 不存在。" />
      )}

      <Link className="btn btn-secondary" to="/services">
        返回目录
      </Link>
    </div>
  )
}

export default ServiceDetailPage
