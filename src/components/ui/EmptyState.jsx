import Card from './Card.jsx'

function EmptyState({ title = '暂无数据', description = '请稍后再试或调整筛选条件。', action }) {
  return (
    <Card>
      <div className="state">
        <h3 className="state-title">{title}</h3>
        {description ? <p className="muted">{description}</p> : null}
        {action ? <div className="actions">{action}</div> : null}
      </div>
    </Card>
  )
}

export default EmptyState
