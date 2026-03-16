import { NavLink, Outlet } from 'react-router-dom'

function SiteLayout() {
  return (
    <div className="site">
      <header className="site-header">
        <div className="container header-inner">
          <NavLink to="/" className="brand">
            社区互助
          </NavLink>
          <nav className="nav">
            <NavLink to="/activities" className="nav-link">
              公益活动
            </NavLink>
            <NavLink to="/services" className="nav-link">
              便民服务
            </NavLink>
            <NavLink to="/lost-found" className="nav-link">
              失物招领
            </NavLink>
            <NavLink to="/admin" className="nav-link">
              后台
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="site-main">
        <div className="container">
          <Outlet />
        </div>
      </main>
      <footer className="site-footer">
        <div className="container footer-inner">
          <span>社区便民公益互助平台</span>
          <span className="muted">仅用于学习与演示</span>
        </div>
      </footer>
    </div>
  )
}

export default SiteLayout
