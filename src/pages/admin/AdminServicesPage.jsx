import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Alert from '../../components/ui/Alert.jsx'
import Badge from '../../components/ui/Badge.jsx'
import Card from '../../components/ui/Card.jsx'
import EmptyState from '../../components/ui/EmptyState.jsx'
import LoadingCard from '../../components/ui/LoadingCard.jsx'
import { useToast } from '../../components/ui/Toast.jsx'
import { apiAdminDeleteService, apiAdminListServices } from '../../services/adminApi.js'

function AdminServicesPage() {
  const navigate = useNavigate()
  const { addToast } = useToast()

  const [categoryInput, setCategoryInput] = useState('')
  const [keywordInput, setKeywordInput] = useState('')
  const [query, setQuery] = useState({ category: '', keyword: '' })

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [items, setItems] = useState([])
  const [total, setTotal] = useState(0)

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      const res = await apiAdminListServices({
        category: query.category || undefined,
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
      const msg = err?.response?.data?.message || err?.message || '加载失败，请稍后重试'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }, [navigate, query.category, query.keyword])

  useEffect(() => {
    load()
  }, [load])

  function onSearch(e) {
    e.preventDefault()
    setQuery({ category: categoryInput, keyword: keywordInput.trim() })
  }

  async function onDelete(id) {
    const ok = window.confirm('确认删除该服务？')
    if (!ok) return
    setError('')
    try {
      await apiAdminDeleteService(id)
      addToast('服务已删除', 'success')
      await load()
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg = err?.response?.data?.message || err?.message || '删除失败，请稍后重试'
      setError(msg)
      addToast(msg, 'danger')
    }
  }

  return (
    <div className="stack">
      <div className="row-between">
        <div className="page-header">
          <h1 className="page-title">便民服务管理</h1>
          <p className="muted">新增、编辑、删除便民服务{total ? `（共 ${total} 条）` : ''}。</p>
        </div>
        <Link className="btn btn-primary" to="/admin/services/new">
          新增服务
        </Link>
      </div>

      <Card as="form" onSubmit={onSearch}>
        <div className="filters">
          <label className="field">
            <span className="label">分类</span>
            <select value={categoryInput} onChange={(e) => setCategoryInput(e.target.value)}>
              <option value="">全部</option>
              <option value="repair">维修</option>
              <option value="housekeeping">家政</option>
              <option value="guide">指南</option>
              <option value="other">其他</option>
            </select>
          </label>
          <label className="field span-2">
            <span className="label">关键词</span>
            <input
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              placeholder="名称或描述关键词"
            />
          </label>
          <div className="filters-actions">
            <div className="actions">
              <button className="btn btn-primary" type="submit" disabled={loading}>
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
          <LoadingCard title="正在加载…" />
          <LoadingCard />
          <LoadingCard />
        </div>
      ) : items.length ? (
        items.map((it) => (
          <Card key={it.id}>
            <div className="row-between">
              <div className="grow">
                <h2 className="card-title">{it.name || '未命名服务'}</h2>
                <div className="chips">
                  {it.category ? <Badge variant="neutral">{`分类：${it.category}`}</Badge> : null}
                  {it.phone ? <Badge variant="neutral">{`电话：${it.phone}`}</Badge> : null}
                  {it.address ? <Badge variant="neutral">{`地址：${it.address}`}</Badge> : null}
                </div>
                {it.description ? <p className="muted">{it.description}</p> : null}
              </div>
              <div className="stack minw-240">
                <div className="actions">
                  <Link className="btn btn-primary" to={`/admin/services/${it.id}/edit`}>
                    编辑
                  </Link>
                  <button className="btn btn-danger" type="button" onClick={() => onDelete(it.id)}>
                    删除
                  </button>
                </div>
              </div>
            </div>
          </Card>
        ))
      ) : (
        <EmptyState description="暂无服务，点击“新增服务”创建第一条便民服务信息。" />
      )}
    </div>
  )
}

export default AdminServicesPage

