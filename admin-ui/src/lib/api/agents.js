import { request } from './client.js'

export const agentsApi = {
  list: () => request('/agents'),
  get: (id) => request(`/agents/${id}`),
  create: (a) => request('/agents', { method: 'POST', body: JSON.stringify(a) }),
  update: (id, a) => request(`/agents/${id}`, { method: 'PUT', body: JSON.stringify(a) }),
  delete: (id) => request(`/agents/${id}`, { method: 'DELETE' }),
  listMCPs: (id) => request(`/agents/${id}/mcps`),
  linkMCP: (id, mcpId) => request(`/agents/${id}/mcps/${mcpId}`, { method: 'PUT' }),
  unlinkMCP: (id, mcpId) => request(`/agents/${id}/mcps/${mcpId}`, { method: 'DELETE' }),
}
