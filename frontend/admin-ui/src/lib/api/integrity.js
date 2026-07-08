// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { request } from './client.js'

export const integrityApi = {
  deadReferences: () => request('/integrity/dead-references'),
  cleanDeadReferences: () => request('/integrity/dead-references/clean', { method: 'POST' }),
}
