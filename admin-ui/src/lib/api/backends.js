import { request } from './client.js'

export const backendsApi = {
  list: () => request('/backends'),
  get: (id) => request(`/backends/${id}`),
  create: (b) => request('/backends', { method: 'POST', body: JSON.stringify(b) }),
  update: (id, b) => request(`/backends/${id}`, { method: 'PUT', body: JSON.stringify(b) }),
  delete: (id) => request(`/backends/${id}`, { method: 'DELETE' }),
}
