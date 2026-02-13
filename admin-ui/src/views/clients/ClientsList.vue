<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Clients</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Client
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.clients.length" />

    <EmptyState v-else-if="!store.clients.length" title="No clients configured" subtitle="Create a client to connect devices, Telegram bots, and more" icon="phone" color="lava" actionLabel="+ New Client" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="c in store.clients" :key="c.id">
        <div class="flex items-start justify-between gap-3 mb-3" :class="{ 'opacity-60': !c.enabled }">
          <div class="flex items-center gap-3 min-w-0">
            <div class="relative w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
              :class="c.enabled ? 'bg-lava-500/15' : 'bg-piedra-800'">
              <span class="text-[10px] font-mono font-bold" :class="c.enabled ? 'text-lava-400' : 'text-arena-500'">
                {{ typeLabel(c.type).slice(0, 3).toUpperCase() }}
              </span>
              <span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900"
                :class="c.enabled ? 'bg-green-500' : 'bg-lava-500'"
                :title="c.enabled ? 'Enabled' : 'Disabled'" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ c.name }}</h3>
              <p class="text-[10px] text-arena-500">{{ typeLabel(c.type) }}</p>
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
        <div v-if="agentRefs(c).length" class="flex flex-wrap gap-1">
          <Tooltip v-for="ref in agentRefs(c)" :key="ref.name" :text="ref.tooltip">
            <Badge variant="sol">{{ ref.name }}</Badge>
          </Tooltip>
        </div>
        <p v-else class="text-[10px] text-arena-600">No agents assigned</p>
      </Card>
    </div>

    <ClientDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
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
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function typeLabel(type) {
  const t = store.clientTypes.find(t => t.type === type)
  return t?.displayName || type || 'device'
}

function agentRefs(c) {
  return (c.allowedAgents || []).map(id => {
    const a = store.agents.find(a => a.id === id)
    const name = a?.name || a?.id || id
    const prompt = a?.systemPrompt ? a.systemPrompt.slice(0, 80) + (a.systemPrompt.length > 80 ? '...' : '') : ''
    return { name, tooltip: a?.description || prompt }
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
