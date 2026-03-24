import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import Card from '../components/ui/Card.jsx'
import Badge from '../components/ui/Badge.jsx'
import EmptyState from '../components/ui/EmptyState.jsx'
import LoadingCard from '../components/ui/LoadingCard.jsx'
import { useToast } from '../components/ui/Toast.jsx'
import { apiGetActivity } from '../services/publicApi.js'

function UserCenterPage() {
  const navigate = useNavigate()
  const { addToast } = useToast()
  const [registeredActivities, setRegisteredActivities] = useState([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    async function fetchRegisteredActivities() {
      const ids = JSON.parse(localStorage.getItem('registered_activities') || '[]')
      if (ids.length === 0) {
        setRegisteredActivities([])
        return
      }

      setLoading(true)
      try {
        const promises = ids.map((id) => apiGetActivity(id))
        const results = await Promise.all(promises)
        setRegisteredActivities(results)
      } catch (err) {
        console.error('Failed to fetch registered activities:', err)
        addToast('加载报名活动失败', 'danger')
      } finally {
        setLoading(false)
      }
    }

    fetchRegisteredActivities()
  }, [addToast])

  const handleLogout = () => {
    localStorage.removeItem('admin_token')
    addToast('已退出登录', 'success')
    navigate('/', { replace: true })
  }

  const handleToAdmin = () => {
    navigate('/admin')
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
            <span className="value">管理员 (Admin)</span>
          </div>
          <div className="info-item">
            <span className="label">角色：</span>
            <span className="value">超级管理员</span>
          </div>
          <div className="info-item">
            <span className="label">状态：</span>
            <span className="value text-success">已登录</span>
          </div>
        </div>

        <div className="actions mt-6" style={{ display: 'flex', gap: '1rem' }}>
          <button className="btn" onClick={handleToAdmin}>
            进入管理后台
          </button>
          <button className="btn btn-outline" onClick={handleLogout}>
            退出登录
          </button>
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
    </div>
  )
}

export default UserCenterPage
