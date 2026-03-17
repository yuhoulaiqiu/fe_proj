import Card from './Card.jsx'

function LoadingCard({ title, lines = 2 }) {
  return (
    <Card>
      <div className="state">
        <div className="skeleton" style={{ width: '32%', height: 14 }} />
        <div className="skeleton" style={{ width: '55%', height: 12 }} />
        {Array.from({ length: Math.max(0, lines - 1) }).map((_, i) => (
          <div
            key={i}
            className="skeleton"
            style={{ width: `${70 - i * 8}%`, height: 12 }}
          />
        ))}
        {title && <p className="muted">{title}</p>}
      </div>
    </Card>
  )
}

export default LoadingCard
