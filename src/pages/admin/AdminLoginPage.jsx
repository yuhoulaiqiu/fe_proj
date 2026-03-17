import { useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import Alert from '../../components/ui/Alert.jsx'
import Card from '../../components/ui/Card.jsx'
import { useToast } from '../../components/ui/Toast.jsx'
import { apiLogin } from '../../services/adminApi.js'

function AdminLoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { addToast } = useToast()
  const redirectTo = useMemo(() => {
    const from = location.state?.from?.pathname
    return from && from.startsWith('/admin') ? from : '/admin'
  }, [location.state])

  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin123')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const res = await apiLogin({ username, password })
      localStorage.setItem('admin_token', res.token)
      addToast('登录成功', 'success')
      navigate(redirectTo, { replace: true })
    } catch (err) {
      const msg =
        err?.response?.data?.message || err?.message || '登录失败，请稍后重试'
      setError(msg)
      addToast(msg, 'danger')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">后台登录</h1>
        <p className="muted">用于管理失物招领信息（CRUD）。</p>
      </div>

      <Card as="form" className="form" onSubmit={onSubmit}>
        <label className="field">
          <span className="label">账号</span>
          <input
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="请输入账号"
            autoComplete="username"
          />
        </label>
        <label className="field">
          <span className="label">密码</span>
          <input
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="请输入密码"
            type="password"
            autoComplete="current-password"
          />
        </label>

        {error ? <Alert variant="danger">{error}</Alert> : null}

        <button className="btn" type="submit" disabled={submitting}>
          {submitting ? '登录中…' : '登录'}
        </button>
      </Card>
    </div>
  )
}

export default AdminLoginPage
