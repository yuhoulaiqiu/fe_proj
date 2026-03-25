import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'

function SiteLayout() {
  const [isMenuOpen, setIsMenuOpen] = useState(false)

  const toggleMenu = () => setIsMenuOpen(!isMenuOpen)
  const closeMenu = () => setIsMenuOpen(false)

  return (
    <div className="site">
      <header className="site-header">
        <div className="container header-inner">
          <NavLink to="/" className="brand" onClick={closeMenu}>
            <div className="logo-placeholder">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                <polyline points="9 22 9 12 15 12 15 22" />
              </svg>
            </div>
            <span>社区互助</span>
          </NavLink>

          <button
            className="menu-toggle"
            onClick={toggleMenu}
            aria-label="切换菜单"
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              {isMenuOpen ? (
                <path d="M18 6L6 18M6 6l12 12" />
              ) : (
                <path d="M3 12h18M3 6h18M3 18h18" />
              )}
            </svg>
          </button>

          <nav className={`nav ${isMenuOpen ? 'open' : ''}`}>
            <NavLink
              to="/activities"
              className="nav-link"
              onClick={closeMenu}
            >
              公益活动
            </NavLink>
            <NavLink
              to="/services"
              className="nav-link"
              onClick={closeMenu}
            >
              便民服务
            </NavLink>
            <NavLink
              to="/lost-found"
              className="nav-link"
              onClick={closeMenu}
            >
              失物招领
            </NavLink>
            <NavLink
              to="/user-center"
              className="nav-link"
              onClick={closeMenu}
            >
              个人中心
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
