import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiDeleteActivity, apiGetActivity, apiUpdateActivity } from '../services/publicApi.js'

function ActivityEditPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { addToast } = useToast()

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const [data, setData] = useState(null)

  const [title, setTitle] = useState('')
  const [category, setCategory] = useState('垃圾分类')
  const [status, setStatus] = useState('active')
  const [location, setLocation] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [coverUrl, setCoverUrl] = useState('')
  const [summary, setSummary] = useState('')
  const [content, setContent] = useState('')

  const ownerId = useMemo(() => {
    try {
      return JSON.parse(localStorage.getItem('auth_user') || '{}')?.id
    } catch {
      return undefined
    }
  }, [])

  const isOwner = useMemo(() => {
    if (!data) return false
    if (!ownerId) return false
    return String(ownerId) === String(data.userId)
  }, [data, ownerId])

  const canSubmit = useMemo(() => {
    if (!isOwner) return false
    if (!title.trim()) return false
    if (!location.trim()) return false
    if (!startTime.trim()) return false
    if (!endTime.trim()) return false
    return true
  }, [isOwner, title, location, startTime, endTime])

  useEffect(() => {
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const res = await apiGetActivity(id)
        if (cancelled) return
        setData(res)
        setTitle(res?.title || '')
        setCategory(res?.category || '其他')
        setStatus(res?.status || 'active')
        setLocation(res?.location || '')
        setStartTime(res?.startTime || '')
        setEndTime(res?.endTime || '')
        setCoverUrl(res?.coverUrl || '')
        setSummary(res?.summary || '')
        setContent(res?.content || '')
      } catch (err) {
        if (cancelled) return
        const msg = err?.response?.data?.message || err?.message || '加载失败，请稍后重试'
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

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    if (!canSubmit) {
      setError(isOwner ? '请填写活动标题、地点、开始时间与结束时间' : '仅活动发布者可编辑')
      return
    }
    setSubmitting(true)
    try {
      const updated = await apiUpdateActivity(id, {
        title: title.trim(),
        category: category.trim(),
        status,
        coverUrl: coverUrl.trim(),
        summary: summary.trim(),
        content: content.trim(),
        location: location.trim(),
        startTime: startTime.trim(),
        endTime: endTime.trim(),
      })
      addToast('保存成功', 'success')
      navigate(`/activities/${updated?.id || id}`, { replace: true })
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '保存失败，请稍后重试'
      setError(msg)
      addToast(msg, 'danger')
    } finally {
      setSubmitting(false)
    }
  }

  async function onDelete() {
    if (!isOwner) {
      addToast('仅活动发布者可删除', 'warning')
      return
    }
    const ok = window.confirm('确认删除该活动吗？删除后将不再出现在列表中。')
    if (!ok) return
    setSubmitting(true)
    try {
      await apiDeleteActivity(id)
      addToast('已删除', 'success')
      navigate('/user-center', { replace: true })
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '删除失败，请稍后重试'
      addToast(msg, 'danger')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <div className="row-between">
          <div>
            <h1 className="page-title">编辑活动</h1>
            <p className="muted">活动 ID：{id}</p>
          </div>
          <Link className="btn btn-secondary" to={`/activities/${id}`}>
            返回详情
          </Link>
        </div>
      </div>

      {loading ? <LoadingCard title="正在加载活动信息…" lines={5} /> : null}
      {error ? <Alert variant="danger">{error}</Alert> : null}
      {!loading && data && !isOwner ? (
        <Alert variant="warning">仅活动发布者可编辑/删除该活动。</Alert>
      ) : null}

      {!loading && data ? (
        <Card as="form" className="form" onSubmit={onSubmit}>
          <label className="field">
            <span className="label">活动标题</span>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="例如：周末社区垃圾分类宣传"
              disabled={!isOwner || submitting}
            />
          </label>

          <div className="grid2">
            <label className="field">
              <span className="label">活动类型</span>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                disabled={!isOwner || submitting}
              >
                <option value="垃圾分类">垃圾分类</option>
                <option value="健康义诊">健康义诊</option>
                <option value="敬老关爱">敬老关爱</option>
                <option value="社区建设">社区建设</option>
                <option value="其他">其他</option>
              </select>
            </label>
            <label className="field">
              <span className="label">活动状态</span>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                disabled={!isOwner || submitting}
              >
                <option value="active">报名中</option>
                <option value="cancelled">已取消</option>
              </select>
            </label>
          </div>

          <label className="field">
            <span className="label">地点</span>
            <input
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder="例如：社区中心广场"
              disabled={!isOwner || submitting}
            />
          </label>

          <div className="grid2">
            <label className="field">
              <span className="label">开始时间</span>
              <input
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
                placeholder="例如：2026-03-23 09:30:00"
                disabled={!isOwner || submitting}
              />
            </label>
            <label className="field">
              <span className="label">结束时间</span>
              <input
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
                placeholder="例如：2026-03-23 11:30:00"
                disabled={!isOwner || submitting}
              />
            </label>
          </div>

          <label className="field">
            <span className="label">封面图片链接（可选）</span>
            <input
              value={coverUrl}
              onChange={(e) => setCoverUrl(e.target.value)}
              placeholder="https://..."
              disabled={!isOwner || submitting}
            />
          </label>

          <label className="field">
            <span className="label">活动简介（可选）</span>
            <textarea
              value={summary}
              onChange={(e) => setSummary(e.target.value)}
              placeholder="一句话说明活动内容与参与方式"
              rows={3}
              disabled={!isOwner || submitting}
            />
          </label>

          <label className="field">
            <span className="label">活动详情（可选）</span>
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="填写更详细的活动介绍、流程与注意事项"
              rows={8}
              disabled={!isOwner || submitting}
            />
          </label>

          <div className="actions" style={{ display: 'flex', gap: '12px' }}>
            <button className="btn" type="submit" disabled={submitting || !canSubmit}>
              {submitting ? '保存中…' : '保存'}
            </button>
            <button
              className="btn btn-secondary"
              type="button"
              disabled={submitting}
              onClick={onDelete}
            >
              删除活动
            </button>
          </div>
        </Card>
      ) : null}
    </div>
  )
}

export default ActivityEditPage

