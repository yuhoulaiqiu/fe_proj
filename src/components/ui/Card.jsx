import { createElement } from 'react'

function Card({ as = 'div', className = '', ...props }) {
  return createElement(as, {
    className: ['card', className].filter(Boolean).join(' '),
    ...props,
  })
}

export default Card
