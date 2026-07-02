import { request } from './client.js'

export const runsApi = {
  list: (params = {}) => {
    const query = new URLSearchParams()
    if (params.appName) query.set('appName', params.appName)
    if (params.status) query.set('status', params.status)
    if (params.limit != null) query.set('limit', params.limit)
    if (params.offset != null) query.set('offset', params.offset)
    const qs = query.toString()
    return request(`/runs${qs ? '?' + qs : ''}`)
  },
  get: (id, params = {}) => {
    const query = new URLSearchParams()
    if (params.raw != null) query.set('raw', String(params.raw))
    const qs = query.toString()
    return request(`/runs/${id}${qs ? '?' + qs : ''}`)
  }
}
