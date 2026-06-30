// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { getAuthHeaders } from '../auth.js'

const BASE = '/api/v1/admin'

export const backupApi = {
  async download() {
    const res = await fetch(`${BASE}/settings/backup`, { headers: getAuthHeaders() })
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new Error(data.error || `HTTP ${res.status}`)
    }
    const blob = await res.blob()
    const cd = res.headers.get('content-disposition') || ''
    const match = cd.match(/filename="?([^"]+)"?/)
    const filename = match ? match[1] : 'magec-backup.tar.gz'

    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = filename
    a.click()
    URL.revokeObjectURL(a.href)
  },

  async restore(file) {
    const res = await fetch(`${BASE}/settings/restore`, {
      method: 'POST',
      headers: { ...getAuthHeaders(), 'Content-Type': 'application/gzip' },
      body: file,
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
    return data
  },
}
