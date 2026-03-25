import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import Card from '../components/ui/Card.jsx'
import Badge from '../components/ui/Badge.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import {
  apiDeleteActivity,
  apiGetNotifications,
  apiGetUserPublishedActivities,
  apiGetUserRegisteredActivities,
  apiMarkNotificationRead,
} from '../services/publicApi.js'
import { apiMe } from '../services/authApi.js'

function UserCenterPage() {
  const navigate = useNavigate()
  const { addToast } = useToast()
  const [user, setUser] = useState(() => {
    try {
      return JSON.parse(localStorage.getItem('auth_user') || '{}')
    } catch {
      return {}
    }
  })
  const [registeredActivities, setRegisteredActivities] = useState([])
  const [publishedActivities, setPublishedActivities] = useState([])
  const [notifications, setNotifications] = useState([])
  const [loading, setLoading] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)

  useEffect(() => {
    async function fetchRegisteredActivities() {
      const token = localStorage.getItem('auth_token') || localStorage.getItem('admin_token')
      if (!token) {
        setRegisteredActivities([])
        return
      }

      setLoading(true)
      try {
        const me = await apiMe()
        const nextUser = me?.user || {}
        localStorage.setItem('auth_user', JSON.stringify(nextUser))
        setUser(nextUser)
        const [registered, published, notifyRes] = await Promise.all([
          apiGetUserRegisteredActivities(),
          apiGetUserPublishedActivities(),
          apiGetNotifications({ page: 1, pageSize: 20 }),
        ])
        setRegisteredActivities(Array.isArray(registered) ? registered : [])
        setPublishedActivities(Array.isArray(published) ? published : [])
        setNotifications(Array.isArray(notifyRes?.items) ? notifyRes.items : [])
      } catch (err) {
        addToast('加载个人活动失败', 'danger')
      } finally {
        setLoading(false)
      }
    }

    fetchRegisteredActivities()
  }, [addToast])

  const refreshPublished = async () => {
    try {
      const published = await apiGetUserPublishedActivities()
      setPublishedActivities(Array.isArray(published) ? published : [])
    } catch {
      addToast('刷新我发布的活动失败', 'danger')
    }
  }

  const refreshNotifications = async () => {
    try {
      const notifyRes = await apiGetNotifications({ page: 1, pageSize: 20 })
      setNotifications(Array.isArray(notifyRes?.items) ? notifyRes.items : [])
    } catch {
      addToast('刷新提醒失败', 'danger')
    }
  }

  const handleLogout = () => {
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    localStorage.removeItem('admin_token')
    addToast('已退出登录', 'success')
    navigate('/', { replace: true })
  }

  const handleToAdmin = () => {
    navigate('/admin')
  }

  const handleDeleteActivity = async (activityId) => {
    const ok = window.confirm('确认删除该活动吗？删除后将不再出现在列表中。')
    if (!ok) return
    setActionLoading(true)
    try {
      await apiDeleteActivity(activityId)
      addToast('已删除', 'success')
      await refreshPublished()
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '删除失败，请稍后重试'
      addToast(msg, 'danger')
    } finally {
      setActionLoading(false)
    }
  }

  const handleMarkRead = async (notificationId) => {
    setActionLoading(true)
    try {
      await apiMarkNotificationRead(notificationId)
      await refreshNotifications()
    } catch (err) {
      const msg = err?.response?.data?.message || err?.message || '操作失败，请稍后重试'
      addToast(msg, 'danger')
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <div className="stack">
      <div className="page-header">
        <h1 className="page-title">个人中心</h1>
        <p className="muted">管理您的个人信息和账户设置。</p>
      </div>

      <Card>
        <div className="user-info">
          <div className="info-item">
            <span className="label">用户名：</span>
            <span className="value">{user?.username || '-'}</span>
          </div>
          <div className="info-item">
            <span className="label">角色：</span>
            <span className="value">{user?.role || '-'}</span>
          </div>
          <div className="info-item">
            <span className="label">状态：</span>
            <span className="value text-success">
              {localStorage.getItem('auth_token') || localStorage.getItem('admin_token') ? '已登录' : '未登录'}
            </span>
          </div>
        </div>

        <div className="actions mt-6" style={{ display: 'flex', gap: '1rem' }}>
          {user?.role === 'admin' ? (
            <button className="btn" onClick={handleToAdmin}>
              进入管理后台
            </button>
          ) : null}
          {localStorage.getItem('auth_token') || localStorage.getItem('admin_token') ? (
            <>
              <button className="btn" onClick={() => navigate('/activities/new')}>
                发布活动
              </button>
              <button className="btn btn-outline" onClick={handleLogout}>
                退出登录
              </button>
            </>
          ) : (
            <>
              <button className="btn" onClick={() => navigate('/login')}>
                去登录
              </button>
              <button className="btn btn-outline" onClick={() => navigate('/register')}>
                去注册
              </button>
            </>
          )}
        </div>
      </Card>

      <Card>
        <h2 className="card-title">我报名的活动</h2>
        <div className="mt-4">
          {loading ? (
            <LoadingCard title="正在加载已报名活动..." lines={3} />
          ) : registeredActivities.length > 0 ? (
            <div className="stack" style={{ gap: '12px' }}>
              {registeredActivities.map((activity) => (
                <div
                  key={activity.id}
                  className="row-between"
                  style={{
                    padding: '12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                  }}
                >
                  <div>
                    <h3 style={{ fontSize: '16px', fontWeight: 'bold' }}>
                      {activity.title}
                    </h3>
                    <div className="muted" style={{ fontSize: '14px', marginTop: '4px' }}>
                      {activity.startTime} ~ {activity.endTime}
                    </div>
                  </div>
                  <div className="row" style={{ gap: '12px' }}>
                    <Badge variant={activity.status === 'active' ? 'success' : 'warning'}>
                      {activity.status === 'active' ? '进行中' : '已结束'}
                    </Badge>
                    <Link
                      to={`/activities/${activity.id}`}
                      className="btn btn-secondary"
                      style={{ padding: '4px 12px', fontSize: '14px' }}
                    >
                      查看详情
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState description="您还没有报名任何活动" />
          )}
        </div>
      </Card>

      <Card>
        <div className="row-between">
          <h2 className="card-title">我发布的活动</h2>
          <Link className="btn btn-secondary" to="/activities/new">
            发布新活动
          </Link>
        </div>
        <div className="mt-4">
          {loading ? (
            <LoadingCard title="正在加载我发布的活动..." lines={3} />
          ) : publishedActivities.length > 0 ? (
            <div className="stack" style={{ gap: '12px' }}>
              {publishedActivities.map((activity) => (
                <div
                  key={activity.id}
                  className="row-between"
                  style={{
                    padding: '12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                  }}
                >
                  <div>
                    <h3 style={{ fontSize: '16px', fontWeight: 'bold' }}>
                      {activity.title}
                    </h3>
                    <div className="muted" style={{ fontSize: '14px', marginTop: '4px' }}>
                      {activity.startTime} ~ {activity.endTime}
                    </div>
                  </div>
                  <div className="row" style={{ gap: '12px' }}>
                    <Badge variant={activity.status === 'active' ? 'success' : 'warning'}>
                      {activity.status === 'active'
                        ? '报名中'
                        : activity.status === 'closed'
                          ? '报名截止'
                          : activity.status === 'cancelled'
                            ? '已取消'
                            : '已结束'}
                    </Badge>
                    <Link
                      to={`/activities/${activity.id}`}
                      className="btn btn-secondary"
                      style={{ padding: '4px 12px', fontSize: '14px' }}
                    >
                      查看详情
                    </Link>
                    <Link
                      to={`/activities/${activity.id}/edit`}
                      className="btn btn-secondary"
                      style={{ padding: '4px 12px', fontSize: '14px' }}
                    >
                      编辑
                    </Link>
                    <Link
                      to={`/activities/${activity.id}/registrations`}
                      className="btn btn-secondary"
                      style={{ padding: '4px 12px', fontSize: '14px' }}
                    >
                      报名管理
                    </Link>
                    <button
                      className="btn btn-secondary"
                      type="button"
                      disabled={actionLoading}
                      style={{ padding: '4px 12px', fontSize: '14px' }}
                      onClick={() => handleDeleteActivity(activity.id)}
                    >
                      删除
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState description="您还没有发布任何活动" />
          )}
        </div>
      </Card>

      <Card>
        <div className="row-between">
          <h2 className="card-title">活动提醒</h2>
          <button className="btn btn-secondary" type="button" onClick={refreshNotifications} disabled={loading}>
            刷新
          </button>
        </div>
        <div className="mt-4">
          {loading ? (
            <LoadingCard title="正在加载提醒..." lines={3} />
          ) : notifications.length ? (
            <div className="stack" style={{ gap: '12px' }}>
              {notifications.map((n) => (
                <div
                  key={n.id}
                  className="row-between"
                  style={{
                    padding: '12px',
                    border: '1px solid var(--border-color)',
                    borderRadius: '8px',
                  }}
                >
                  <div className="grow">
                    <div style={{ fontSize: '16px', fontWeight: 600 }}>{n.title || '提醒'}</div>
                    <div className="muted" style={{ fontSize: '14px', marginTop: '6px', whiteSpace: 'pre-wrap' }}>
                      {n.content || '-'}
                    </div>
                    <div className="muted" style={{ fontSize: '12px', marginTop: '6px' }}>
                      触达时间：{n.scheduledFor || '-'}
                    </div>
                  </div>
                  <div className="row" style={{ gap: '10px', alignItems: 'center' }}>
                    <Badge variant={n.readAt ? 'neutral' : 'warning'}>
                      {n.readAt ? '已读' : '未读'}
                    </Badge>
                    {!n.readAt ? (
                      <button
                        className="btn btn-secondary"
                        type="button"
                        disabled={actionLoading}
                        onClick={() => handleMarkRead(n.id)}
                      >
                        标记已读
                      </button>
                    ) : null}
                    {n.activityId ? (
                      <Link className="btn btn-secondary" to={`/activities/${n.activityId}`}>
                        查看活动
                      </Link>
                    ) : null}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState description="暂无提醒" />
          )}
        </div>
      </Card>
    </div>
  )
}

export default UserCenterPage
