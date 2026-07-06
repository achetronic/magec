import { request } from './client.js'

export const secretsApi = {
  list: () => request('/secrets'),
  get: (id) => request(`/secrets/${id}`),
  create: (s) => request('/secrets', { method: 'POST', body: JSON.stringify(s) }),
  update: (id, s) => request(`/secrets/${id}`, { method: 'PUT', body: JSON.stringify(s) }),
  delete: (id, force = false) => request(`/secrets/${id}${force ? '?force=true' : ''}`, { method: 'DELETE' }),
}
