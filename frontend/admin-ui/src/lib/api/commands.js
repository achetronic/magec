// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'

export const commandsApi = {
  list: () => request('/commands'),
  get: (id) => request(`/commands/${id}`),
  create: (c) => request('/commands', { method: 'POST', body: JSON.stringify(c) }),
  update: (id, c) => request(`/commands/${id}`, { method: 'PUT', body: JSON.stringify(c) }),
  delete: (id) => request(`/commands/${id}`, { method: 'DELETE' }),
}
