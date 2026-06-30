// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'
import { getAuthHeaders } from '../auth.js'

const BASE = '/api/v1/admin'

// skillsApi covers the new on-disk skills layout (decision #29). The
// store keeps {id, slug}; everything else (name, description,
// instructions, resources) is read live from the SKILL.md package the
// operator uploaded. Mutations are upload-only — there is no manual
// create/edit form, you replace the package and it shows up.
export const skillsApi = {
  list: () => request('/skills'),
  get: (id) => request(`/skills/${id}`),
  delete: (id) => request(`/skills/${id}`, { method: 'DELETE' }),

  // upload accepts a SKILL.md or a .zip/.tar.gz package. Pass
  // replace=true to overwrite an existing skill that owns the same
  // slug (frontmatter name); the existing store ID is preserved so
  // agent links keep working.
  upload: async (file, { replace = false } = {}) => {
    const form = new FormData()
    form.append('file', file)
    const url = `${BASE}/skills/upload${replace ? '?replace=true' : ''}`
    const res = await fetch(url, {
      method: 'POST',
      headers: { ...getAuthHeaders() },
      body: form,
    })
    const data = await res.json()
    if (!res.ok) {
      const err = new Error(data.error || `HTTP ${res.status}`)
      err.status = res.status
      err.payload = data
      throw err
    }
    return data
  },

  // downloadUrl returns the path of the tar.gz export endpoint. The
  // caller fetches it with the admin Bearer header and triggers a
  // browser download — the URL itself is NOT signed; access is
  // gated by the standard admin auth middleware.
  downloadUrl: (id) => `${BASE}/skills/${id}/download`,
}
