import { request } from './client.js'

export const clientsApi = {
  list: () => request('/clients'),
  get: (id) => request(`/clients/${id}`),
  create: (c) => request('/clients', { method: 'POST', body: JSON.stringify(c) }),
  update: (id, c) => request(`/clients/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
  delete: (id) => request(`/clients/${id}`, { method: 'DELETE' }),
  regenerateToken: (id) => request(`/clients/${id}/regenerate-token`, { method: 'POST' }),
  telegramTest: (id, payload = null) => request(`/clients/${id}/telegram-test`, {
    method: 'POST',
    ...(payload ? { body: JSON.stringify(payload) } : {}),
  }),
  listTypes: () => request('/clients/types'),
}
