import { request } from './client.js'

export const settingsApi = {
  get: () => request('/settings'),
  update: (s) => request('/settings', { method: 'PUT', body: JSON.stringify(s) }),
}
