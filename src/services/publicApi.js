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

export async function apiRegisterActivity(id) {
  const { data } = await http.post(`/api/activities/${id}/register`)
  return data
}

export async function apiCancelActivityRegistration(id) {
  const { data } = await http.delete(`/api/activities/${id}/register`)
  return data
}

export async function apiGetUserRegisteredActivities() {
  const { data } = await http.get('/api/user/activities/registered')
  return data
}

export async function apiGetUserPublishedActivities() {
  const { data } = await http.get('/api/user/activities/published')
  return data
}

export async function apiCreateActivity(payload) {
  const { data } = await http.post('/api/activities', payload)
  return data
}

export async function apiUpdateActivity(id, payload) {
  const { data } = await http.put(`/api/activities/${id}`, payload)
  return data
}

export async function apiDeleteActivity(id) {
  const { data } = await http.delete(`/api/activities/${id}`)
  return data
}

export async function apiGetActivityRegistrations(id) {
  const { data } = await http.get(`/api/activities/${id}/registrations`)
  return data
}

export async function apiExportActivityRegistrationsCsv(id) {
  const res = await http.get(`/api/activities/${id}/registrations.csv`, {
    responseType: 'blob',
  })
  return res.data
}

export async function apiGetNotifications(params) {
  const { data } = await http.get('/api/notifications', { params })
  return normalizeList(data)
}

export async function apiMarkNotificationRead(id) {
  const { data } = await http.post(`/api/notifications/${id}/read`)
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
