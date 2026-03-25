import { Navigate, useLocation } from 'react-router-dom'

function RequireAdmin({ children }) {
  const location = useLocation()
  const token = localStorage.getItem('auth_token') || localStorage.getItem('admin_token')
  const user = (() => {
    try {
      return JSON.parse(localStorage.getItem('auth_user') || '{}')
    } catch {
      return {}
    }
  })()

  if (!token || user?.role !== 'admin') {
    return <Navigate to="/admin/login" replace state={{ from: location }} />
  }

  return children
}

export default RequireAdmin
