import { Link } from 'react-router-dom'

function NotFoundPage() {
  return (
    <div className="stack">
      <h1 className="page-title">页面不存在</h1>
      <p className="muted">你访问的页面可能已被移动或删除。</p>
      <Link className="btn" to="/">
        返回首页
      </Link>
    </div>
  )
}

export default NotFoundPage
