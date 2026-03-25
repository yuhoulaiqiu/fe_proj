import axios from 'axios'

export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 15000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token') || localStorage.getItem('admin_token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => {
    const contentType = String(res.headers?.['content-type'] || '')
    if (contentType.includes('text/html')) {
      const url = String(res.config?.baseURL || '') + String(res.config?.url || '')
      return Promise.reject(new Error(`API request returned HTML: ${url}`))
    }
    return res
  },
  (err) => Promise.reject(err),
)
