import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import Alert from '../../components/ui/Alert.jsx'
import Card from '../../components/ui/Card.jsx'
import LoadingCard from '../../components/ui/LoadingCard.jsx'
import { useToast } from '../../components/ui/Toast.jsx'
import {
  apiAdminCreateService,
  apiAdminGetService,
  apiAdminUpdateService,
} from '../../services/adminApi.js'

function AdminServiceFormPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { addToast } = useToast()
  const isEdit = Boolean(id)

  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const [name, setName] = useState('')
  const [category, setCategory] = useState('repair')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [description, setDescription] = useState('')

  const pageTitle = useMemo(() => (isEdit ? '编辑服务' : '新增服务'), [isEdit])

  useEffect(() => {
    if (!isEdit) return
    let cancelled = false
    async function run() {
      setLoading(true)
      setError('')
      try {
        const res = await apiAdminGetService(id)
        if (cancelled) return
        setName(res.name || '')
        setCategory(res.category || 'other')
        setPhone(res.phone || '')
        setAddress(res.address || '')
        setDescription(res.description || '')
      } catch (err) {
        if (cancelled) return
        if (err?.response?.status === 401) {
          localStorage.removeItem('admin_token')
          navigate('/admin/login', { replace: true })
          return
        }
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
  }, [id, isEdit, navigate])

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    if (!name.trim()) {
      setError('服务名称不能为空')
      return
    }
    setSubmitting(true)
    try {
      const payload = {
        name: name.trim(),
        category,
        phone: phone.trim(),
        address: address.trim(),
        description: description.trim(),
      }
      if (isEdit) {
        await apiAdminUpdateService(id, payload)
        addToast('服务更新成功', 'success')
      } else {
        await apiAdminCreateService(payload)
        addToast('服务创建成功', 'success')
      }
      navigate('/admin/services', { replace: true })
    } catch (err) {
      if (err?.response?.status === 401) {
        localStorage.removeItem('admin_token')
        navigate('/admin/login', { replace: true })
        return
      }
      const msg = err?.response?.data?.message || err?.message || '提交失败，请稍后重试'
      setError(msg)
      addToast(msg, 'danger')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">{pageTitle}</h1>
        <p className="muted">用于发布与维护便民服务信息。</p>
      </div>

      {error ? <Alert variant="danger">{error}</Alert> : null}
      {loading ? <LoadingCard title="正在加载服务…" lines={2} /> : null}

      <Card as="form" className="form stack" onSubmit={onSubmit}>
        <label className="field">
          <span className="label">服务名称</span>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="例如：社区水电维修"
            disabled={loading || submitting}
          />
        </label>

        <div className="grid2">
          <label className="field">
            <span className="label">分类</span>
            <select value={category} onChange={(e) => setCategory(e.target.value)} disabled={loading || submitting}>
              <option value="repair">维修</option>
              <option value="housekeeping">家政</option>
              <option value="guide">指南</option>
              <option value="other">其他</option>
            </select>
          </label>
          <label className="field">
            <span className="label">电话</span>
            <input
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="例如：400-000-0001"
              disabled={loading || submitting}
            />
          </label>
        </div>

        <label className="field">
          <span className="label">地址</span>
          <input
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="例如：A区物业服务中心"
            disabled={loading || submitting}
          />
        </label>

        <label className="field">
          <span className="label">描述</span>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
            placeholder="服务介绍与预约方式"
            disabled={loading || submitting}
          />
        </label>

        <div className="actions">
          <button className="btn btn-primary" type="submit" disabled={loading || submitting}>
            {submitting ? '提交中…' : isEdit ? '保存修改' : '创建'}
          </button>
          <Link className="btn btn-secondary" to="/admin/services">
            返回列表
          </Link>
        </div>
      </Card>
    </div>
  )
}

export default AdminServiceFormPage

