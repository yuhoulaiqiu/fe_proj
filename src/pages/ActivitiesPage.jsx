import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { apiGetActivities } from '../services/publicApi.js'

function ActivitiesPage() {
  const [keywordInput, setKeywordInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [items, setItems] = useState([])
  const [total, setTotal] = useState(0)

  useEffect(() => {
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const res = await apiGetActivities({ keyword, page: 1, pageSize: 20 })
        if (cancelled) return
        setItems(res.items || [])
        setTotal(res.total || 0)
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
  }, [keyword])

  function onSearch(e) {
    e.preventDefault()
    setKeyword(keywordInput.trim())
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">公益活动</h1>
        <p className="muted">
          浏览社区公益活动与参与信息{total ? `（共 ${total} 条）` : ''}。
        </p>
      </div>

      <Card as="form" onSubmit={onSearch}>
        <div className="filters">
          <label className="field span-2">
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="例如：垃圾分类、义诊、敬老"
            />
          </label>
          <div className="filters-actions">
            <div className="actions">
              <button className="btn" type="submit" disabled={loading}>
                {loading ? '加载中…' : '搜索'}
              </button>
            </div>
          </div>
        </div>
        {error ? (
          <Alert className="mt-3" variant="danger">
            {error}
          </Alert>
        ) : null}
      </Card>

      {items.length ? (
        items.map((it) => (
          <Card key={it.id}>
            <div className="row-between">
              <div>
                <h2 className="card-title">{it.title || '未命名活动'}</h2>
                <p className="muted">{it.summary || '暂无简介'}</p>
              </div>
              <Link className="btn btn-secondary" to={`/activities/${it.id}`}>
                查看详情
              </Link>
            </div>
          </Card>
        ))
      ) : loading ? (
        <LoadingCard title="正在加载活动列表…" />
      ) : (
        <EmptyState description="暂未找到符合条件的活动，试试调整关键词。" />
      )}
    </div>
  )
}

export default ActivitiesPage
