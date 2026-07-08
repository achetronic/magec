// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  backendsApi,
  agentsApi,
  memoryApi,
  mcpsApi,
  clientsApi,
  flowsApi,
  commandsApi,
  skillsApi,
  settingsApi,
  secretsApi,
  voiceApi,
} from '../api/index.js'

export const useDataStore = defineStore('data', () => {
  const backends = ref([])
  const agents = ref([])
  const memory = ref([])
  const mcps = ref([])
  const clients = ref([])
  const flows = ref([])
  const commands = ref([])
  const skills = ref([])
  const secrets = ref([])
  const memoryTypes = ref([])
  const clientTypes = ref([])
  const voiceTypes = ref([])
  const settings = ref({ sessionProvider: '', longTermProvider: '' })
  const loading = ref(false)

  async function init() {
    try { memoryTypes.value = await memoryApi.listTypes() } catch { memoryTypes.value = [] }
    try { clientTypes.value = await clientsApi.listTypes() } catch { clientTypes.value = [] }
    try { voiceTypes.value = await voiceApi.listTypes() } catch { voiceTypes.value = [] }
    await refresh()
  }

  async function refresh() {
    loading.value = true
    try {
      const results = await Promise.all([
        backendsApi.list(),
        agentsApi.list(),
        memoryApi.list(),
        mcpsApi.list(),
        clientsApi.list(),
        flowsApi.list(),
        commandsApi.list(),
        skillsApi.list(),
        settingsApi.get(),
        secretsApi.list(),
      ])
      backends.value = results[0] || []
      agents.value = results[1] || []
      memory.value = results[2] || []
      mcps.value = results[3] || []
      clients.value = results[4] || []
      flows.value = results[5] || []
      commands.value = results[6] || []
      skills.value = results[7] || []
      settings.value = results[8] || { sessionProvider: '', longTermProvider: '' }
      secrets.value = results[9] || []
    } catch (e) {
      console.error('Failed to load data:', e)
    } finally {
      loading.value = false
    }
  }

  function backendLabel(id) {
    if (!id) return ''
    const b = backends.value.find((b) => b.id === id)
    return b?.name || id
  }

  function agentLabel(id) {
    if (!id) return ''
    const a = agents.value.find((a) => a.id === id)
    return a?.name || a?.id || id
  }

  function commandLabel(id) {
    if (!id) return ''
    const c = commands.value.find((c) => c.id === id)
    return c?.name || id
  }

  function skillLabel(id) {
    if (!id) return ''
    const sk = skills.value.find((sk) => sk.id === id)
    return sk?.name || id
  }

  async function saveSettings(newSettings) {
    settings.value = await settingsApi.update(newSettings)
  }

  return {
    backends,
    agents,
    memory,
    mcps,
    clients,
    flows,
    commands,
    skills,
    secrets,
    memoryTypes,
    clientTypes,
    voiceTypes,
    settings,
    loading,
    init,
    refresh,
    saveSettings,
    backendLabel,
    agentLabel,
    commandLabel,
    skillLabel,
  }
})
