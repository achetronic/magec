import { request } from './client.js'

export const mcpsApi = {
  list: () => request('/mcps'),
  get: (id) => request(`/mcps/${id}`),
  create: (m) => request('/mcps', { method: 'POST', body: JSON.stringify(m) }),
  update: (id, m) => request(`/mcps/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
  delete: (id) => request(`/mcps/${id}`, { method: 'DELETE' }),
}
