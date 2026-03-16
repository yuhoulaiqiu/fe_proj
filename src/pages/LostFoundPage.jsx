import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGetLostItems } from '../services/publicApi.js'

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

      <form className="card" onSubmit={onSearch}>
        <div className="row-between">
          <label className="field" style={{ flex: 1 }}>
            <span className="label">类型</span>
            <select value={typeInput} onChange={(e) => setTypeInput(e.target.value)}>
              <option value="">全部</option>
              <option value="lost">失物</option>
              <option value="found">招领</option>
            </select>
          </label>
          <label className="field" style={{ flex: 1 }}>
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
          <label className="field" style={{ flex: 2 }}>
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="例如：钥匙、手机、雨伞"
            />
          </label>
          <div className="actions" style={{ alignSelf: 'end' }}>
            <button className="btn" type="submit" disabled={loading}>
              {loading ? '加载中…' : '搜索'}
            </button>
          </div>
        </div>
        {error ? (
          <div className="alert alert-danger" style={{ marginTop: 12 }}>
            {error}
          </div>
        ) : null}
      </form>

      {items.length ? (
        items.map((it) => (
          <div className="card" key={it.id}>
            <div className="row-between">
              <div>
                <h2 className="card-title">{it.title || '未命名记录'}</h2>
                <p className="muted">
                  {(it.itemType && `类型：${it.itemType}`) ||
                    (it.type && `类型：${it.type}`) ||
                    '类型：-'}
                  {it.status ? `｜状态：${it.status}` : ''}
                  {it.location ? `｜地点：${it.location}` : ''}
                </p>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <Link className="btn btn-secondary" to={`/lost-found/${it.id}`}>
                查看详情
              </Link>
            </div>
          </div>
        ))
      ) : loading ? (
        <div className="card">
          <p className="muted">正在加载列表…</p>
        </div>
      ) : (
        <div className="card">
          <p className="muted">暂无数据。</p>
        </div>
      )}
    </div>
  )
}

export default LostFoundPage
