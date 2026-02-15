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
  settingsApi,
} from '../api/index.js'

export const useDataStore = defineStore('data', () => {
  const backends = ref([])
  const agents = ref([])
  const memory = ref([])
  const mcps = ref([])
  const clients = ref([])
  const flows = ref([])
  const commands = ref([])
  const memoryTypes = ref([])
  const clientTypes = ref([])
  const settings = ref({ sessionProvider: '', longTermProvider: '' })
  const loading = ref(false)

  async function init() {
    try { memoryTypes.value = await memoryApi.listTypes() } catch { memoryTypes.value = [] }
    try { clientTypes.value = await clientsApi.listTypes() } catch { clientTypes.value = [] }
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
        settingsApi.get(),
      ])
      backends.value = results[0] || []
      agents.value = results[1] || []
      memory.value = results[2] || []
      mcps.value = results[3] || []
      clients.value = results[4] || []
      flows.value = results[5] || []
      commands.value = results[6] || []
      settings.value = results[7] || { sessionProvider: '', longTermProvider: '' }
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

  function memoryLabel(id) {
    if (!id) return ''
    const m = memory.value.find((m) => m.id === id)
    return m?.name || id
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
    memoryTypes,
    clientTypes,
    settings,
    loading,
    init,
    refresh,
    saveSettings,
    backendLabel,
    memoryLabel,
    agentLabel,
    commandLabel,
  }
})
