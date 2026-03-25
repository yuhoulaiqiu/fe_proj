import { useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import Alert from '../components/ui/Alert.jsx'
import Card from '../components/ui/Card.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiLogin, apiRegister } from '../services/authApi.js'

function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { addToast } = useToast()
  const redirectTo = useMemo(() => {
    const from = location.state?.from?.pathname
    if (from && from !== '/login' && from !== '/register') return from
    return '/'
  }, [location.state])

  const [mode, setMode] = useState('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function onSubmit(e) {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const res =
        mode === 'register'
          ? await apiRegister({ username, password })
          : await apiLogin({ username, password })
      localStorage.setItem('auth_token', res.token)
      localStorage.setItem('auth_user', JSON.stringify(res.user || {}))
      if (res?.user?.role === 'admin') {
        localStorage.setItem('admin_token', res.token)
      } else {
        localStorage.removeItem('admin_token')
      }
      addToast(mode === 'register' ? '注册成功' : '登录成功', 'success')
      const next =
        redirectTo.startsWith('/admin') && res?.user?.role !== 'admin' ? '/' : redirectTo
      navigate(next, { replace: true })
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '登录失败，请稍后重试'
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
          <h1 className="page-title">{mode === 'register' ? '注册' : '登录'}</h1>
          <p className="muted">登录/注册后可报名活动并查看个人信息。</p>
        </div>

        <Card as="form" className="form" onSubmit={onSubmit} style={{ maxWidth: '100%' }}>
          <div className="btn-group" style={{ justifyContent: 'center' }}>
            <button
              type="button"
              className={`btn ${mode === 'login' ? '' : 'btn-secondary'}`}
              style={{ padding: '8px 14px' }}
              onClick={() => setMode('login')}
              disabled={submitting}
            >
              登录
            </button>
            <button
              type="button"
              className={`btn ${mode === 'register' ? '' : 'btn-secondary'}`}
              style={{ padding: '8px 14px' }}
              onClick={() => setMode('register')}
              disabled={submitting}
            >
              注册
            </button>
          </div>

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
            {submitting ? (mode === 'register' ? '注册中…' : '登录中…') : mode === 'register' ? '注册' : '登录'}
          </button>
        </Card>
      </div>
    </div>
  )
}

export default LoginPage
