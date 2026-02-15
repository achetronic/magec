<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Clients</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Client
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.clients.length" />

    <EmptyState v-else-if="!store.clients.length" title="No clients configured" subtitle="Create a client to connect devices, bots, cron jobs, or webhooks" icon="phone" color="lava" actionLabel="+ New Client" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="c in store.clients" :key="c.id" color="lava">
        <div class="flex items-start justify-between gap-3 mb-2" :class="{ 'opacity-60': !c.enabled }">
          <div class="flex items-center gap-3 min-w-0">
            <div class="relative w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
              :class="c.enabled ? 'bg-lava-500/15' : 'bg-piedra-800'">
              <span class="text-[10px] font-mono font-bold" :class="c.enabled ? 'text-lava-400' : 'text-arena-500'">
                {{ clientTypeAbbrev(c.type) }}
              </span>
              <span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900"
                :class="c.enabled ? 'bg-green-500' : 'bg-lava-500'"
                :title="c.enabled ? 'Enabled' : 'Disabled'" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ c.name }}</h3>
              <div class="flex items-center gap-1.5 mt-0.5">
                <Badge :variant="clientTypeBadge(c.type)">{{ clientTypeLabel(c.type) }}</Badge>
                <span v-if="c.type === 'cron' && c.config?.cron?.schedule" class="text-[10px] text-arena-500 font-mono">{{ c.config.cron.schedule }}</span>
                <Badge v-if="c.type === 'webhook' && c.config?.webhook?.passthrough" variant="muted">passthrough</Badge>
              </div>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(c)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(c)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <div v-if="agentRefs(c).length || commandRef(c)" class="border-t border-piedra-700/30 pt-2 mt-2 flex flex-wrap gap-1.5">
          <Tooltip v-if="commandRef(c)" :text="commandRef(c).tooltip">
            <Badge variant="muted">{{ commandRef(c).name }}</Badge>
          </Tooltip>
          <template v-for="(ref, i) in agentRefs(c)" :key="ref.name">
            <Tooltip v-if="i < 2 || expandedChips[c.id]" :text="ref.tooltip">
              <Badge variant="muted">{{ ref.isFlow ? '⤳ ' : '' }}{{ ref.name }}</Badge>
            </Tooltip>
          </template>
          <button
            v-if="agentRefs(c).length > 2"
            type="button"
            @click="expandedChips[c.id] = !expandedChips[c.id]"
            class="px-2 py-0.5 text-[10px] font-medium rounded bg-piedra-800 text-arena-500 hover:text-arena-300 border border-piedra-700/40 hover:border-piedra-600 transition-all cursor-pointer"
          >
            {{ expandedChips[c.id] ? 'less' : `+${agentRefs(c).length - 2} more` }}
          </button>
        </div>
        <p v-if="!agentRefs(c).length && !commandRef(c)" class="text-[10px] text-arena-600">No agents or flows assigned</p>
      </Card>
    </div>

    <ClientDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, reactive, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { clientsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Tooltip from '../../components/Tooltip.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import ClientDialog from './ClientDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const expandedChips = reactive({})
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function clientTypeLabel(t) {
  const info = store.clientTypes.find(ct => ct.type === t)
  return info?.displayName || t || 'Direct'
}

function clientTypeAbbrev(t) {
  return clientTypeLabel(t).slice(0, 3).toUpperCase()
}

function clientTypeBadge() {
  return 'muted'
}

function commandRef(c) {
  let cmdId = c.config?.cron?.commandId || c.config?.webhook?.commandId
  if (!cmdId) return null
  const cmd = store.commands.find(cmd => cmd.id === cmdId)
  if (!cmd) return null
  const tooltip = cmd.prompt ? cmd.prompt.slice(0, 100) + (cmd.prompt.length > 100 ? '...' : '') : ''
  return { name: cmd.name, tooltip }
}

function agentRefs(c) {
  return (c.allowedAgents || []).map(id => {
    const a = store.agents.find(a => a.id === id)
    if (a) {
      const prompt = a.systemPrompt ? a.systemPrompt.slice(0, 80) + (a.systemPrompt.length > 80 ? '...' : '') : ''
      return { name: a.name || a.id || id, tooltip: a.description || prompt, isFlow: false }
    }
    const f = store.flows.find(f => f.id === id)
    if (f) {
      return { name: f.name || f.id || id, tooltip: f.description || '', isFlow: true }
    }
    return { name: id, tooltip: '', isFlow: false }
  })
}

function openDialog(client = null) {
  dialog.value?.open(client)
}

function handleDelete(c) {
  requestDelete(`Delete client "${c.name}"? This cannot be undone.`, async () => {
    try {
      await clientsApi.delete(c.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
