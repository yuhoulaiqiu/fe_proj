import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiGetActivity } from '../services/publicApi.js'

function ActivityDetailPage() {
  const { id } = useParams()
  const { addToast } = useToast()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [data, setData] = useState(null)
  const [isRegistered, setIsRegistered] = useState(false)

  useEffect(() => {
    const registered = JSON.parse(localStorage.getItem('registered_activities') || '[]')
    // 将 id 转为字符串进行比较，确保一致性
    setIsRegistered(registered.includes(String(id)))
  }, [id])

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

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      addToast('电话已复制', 'success')
    })
  }

  const onRegister = () => {
    const registered = JSON.parse(localStorage.getItem('registered_activities') || '[]')
    const stringId = String(id)
    if (!registered.includes(stringId)) {
      registered.push(stringId)
      localStorage.setItem('registered_activities', JSON.stringify(registered))
      setIsRegistered(true)
      addToast('报名成功！', 'success')
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <div className="row-between">
          <div>
            <h1 className="page-title">活动详情</h1>
            <p className="muted">活动 ID：{id}</p>
          </div>
          <Link className="btn btn-secondary" to="/activities">
            返回列表
          </Link>
        </div>
      </div>

      {error ? <Alert variant="danger">{error}</Alert> : null}

      {loading ? (
        <LoadingCard title="正在加载活动详情…" lines={5} />
      ) : data ? (
        <div className="stack">
          <Card>
            <div className="row-between">
              <div className="grow">
                <div className="chips" style={{ marginBottom: '12px' }}>
                  {data.category && <Badge variant="neutral">{data.category}</Badge>}
                  {data.status && (
                    <Badge variant={data.status === 'active' ? 'success' : 'warning'}>
                      {data.status === 'active' ? '进行中' : '已结束'}
                    </Badge>
                  )}
                </div>
                <h2 className="card-title" style={{ fontSize: '24px', marginBottom: '16px' }}>
                  {data.title || '未命名活动'}
                </h2>
                <div className="stack" style={{ gap: '8px' }}>
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <span className="label muted">地点：</span>
                    <span>{data.location || '社区活动中心'}</span>
                  </div>
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <span className="label muted">时间：</span>
                    <span>
                      {data.startTime || '-'} ~ {data.endTime || '-'}
                    </span>
                  </div>
                  {data.phone && (
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <span className="label muted">咨询电话：</span>
                      <button
                        className="btn btn-secondary"
                        style={{ padding: '2px 8px', fontSize: '14px' }}
                        onClick={() => handleCopy(data.phone)}
                        title="点击复制"
                      >
                        {data.phone}
                      </button>
                    </div>
                  )}
                </div>
              </div>
              <div className="minw-240" style={{ display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
                <button
                  className="btn"
                  style={{ padding: '12px 32px', fontSize: '18px', width: '100%' }}
                  disabled={data.status !== 'active' || isRegistered}
                  onClick={onRegister}
                >
                  {isRegistered ? '已报名' : data.status === 'active' ? '立即报名' : '报名已结束'}
                </button>
              </div>
            </div>
          </Card>

          <Card>
            <h3 className="card-title">活动介绍</h3>
            <div className="mt-3">
              <p className="lead" style={{ whiteSpace: 'pre-wrap', lineHeight: '1.6' }}>
                {data.content || data.summary || '暂无详细介绍内容。'}
              </p>
            </div>
          </Card>
        </div>
      ) : (
        <EmptyState description="未找到该活动，可能已被删除或 ID 不存在。" />
      )}
    </div>
  )
}

export default ActivityDetailPage
