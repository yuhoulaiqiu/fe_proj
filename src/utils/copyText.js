export async function copyText(text) {
  if (text === null || text === undefined) return false

  const value = String(text)

  try {
    if (window.isSecureContext && navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(value)
      return true
    }
  } catch {}

  let textarea
  try {
    textarea = document.createElement('textarea')
    textarea.value = value
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.top = '-1000px'
    textarea.style.left = '-1000px'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.focus()
    textarea.select()
    textarea.setSelectionRange(0, textarea.value.length)
    const ok = document.execCommand('copy')
    return Boolean(ok)
  } catch {
    return false
  } finally {
    if (textarea?.parentNode) textarea.parentNode.removeChild(textarea)
  }
}
