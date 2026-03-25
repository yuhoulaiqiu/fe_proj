import { http } from './http.js'

export async function apiAdminListLostItems(params) {
  const { data } = await http.get('/api/admin/lost-items', { params })
  return data
}

export async function apiAdminGetLostItem(id) {
  const { data } = await http.get(`/api/admin/lost-items/${id}`)
  return data
}

export async function apiAdminCreateLostItem(payload) {
  const { data } = await http.post('/api/admin/lost-items', payload)
  return data
}

export async function apiAdminUpdateLostItem(id, payload) {
  const { data } = await http.put(`/api/admin/lost-items/${id}`, payload)
  return data
}

export async function apiAdminDeleteLostItem(id) {
  const { data } = await http.delete(`/api/admin/lost-items/${id}`)
  return data
}
