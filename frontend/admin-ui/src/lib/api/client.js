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
  if (!res.ok) {
    // Attach status and payload so callers can react to structured errors
    // (e.g. the 409 reference breakdown of referential-integrity deletes).
    const err = new Error(data.error || `HTTP ${res.status}`)
    err.status = res.status
    err.data = data
    throw err
  }
  return data
}
