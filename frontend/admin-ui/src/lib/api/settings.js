// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'

export const settingsApi = {
  get: () => request('/settings'),
  update: (s) => request('/settings', { method: 'PUT', body: JSON.stringify(s) }),
}
