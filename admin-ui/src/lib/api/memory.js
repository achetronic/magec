import { request } from './client.js'

export const memoryApi = {
  list: () => request('/memory'),
  get: (id) => request(`/memory/${id}`),
  create: (m) => request('/memory', { method: 'POST', body: JSON.stringify(m) }),
  update: (id, m) => request(`/memory/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
  delete: (id) => request(`/memory/${id}`, { method: 'DELETE' }),
  checkHealth: (id) => request(`/memory/${id}/health`),
  listTypes: () => request('/memory/types'),
}
