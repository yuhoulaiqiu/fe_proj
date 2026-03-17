import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiGetServices } from '../services/publicApi.js'

function ServicesPage() {
  const { addToast } = useToast()
  const [categoryInput, setCategoryInput] = useState('')
  const [keywordInput, setKeywordInput] = useState('')
  const [query, setQuery] = useState({ category: '', keyword: '' })
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
        const res = await apiGetServices({
          category: query.category || undefined,
          keyword: query.keyword || undefined,
          page: 1,
          pageSize: 30,
        })
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
  }, [query.category, query.keyword])

  const handleCopy = (text) => {
    navigator.clipboard.writeText(text).then(() => {
      addToast('号码已复制', 'success')
    })
  }

  function onSearch(e) {
    e.preventDefault()
    setQuery({ category: categoryInput, keyword: keywordInput.trim() })
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">便民服务</h1>
        <p className="muted">
          按类别与关键词快速查询服务信息{total ? `（共 ${total} 条）` : ''}。
        </p>
      </div>

      <Card as="form" onSubmit={onSearch}>
        <div className="filters">
          <label className="field">
            <span className="label">类别</span>
            <select
              value={categoryInput}
              onChange={(e) => setCategoryInput(e.target.value)}
            >
              <option value="">全部</option>
              <option value="repair">维修</option>
              <option value="housekeeping">家政</option>
              <option value="medical">医疗</option>
              <option value="guide">办事指南</option>
            </select>
          </label>
          <label className="field span-2">
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="例如：水电、开锁、医保"
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

      {loading ? (
        <div className="stack">
          <LoadingCard title="正在加载服务目录…" />
          <LoadingCard />
          <LoadingCard />
        </div>
      ) : items.length ? (
        <div className="stack">
          {items.map((it) => (
            <Card key={it.id}>
              <div className="row-between">
                <div className="grow">
                  <div className="chips" style={{ marginBottom: '8px' }}>
                    {it.category && (
                      <Badge variant="neutral">
                        {it.category === 'repair' ? '维修' :
                         it.category === 'housekeeping' ? '家政' :
                         it.category === 'medical' ? '医疗' :
                         it.category === 'guide' ? '办事指南' : it.category}
                      </Badge>
                    )}
                    {it.phone && (
                      <button
                        className="badge badge-neutral"
                        style={{ cursor: 'pointer', borderStyle: 'dashed' }}
                        onClick={() => handleCopy(it.phone)}
                        title="点击复制号码"
                      >
                        📞 {it.phone} (点击复制)
                      </button>
                    )}
                  </div>
                  <h2 className="card-title">{it.name || '未命名服务'}</h2>
                  {it.description ? <p className="muted">{it.description}</p> : null}
                </div>
                <Link className="btn btn-secondary" to={`/services/${it.id}`}>
                  查看详情
                </Link>
              </div>
            </Card>
          ))}
        </div>
      ) : (
        <EmptyState description="暂未找到符合条件的服务，试试调整筛选条件。" />
      )}
    </div>
  )
}

export default ServicesPage
