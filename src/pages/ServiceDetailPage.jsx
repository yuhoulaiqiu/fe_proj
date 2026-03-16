import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
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

      <div className="card">
        {error ? <div className="alert alert-danger">{error}</div> : null}
        {loading ? (
          <p className="muted">正在加载…</p>
        ) : data ? (
          <>
            <h2 className="card-title">{data.name || '未命名服务'}</h2>
            <p className="muted">
              {(data.category && `类别：${data.category}`) || '类别：-'}
            </p>
            {data.phone ? <p className="muted">电话：{data.phone}</p> : null}
            {data.address ? <p className="muted">地址：{data.address}</p> : null}
            <div style={{ marginTop: 10 }}>
              <p className="muted">{data.description || '暂无说明'}</p>
            </div>
          </>
        ) : (
          <p className="muted">暂无数据。</p>
        )}
      </div>

      <Link className="btn btn-secondary" to="/services">
        返回目录
      </Link>
    </div>
  )
}

export default ServiceDetailPage
