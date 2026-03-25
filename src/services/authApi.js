import { http } from './http.js'

export async function apiLogin(payload) {
  const { data } = await http.post('/api/auth/login', payload)
  return data
}

export async function apiRegister(payload) {
  const { data } = await http.post('/api/auth/register', payload)
  return data
}

export async function apiMe() {
  const { data } = await http.get('/api/auth/me')
  return data
}

