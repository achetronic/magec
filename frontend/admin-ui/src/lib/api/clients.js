// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'

export const clientsApi = {
  list: () => request('/clients'),
  get: (id) => request(`/clients/${id}`),
  create: (c) => request('/clients', { method: 'POST', body: JSON.stringify(c) }),
  update: (id, c) => request(`/clients/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
  delete: (id, force = false) => request(`/clients/${id}${force ? '?force=true' : ''}`, { method: 'DELETE' }),
  regenerateToken: (id) => request(`/clients/${id}/regenerate-token`, { method: 'POST' }),
  listTypes: () => request('/clients/types'),
}
