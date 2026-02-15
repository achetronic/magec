import { request } from './client.js'

export const conversationsApi = {
  list: (params = {}) => {
    const query = new URLSearchParams()
    if (params.agentId) query.set('agentId', params.agentId)
    if (params.source) query.set('source', params.source)
    if (params.clientId) query.set('clientId', params.clientId)
    if (params.perspective) query.set('perspective', params.perspective)
    if (params.limit != null) query.set('limit', params.limit)
    if (params.offset != null) query.set('offset', params.offset)
    const qs = query.toString()
    return request(`/conversations${qs ? '?' + qs : ''}`)
  },
  get: (id, params = {}) => {
    const query = new URLSearchParams()
    if (params.msgLimit != null) query.set('msgLimit', params.msgLimit)
    if (params.msgOffset != null) query.set('msgOffset', params.msgOffset)
    const qs = query.toString()
    return request(`/conversations/${id}${qs ? '?' + qs : ''}`)
  },
  delete: (id) => request(`/conversations/${id}`, { method: 'DELETE' }),
  clear: () => request('/conversations/clear', { method: 'DELETE' }),
  stats: () => request('/conversations/stats'),
  updateSummary: (id, summary) =>
    request(`/conversations/${id}/summary`, {
      method: 'PUT',
      body: JSON.stringify({ summary }),
    }),
  resetSession: (id) =>
    request(`/conversations/${id}/reset-session`, { method: 'POST' }),
  findPair: (id) => request(`/conversations/${id}/pair`),
}
