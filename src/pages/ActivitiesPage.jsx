import { useEffect, useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Badge from '../components/ui/Badge.jsx'
import Card from '../components/ui/Card.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import Pagination from '../components/ui/Pagination.jsx'
import { apiGetActivities } from '../services/publicApi.js'

function ActivitiesPage() {
  const [keywordInput, setKeywordInput] = useState('')
  const [keyword, setKeyword] = useState('')
  const [category, setCategory] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 5 // 设置较小的 pageSize 以便于演示分页

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
        const res = await apiGetActivities({
          keyword,
          category,
          status,
          page,
          pageSize,
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
  }, [keyword, category, status, page])

  // 自动触发搜索：当类别或状态改变时，重置页码并自动执行
  const handleCategoryChange = (e) => {
    setCategory(e.target.value)
    setPage(1)
  }

  const handleStatusChange = (e) => {
    setStatus(e.target.value)
    setPage(1)
  }

  function onSearch(e) {
    e.preventDefault()
    setKeyword(keywordInput.trim())
    setPage(1)
  }

  // 如果后端不支持分页（返回的数据总量等于 items 长度，且 total 也等于 items 长度），
  // 并且我们请求了分页，但返回了全部数据，则在前端模拟。
  // 注意：这取决于后端实现。如果后端返回了所有数据，我们需要在前端 slice。
  const displayItems = useMemo(() => {
    if (items.length > pageSize) {
      const start = (page - 1) * pageSize
      return items.slice(start, start + pageSize)
    }
    return items
  }, [items, page, pageSize])

  const displayTotal = items.length > pageSize ? items.length : total

  return (
    <div className="stack">
      <div className="page-header">
        <div className="row-between">
          <div>
            <h1 className="page-title">公益活动</h1>
            <p className="muted">
              浏览社区公益活动与参与信息{total ? `（共 ${total} 条）` : ''}。
            </p>
          </div>
          <Link className="btn" to="/activities/new">
            发布活动
          </Link>
        </div>
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
          <label className="field">
            <span className="label">活动类型</span>
            <select value={category} onChange={handleCategoryChange}>
              <option value="">全部类型</option>
              <option value="垃圾分类">垃圾分类</option>
              <option value="健康义诊">健康义诊</option>
              <option value="敬老关爱">敬老关爱</option>
              <option value="社区建设">社区建设</option>
            </select>
          </label>
          <label className="field">
            <span className="label">活动状态</span>
            <select value={status} onChange={handleStatusChange}>
              <option value="">全部状态</option>
              <option value="active">报名中</option>
              <option value="finished">已结束</option>
            </select>
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
          <LoadingCard title="正在加载活动列表…" />
          <LoadingCard />
          <LoadingCard />
        </div>
      ) : displayItems.length ? (
        <>
          <div className="stack">
            {displayItems.map((it) => (
              <Card key={it.id}>
                <div className="row-between">
                  <div className="grow">
                    <div className="chips" style={{ marginBottom: '8px' }}>
                      {it.category && <Badge variant="neutral">{it.category}</Badge>}
                      {it.status && (
                        <Badge variant={it.status === 'active' ? 'success' : 'warning'}>
                          {it.status === 'active' ? '报名中' : '已结束'}
                        </Badge>
                      )}
                    </div>
                    <h2 className="card-title">{it.title || '未命名活动'}</h2>
                    <p className="muted">{it.summary || '暂无简介'}</p>
                  </div>
                  <Link className="btn btn-secondary" to={`/activities/${it.id}`}>
                    查看详情
                  </Link>
                </div>
              </Card>
            ))}
          </div>
          <Pagination
            current={page}
            total={displayTotal}
            pageSize={pageSize}
            onChange={setPage}
          />
        </>
      ) : (
        <EmptyState description="暂未找到符合条件的活动，试试调整关键词或切换筛选条件。" />
      )}
    </div>
  )
}

export default ActivitiesPage
