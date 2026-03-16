import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { apiGetLostItem } from '../services/publicApi.js'

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

      <div className="card">
        {error ? <div className="alert alert-danger">{error}</div> : null}
        {loading ? (
          <p className="muted">正在加载…</p>
        ) : data ? (
          <>
            <h2 className="card-title">{data.title || '未命名记录'}</h2>
            {data.itemType || data.type ? (
              <p className="muted">类型：{data.itemType || data.type}</p>
            ) : null}
            {data.status ? <p className="muted">状态：{data.status}</p> : null}
            {data.location ? <p className="muted">地点：{data.location}</p> : null}
            {data.occurredAt ? (
              <p className="muted">时间：{data.occurredAt}</p>
            ) : null}
            {data.contact ? <p className="muted">联系方式：{data.contact}</p> : null}
            <div style={{ marginTop: 10 }}>
              <p className="muted">{data.description || '暂无描述'}</p>
            </div>
          </>
        ) : (
          <p className="muted">暂无数据。</p>
        )}
      </div>

      <Link className="btn btn-secondary" to="/lost-found">
        返回列表
      </Link>
    </div>
  )
}

export default LostItemDetailPage
