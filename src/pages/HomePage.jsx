import { Link } from 'react-router-dom'
import hero from '../assets/hero.png'

function HomePage() {
  return (
    <div className="stack">
      <section className="hero-card">
        <div className="hero">
          <div className="stack">
            <h1 className="page-title">邻里互助，温暖社区</h1>
            <p className="lead">
              为您提供公益活动、便民服务目录与失物招领信息，让社区生活更美好，信息更触达。
            </p>
            <div className="actions">
              <Link className="btn" to="/activities">
                参与公益活动
              </Link>
              <Link className="btn btn-secondary" to="/lost-found">
                寻找失物招领
              </Link>
            </div>
          </div>
          <div className="hero-media-placeholder">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
              <path d="M16 3.13a4 4 0 0 1 0 7.75" />
            </svg>
          </div>
        </div>
      </section>

      <section className="grid3">
        <Link className="card" to="/activities">
          <div className="card-icon card-icon-heart">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
            </svg>
          </div>
          <h2 className="card-title">公益活动</h2>
          <p className="muted">社区志愿服务、公益宣传与活动报名信息。</p>
        </Link>
        <Link className="card" to="/services">
          <div className="card-icon card-icon-tools">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
            </svg>
          </div>
          <h2 className="card-title">便民服务</h2>
          <p className="muted">维修、家政、医疗、办事指南等服务目录。</p>
        </Link>
        <Link className="card" to="/lost-found">
          <div className="card-icon card-icon-envelope">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z" />
              <polyline points="22,6 12,13 2,6" />
            </svg>
          </div>
          <h2 className="card-title">失物招领</h2>
          <p className="muted">发布与查询失物/招领信息，促进物品快速归还。</p>
        </Link>
      </section>
    </div>
  )
}

export default HomePage
