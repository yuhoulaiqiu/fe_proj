import { http } from './http.js'

function normalizeList(data) {
  if (Array.isArray(data)) {
    return { items: data, total: data.length }
  }
  if (data && Array.isArray(data.items)) {
    return { items: data.items, total: data.total ?? data.items.length }
  }
  return { items: [], total: 0 }
}

export async function apiGetActivities(params) {
  const { data } = await http.get('/api/activities', { params })
  return normalizeList(data)
}

export async function apiGetActivity(id) {
  const { data } = await http.get(`/api/activities/${id}`)
  return data
}

export async function apiGetServices(params) {
  const { data } = await http.get('/api/services', { params })
  return normalizeList(data)
}

export async function apiGetService(id) {
  const { data } = await http.get(`/api/services/${id}`)
  return data
}

export async function apiGetLostItems(params) {
  const { data } = await http.get('/api/lost-items', { params })
  return normalizeList(data)
}

export async function apiGetLostItem(id) {
  const { data } = await http.get(`/api/lost-items/${id}`)
  return data
}

