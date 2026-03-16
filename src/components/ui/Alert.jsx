function Alert({ variant = 'info', className = '', ...props }) {
  return (
    <div
      className={['alert', `alert-${variant}`, className].filter(Boolean).join(' ')}
      {...props}
    />
  )
}

export default Alert
