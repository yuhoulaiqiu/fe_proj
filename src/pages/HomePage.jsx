import { Link } from 'react-router-dom'
import hero from '../assets/hero.png'

function HomePage() {
  return (
    <div className="stack">
      <section className="hero-card">
        <div className="hero">
          <div className="stack">
            <h1 className="page-title">社区便民公益互助平台</h1>
            <p className="lead">
              为社区居民提供公益活动、便民服务目录与失物招领信息，帮助信息更快触达需要的人。
            </p>
            <div className="actions">
              <Link className="btn" to="/activities">
                查看公益活动
              </Link>
              <Link className="btn btn-secondary" to="/lost-found">
                浏览失物招领
              </Link>
            </div>
          </div>
          <img
            className="hero-media"
            src={hero}
            alt="社区互助信息展示"
            loading="lazy"
          />
        </div>
      </section>

      <section className="grid3">
        <Link className="card" to="/activities">
          <h2 className="card-title">公益活动</h2>
          <p className="muted">社区志愿服务、公益宣传与活动报名信息。</p>
        </Link>
        <Link className="card" to="/services">
          <h2 className="card-title">便民服务</h2>
          <p className="muted">维修、家政、医疗、办事指南等服务目录。</p>
        </Link>
        <Link className="card" to="/lost-found">
          <h2 className="card-title">失物招领</h2>
          <p className="muted">发布与查询失物/招领信息，促进物品快速归还。</p>
        </Link>
      </section>
    </div>
  )
}

export default HomePage
