import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
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

      {error ? <div className="alert alert-danger">{error}</div> : null}

      <div className="card">
        {loading ? (
          <p className="muted">正在加载…</p>
        ) : data ? (
          <>
            <h2 className="card-title">{data.title || '未命名活动'}</h2>
            {data.location ? <p className="muted">地点：{data.location}</p> : null}
            {data.startTime || data.endTime ? (
              <p className="muted">
                时间：{data.startTime || '-'} ~ {data.endTime || '-'}
              </p>
            ) : null}
            <div style={{ marginTop: 10 }}>
              <p className="muted">{data.content || data.summary || '暂无内容'}</p>
            </div>
          </>
        ) : (
          <p className="muted">暂无数据。</p>
        )}
      </div>

      <Link className="btn btn-secondary" to="/activities">
        返回列表
      </Link>
    </div>
  )
}

export default ActivityDetailPage
