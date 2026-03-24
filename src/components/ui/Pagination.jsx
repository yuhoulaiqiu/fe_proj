import React from 'react'

function Pagination({ current, total, pageSize, onChange }) {
  const totalPages = Math.ceil(total / pageSize)
  if (totalPages <= 1) return null

  const pages = []
  for (let i = 1; i <= totalPages; i++) {
    pages.push(i)
  }

  return (
    <div className="pagination row-center mt-6">
      <div className="btn-group">
        <button
          className="btn btn-secondary"
          disabled={current === 1}
          onClick={() => onChange(current - 1)}
        >
          上一页
        </button>
        {pages.map((p) => (
          <button
            key={p}
            className={`btn ${p === current ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => onChange(p)}
          >
            {p}
          </button>
        ))}
        <button
          className="btn btn-secondary"
          disabled={current === totalPages}
          onClick={() => onChange(current + 1)}
        >
          下一页
        </button>
      </div>
    </div>
  )
}

export default Pagination
