import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  apiAdminDeleteLostItem,
  apiAdminListLostItems,
  apiAdminUpdateLostItem,
} from '../../services/adminApi.js'

function AdminLostItemsPage() {
  const navigate = useNavigate()
  const [typeInput, setTypeInput] = useState('')
  const [statusInput, setStatusInput] = useState('')
  const [keywordInput, setKeywordInput] = useState('')
  const [query, setQuery] = useState({ type: '', status: '', keyword: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [items, setItems] = useState([])
  const [total, setTotal] = useState(0)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await apiAdminListLostItems({
        type: query.type || undefined,
        status: query.status || undefined,
        keyword: query.keyword || undefined,
        page: 1,
        pageSize: 50,
      })
      setItems(res.items || [])
      setTotal(res.total || 0)
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg =
        err?.response?.data?.message || err?.message || '加载失败，请稍后重试'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }, [navigate, query.keyword, query.status, query.type])

  useEffect(() => {
    load()
  }, [load])

  function onSearch(e) {
    e.preventDefault()
    setQuery({
      type: typeInput,
      status: statusInput,
      keyword: keywordInput.trim(),
    })
  }

  async function onDelete(id) {
    const ok = window.confirm('确认删除该记录？')
    if (!ok) return
    setError('')
    try {
      await apiAdminDeleteLostItem(id)
      await load()
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg =
        err?.response?.data?.message || err?.message || '删除失败，请稍后重试'
      setError(msg)
    }
  }

  async function onStatusChange(it, nextStatus) {
    setError('')
    try {
      await apiAdminUpdateLostItem(it.id, { ...it, status: nextStatus })
      await load()
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg =
        err?.response?.data?.message || err?.message || '更新失败，请稍后重试'
      setError(msg)
    }
  }

  return (
    <div className="stack">
      <div className="row-between">
        <div className="page-header">
          <h1 className="page-title">失物招领管理</h1>
          <p className="muted">
            新增、编辑、删除与更新状态{total ? `（共 ${total} 条）` : ''}。
          </p>
        </div>
        <Link className="btn" to="/admin/lost-items/new">
          新增记录
        </Link>
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
              placeholder="标题或描述关键词"
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
              <div style={{ flex: 1 }}>
                <h2 className="card-title">{it.title || '未命名记录'}</h2>
                <p className="muted">
                  {(it.itemType && `类型：${it.itemType}`) || '类型：-'}
                  {it.location ? `｜地点：${it.location}` : ''}
                  {it.occurredAt ? `｜时间：${it.occurredAt}` : ''}
                </p>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <div className="stack" style={{ minWidth: 240 }}>
                <label className="field">
                  <span className="label">状态</span>
                  <select
                    value={it.status || 'open'}
                    onChange={(e) => onStatusChange(it, e.target.value)}
                  >
                    <option value="open">未处理</option>
                    <option value="claimed">已认领</option>
                    <option value="returned">已归还</option>
                  </select>
                </label>
                <div className="actions">
                  <Link className="btn btn-secondary" to={`/admin/lost-items/${it.id}/edit`}>
                    编辑
                  </Link>
                  <button
                    className="btn btn-danger"
                    type="button"
                    onClick={() => onDelete(it.id)}
                  >
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        ))
      ) : loading ? (
        <div className="card">
          <p className="muted">正在加载…</p>
        </div>
      ) : (
        <div className="card">
          <p className="muted">暂无数据。</p>
        </div>
      )}
    </div>
  )
}

export default AdminLostItemsPage
