import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apiGetServices } from '../services/publicApi.js'

function ServicesPage() {
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

      <form className="card" onSubmit={onSearch}>
        <div className="row-between">
          <label className="field" style={{ flex: 1 }}>
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
          <label className="field" style={{ flex: 2 }}>
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="例如：水电、开锁、医保"
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
                <h2 className="card-title">{it.name || '未命名服务'}</h2>
                <p className="muted">
                  {(it.category && `类别：${it.category}`) || '类别：-'}{' '}
                  {(it.phone && `｜电话：${it.phone}`) || ''}
                </p>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <Link className="btn btn-secondary" to={`/services/${it.id}`}>
                查看详情
              </Link>
            </div>
          </div>
        ))
      ) : loading ? (
        <div className="card">
          <p className="muted">正在加载服务目录…</p>
        </div>
      ) : (
        <div className="card">
          <p className="muted">暂无数据。</p>
        </div>
      )}
    </div>
  )
}

export default ServicesPage
