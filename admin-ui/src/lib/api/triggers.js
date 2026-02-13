import { request } from './client.js'

export const triggersApi = {
  list: () => request('/triggers'),
  get: (id) => request(`/triggers/${id}`),
  create: (t) => request('/triggers', { method: 'POST', body: JSON.stringify(t) }),
  update: (id, t) => request(`/triggers/${id}`, { method: 'PUT', body: JSON.stringify(t) }),
  delete: (id) => request(`/triggers/${id}`, { method: 'DELETE' }),
}
