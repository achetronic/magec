// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

export const CONFIG = {
  transcription: {
    model: 'parakeet'
  },
  agent: {
    baseUrl: '/api/v1/agent',
    appName: 'default',
    defaultUserId: 'default_user'
  },
  session: {
    storageKey: 'magec_sessions',
    autoRotateMinutes: 30,
    maxStoredSessions: 50
  }
}
