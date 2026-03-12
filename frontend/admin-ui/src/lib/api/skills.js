import { request } from './client.js'
import { getAuthHeaders } from '../auth.js'

const BASE = '/api/v1/admin'

export const skillsApi = {
  list: () => request('/skills'),
  get: (id) => request(`/skills/${id}`),
  create: (sk) => request('/skills', { method: 'POST', body: JSON.stringify(sk) }),
  update: (id, sk) => request(`/skills/${id}`, { method: 'PUT', body: JSON.stringify(sk) }),
  delete: (id) => request(`/skills/${id}`, { method: 'DELETE' }),
  uploadReference: async (id, file) => {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${BASE}/skills/${id}/references`, {
      method: 'POST',
      headers: { ...getAuthHeaders() },
      body: form,
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
    return data
  },
  deleteReference: (id, filename) => request(`/skills/${id}/references/${encodeURIComponent(filename)}`, { method: 'DELETE' }),
  referenceUrl: (id, filename) => `${BASE}/skills/${id}/references/${encodeURIComponent(filename)}`,
  uploadPackage: async (id, file) => {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(`${BASE}/skills/${id}/package`, {
      method: 'POST',
      headers: { ...getAuthHeaders() },
      body: form,
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
    return data
  },
}
