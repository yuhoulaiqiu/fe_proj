import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiCreateActivity } from '../services/publicApi.js'

function ActivityCreatePage() {
  const navigate = useNavigate()
  const { addToast } = useToast()

  const [title, setTitle] = useState('')
  const [category, setCategory] = useState('垃圾分类')
  const [status, setStatus] = useState('active')
  const [location, setLocation] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [coverUrl, setCoverUrl] = useState('')
  const [summary, setSummary] = useState('')
  const [content, setContent] = useState('')

  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const canSubmit = useMemo(() => {
    if (!title.trim()) return false
    if (!location.trim()) return false
    if (!startTime.trim()) return false
    if (!endTime.trim()) return false
    return true
  }, [title, location, startTime, endTime])

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    if (!canSubmit) {
      setError('请填写活动标题、地点、开始时间与结束时间')
      return
    }

    setSubmitting(true)
    try {
      const created = await apiCreateActivity({
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
      addToast('发布成功', 'success')
      navigate(`/activities/${created?.id}`, { replace: true })
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '发布失败，请稍后重试'
      setError(msg)
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
            <h1 className="page-title">发布活动</h1>
            <p className="muted">填写活动信息后即可发布到活动列表。</p>
          </div>
          <Link className="btn btn-secondary" to="/activities">
            返回列表
          </Link>
        </div>
      </div>

      <Card as="form" className="form" onSubmit={onSubmit}>
        <label className="field">
          <span className="label">活动标题</span>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="例如：周末社区垃圾分类宣传"
          />
        </label>

        <div className="grid2">
          <label className="field">
            <span className="label">活动类型</span>
            <select value={category} onChange={(e) => setCategory(e.target.value)}>
              <option value="垃圾分类">垃圾分类</option>
              <option value="健康义诊">健康义诊</option>
              <option value="敬老关爱">敬老关爱</option>
              <option value="社区建设">社区建设</option>
              <option value="其他">其他</option>
            </select>
          </label>
          <label className="field">
            <span className="label">活动状态</span>
            <select value={status} onChange={(e) => setStatus(e.target.value)}>
              <option value="active">报名中</option>
              <option value="finished">已结束</option>
            </select>
          </label>
        </div>

        <label className="field">
          <span className="label">地点</span>
          <input
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="例如：社区中心广场"
          />
        </label>

        <div className="grid2">
          <label className="field">
            <span className="label">开始时间</span>
            <input
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
              placeholder="例如：2026-03-23 09:30:00"
            />
          </label>
          <label className="field">
            <span className="label">结束时间</span>
            <input
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
              placeholder="例如：2026-03-23 11:30:00"
            />
          </label>
        </div>

        <label className="field">
          <span className="label">封面图片链接（可选）</span>
          <input
            value={coverUrl}
            onChange={(e) => setCoverUrl(e.target.value)}
            placeholder="https://..."
          />
        </label>

        <label className="field">
          <span className="label">活动简介（可选）</span>
          <textarea
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            placeholder="一句话说明活动内容与参与方式"
            rows={3}
          />
        </label>

        <label className="field">
          <span className="label">活动详情（可选）</span>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="填写更详细的活动介绍、流程与注意事项"
            rows={8}
          />
        </label>

        {error ? <Alert variant="danger">{error}</Alert> : null}

        <div className="actions" style={{ display: 'flex', gap: '12px' }}>
          <button className="btn" type="submit" disabled={submitting}>
            {submitting ? '发布中…' : '发布活动'}
          </button>
          <button
            className="btn btn-secondary"
            type="button"
            disabled={submitting}
            onClick={() => navigate('/user-center')}
          >
            去个人中心
          </button>
        </div>
      </Card>
    </div>
  )
}

export default ActivityCreatePage
