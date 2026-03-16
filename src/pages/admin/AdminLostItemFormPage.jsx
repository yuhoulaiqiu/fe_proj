import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  apiAdminCreateLostItem,
  apiAdminGetLostItem,
  apiAdminUpdateLostItem,
} from '../../services/adminApi.js'

function AdminLostItemFormPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const isEdit = Boolean(id)

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [title, setTitle] = useState('')
  const [itemType, setItemType] = useState('lost')
  const [status, setStatus] = useState('open')
  const [location, setLocation] = useState('')
  const [occurredAt, setOccurredAt] = useState('')
  const [description, setDescription] = useState('')
  const [contact, setContact] = useState('')

  const pageTitle = useMemo(() => (isEdit ? '编辑记录' : '新增记录'), [isEdit])

  useEffect(() => {
    if (!isEdit) return
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const res = await apiAdminGetLostItem(id)
        if (cancelled) return
        setTitle(res.title || '')
        setItemType(res.itemType || 'lost')
        setStatus(res.status || 'open')
        setLocation(res.location || '')
        setOccurredAt(res.occurredAt || '')
        setDescription(res.description || '')
        setContact(res.contact || '')
      } catch (err) {
        if (cancelled) return
        if (err?.response?.status === 401) {
          localStorage.removeItem('admin_token')
          navigate('/admin/login', { replace: true })
          return
        }
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
  }, [id, isEdit, navigate])

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    if (!title.trim()) {
      setError('标题不能为空')
      return
    }
    if (!contact.trim()) {
      setError('联系方式不能为空')
      return
    }
    setSubmitting(true)
    try {
      const payload = {
        title: title.trim(),
        itemType,
        status,
        location: location.trim(),
        occurredAt: occurredAt.trim(),
        description: description.trim(),
        contact: contact.trim(),
      }
      if (isEdit) {
        await apiAdminUpdateLostItem(id, payload)
      } else {
        await apiAdminCreateLostItem(payload)
      }
      navigate('/admin/lost-items', { replace: true })
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg =
        err?.response?.data?.message || err?.message || '提交失败，请稍后重试'
      setError(msg)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">{pageTitle}</h1>
        <p className="muted">用于发布与维护失物/招领信息。</p>
      </div>

      {error ? <div className="alert alert-danger">{error}</div> : null}

      <form className="card form" onSubmit={onSubmit}>
        <label className="field">
          <span className="label">标题</span>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="例如：地铁站捡到一串钥匙"
            disabled={loading || submitting}
          />
        </label>
        <label className="field">
          <span className="label">类型</span>
          <select
            value={itemType}
            onChange={(e) => setItemType(e.target.value)}
            disabled={loading || submitting}
          >
            <option value="lost">失物</option>
            <option value="found">招领</option>
          </select>
        </label>
        <label className="field">
          <span className="label">状态</span>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            disabled={loading || submitting}
          >
            <option value="open">未处理</option>
            <option value="claimed">已认领</option>
            <option value="returned">已归还</option>
          </select>
        </label>
        <label className="field">
          <span className="label">地点</span>
          <input
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="例如：A区门口"
            disabled={loading || submitting}
          />
        </label>
        <label className="field">
          <span className="label">时间</span>
          <input
            value={occurredAt}
            onChange={(e) => setOccurredAt(e.target.value)}
            placeholder="例如：2026-03-16 14:30"
            disabled={loading || submitting}
          />
        </label>
        <label className="field">
          <span className="label">描述</span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="补充物品特征与领取方式"
            disabled={loading || submitting}
          />
        </label>
        <label className="field">
          <span className="label">联系方式</span>
          <input
            value={contact}
            onChange={(e) => setContact(e.target.value)}
            placeholder="例如：张三 138****0000"
            disabled={loading || submitting}
          />
        </label>

        <div className="actions">
          <button className="btn" type="submit" disabled={loading || submitting}>
            {submitting ? '提交中…' : isEdit ? '保存修改' : '发布'}
          </button>
          <Link className="btn btn-secondary" to="/admin/lost-items">
            返回列表
          </Link>
        </div>
      </form>
    </div>
  )
}

export default AdminLostItemFormPage
