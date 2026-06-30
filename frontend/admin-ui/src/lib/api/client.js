// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { getAuthHeaders } from '../auth.js'

const BASE = '/api/v1/admin'

export async function request(path, opts = {}) {
  const authHeaders = getAuthHeaders()
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...authHeaders, ...opts.headers },
    ...opts,
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
  return data
}
