import { NavLink, Outlet } from 'react-router-dom'

function AdminLayout() {
  return (
    <div className="admin">
      <header className="admin-header">
        <div className="container header-inner">
          <span className="brand">后台管理</span>
          <div className="header-actions">
            <NavLink to="/" className="btn btn-outline">
              返回首页
            </NavLink>
          </div>
        </div>
      </header>
      <div className="admin-body container">
        <aside className="admin-sidebar">
          <NavLink to="/admin/lost-items" className="nav-link">
            失物招领管理
          </NavLink>
        </aside>
        <main className="admin-main">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export default AdminLayout
