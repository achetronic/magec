import { request } from './client.js'

export const cronsApi = {
  list: () => request('/crons'),
  get: (id) => request(`/crons/${id}`),
  create: (c) => request('/crons', { method: 'POST', body: JSON.stringify(c) }),
  update: (id, c) => request(`/crons/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
  delete: (id) => request(`/crons/${id}`, { method: 'DELETE' }),
}
