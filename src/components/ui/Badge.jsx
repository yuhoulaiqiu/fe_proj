function Badge({ variant = 'neutral', className = '', ...props }) {
  return (
    <span
      className={['badge', `badge-${variant}`, className].filter(Boolean).join(' ')}
      {...props}
    />
  )
}

export default Badge
