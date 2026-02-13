import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  backendsApi,
  agentsApi,
  memoryApi,
  mcpsApi,
  cronsApi,
  clientsApi,
  flowsApi,
  commandsApi,
  triggersApi,
} from '../api/index.js'

export const useDataStore = defineStore('data', () => {
  const backends = ref([])
  const agents = ref([])
  const memory = ref([])
  const mcps = ref([])
  const crons = ref([])
  const clients = ref([])
  const flows = ref([])
  const commands = ref([])
  const triggers = ref([])
  const memoryTypes = ref([])
  const clientTypes = ref([])
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
        cronsApi.list(),
        clientsApi.list(),
        flowsApi.list(),
        commandsApi.list(),
        triggersApi.list(),
      ])
      backends.value = results[0] || []
      agents.value = results[1] || []
      memory.value = results[2] || []
      mcps.value = results[3] || []
      crons.value = results[4] || []
      clients.value = results[5] || []
      flows.value = results[6] || []
      commands.value = results[7] || []
      triggers.value = results[8] || []
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

  return {
    backends,
    agents,
    memory,
    mcps,
    crons,
    clients,
    flows,
    commands,
    triggers,
    memoryTypes,
    clientTypes,
    loading,
    init,
    refresh,
    backendLabel,
    memoryLabel,
    agentLabel,
    commandLabel,
  }
})
