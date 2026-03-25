import { useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiRegister } from '../services/authApi.js'

function RegisterPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { addToast } = useToast()
  const redirectTo = useMemo(() => {
    const from = location.state?.from?.pathname
    return from ? from : '/user-center'
  }, [location.state])

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const res = await apiRegister({ username, password })
      localStorage.setItem('auth_token', res.token)
      localStorage.setItem('auth_user', JSON.stringify(res.user || {}))
      addToast('注册成功', 'success')
      navigate(redirectTo, { replace: true })
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '注册失败，请稍后重试'
      setError(msg)
      addToast(msg, 'danger')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="auth-page-container">
      <div className="stack" style={{ width: '100%', maxWidth: '400px' }}>
        <div className="page-header" style={{ textAlign: 'center' }}>
          <h1 className="page-title">注册</h1>
          <p className="muted">创建账号后即可报名活动。</p>
        </div>

        <Card as="form" className="form" onSubmit={onSubmit} style={{ maxWidth: '100%' }}>
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
              placeholder="至少 6 位"
              type="password"
              autoComplete="new-password"
            />
          </label>

          {error ? <Alert variant="danger">{error}</Alert> : null}

          <button className="btn" type="submit" disabled={submitting}>
            {submitting ? '注册中…' : '注册'}
          </button>

          <div className="muted" style={{ textAlign: 'center', marginTop: '10px' }}>
            已有账号？ <Link to="/login">去登录</Link>
          </div>
        </Card>
      </div>
    </div>
  )
}

export default RegisterPage

