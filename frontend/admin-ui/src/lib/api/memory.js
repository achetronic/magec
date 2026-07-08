// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'

export const memoryApi = {
  list: () => request('/memory'),
  get: (id) => request(`/memory/${id}`),
  create: (m) => request('/memory', { method: 'POST', body: JSON.stringify(m) }),
  update: (id, m) => request(`/memory/${id}`, { method: 'PUT', body: JSON.stringify(m) }),
  delete: (id, force = false) => request(`/memory/${id}${force ? '?force=true' : ''}`, { method: 'DELETE' }),
  checkHealth: (id) => request(`/memory/${id}/health`),
  listTypes: () => request('/memory/types'),
}
