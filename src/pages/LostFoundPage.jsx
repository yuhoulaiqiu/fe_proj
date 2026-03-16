import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { apiGetLostItems } from '../services/publicApi.js'

const TYPE_LABEL = { lost: '失物', found: '招领' }
const STATUS_LABEL = { open: '未处理', claimed: '已认领', returned: '已归还' }
const STATUS_BADGE = { open: 'warning', claimed: 'neutral', returned: 'success' }

function LostFoundPage() {
  const [typeInput, setTypeInput] = useState('')
  const [statusInput, setStatusInput] = useState('')
  const [keywordInput, setKeywordInput] = useState('')
  const [query, setQuery] = useState({ type: '', status: '', keyword: '' })
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
        const res = await apiGetLostItems({
          type: query.type || undefined,
          status: query.status || undefined,
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
  }, [query.keyword, query.status, query.type])

  function onSearch(e) {
    e.preventDefault()
    setQuery({
      type: typeInput,
      status: statusInput,
      keyword: keywordInput.trim(),
    })
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">失物招领</h1>
        <p className="muted">
          浏览失物/招领信息，支持筛选与搜索{total ? `（共 ${total} 条）` : ''}。
        </p>
      </div>

      <Card as="form" onSubmit={onSearch}>
        <div className="filters">
          <label className="field">
            <span className="label">类型</span>
            <select value={typeInput} onChange={(e) => setTypeInput(e.target.value)}>
              <option value="">全部</option>
              <option value="lost">失物</option>
              <option value="found">招领</option>
            </select>
          </label>
          <label className="field">
            <span className="label">状态</span>
            <select
              value={statusInput}
              onChange={(e) => setStatusInput(e.target.value)}
            >
              <option value="">全部</option>
              <option value="open">未处理</option>
              <option value="claimed">已认领</option>
              <option value="returned">已归还</option>
            </select>
          </label>
          <label className="field span-2">
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="例如：钥匙、手机、雨伞"
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
                <h2 className="card-title">{it.title || '未命名记录'}</h2>
                <div className="chips">
                  <Badge variant="neutral">
                    {`类型：${TYPE_LABEL[it.type] || it.itemType || it.type || '-'}`}
                  </Badge>
                  {it.status ? (
                    <Badge variant={STATUS_BADGE[it.status] || 'neutral'}>
                      {`状态：${STATUS_LABEL[it.status] || it.status}`}
                    </Badge>
                  ) : null}
                  {it.location ? <Badge variant="neutral">{`地点：${it.location}`}</Badge> : null}
                </div>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <Link className="btn btn-secondary" to={`/lost-found/${it.id}`}>
                查看详情
              </Link>
            </div>
          </Card>
        ))
      ) : loading ? (
        <LoadingCard title="正在加载列表…" />
      ) : (
        <EmptyState description="暂未找到符合条件的记录，试试调整筛选条件。" />
      )}
    </div>
  )
}

export default LostFoundPage
