import { request } from './client.js'

export const flowsApi = {
  list: () => request('/flows'),
  get: (id) => request(`/flows/${id}`),
  create: (f) => request('/flows', { method: 'POST', body: JSON.stringify(f) }),
  update: (id, f) => request(`/flows/${id}`, { method: 'PUT', body: JSON.stringify(f) }),
  delete: (id, force = false) => request(`/flows/${id}${force ? '?force=true' : ''}`, { method: 'DELETE' }),
}
