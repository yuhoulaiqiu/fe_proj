import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Alert from '../../components/ui/Alert.jsx'
import Badge from '../../components/ui/Badge.jsx'
import Card from '../../components/ui/Card.jsx'
import EmptyState from '../../components/ui/EmptyState.jsx'
import LoadingCard from '../../components/ui/LoadingCard.jsx'
import {
  apiAdminDeleteLostItem,
  apiAdminListLostItems,
  apiAdminUpdateLostItem,
} from '../../services/adminApi.js'

const TYPE_LABEL = { lost: '失物', found: '招领' }
const STATUS_LABEL = { open: '未处理', claimed: '已认领', returned: '已归还' }
const STATUS_BADGE = { open: 'warning', claimed: 'neutral', returned: 'success' }

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
              placeholder="标题或描述关键词"
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
              <div className="grow">
                <h2 className="card-title">{it.title || '未命名记录'}</h2>
                <div className="chips">
                  <Badge variant="neutral">
                    {`类型：${TYPE_LABEL[it.itemType] || it.itemType || '-'}`}
                  </Badge>
                  <Badge variant={STATUS_BADGE[it.status] || 'neutral'}>
                    {`状态：${STATUS_LABEL[it.status] || it.status || '-'}`}
                  </Badge>
                  {it.location ? <Badge variant="neutral">{`地点：${it.location}`}</Badge> : null}
                  {it.occurredAt ? (
                    <Badge variant="neutral">{`时间：${it.occurredAt}`}</Badge>
                  ) : null}
                </div>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <div className="stack minw-240">
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
          </Card>
        ))
      ) : loading ? (
        <LoadingCard title="正在加载…" />
      ) : (
        <EmptyState description="暂无记录，点击“新增记录”发布第一条信息。" />
      )}
    </div>
  )
}

export default AdminLostItemsPage
