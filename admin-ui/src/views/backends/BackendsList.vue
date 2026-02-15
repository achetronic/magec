<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Backends</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Backend
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.backends.length" />

    <EmptyState v-else-if="!store.backends.length" title="No backends configured" subtitle="Create a backend to connect AI providers" icon="server" color="purple" actionLabel="+ New Backend" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="b in store.backends" :key="b.id" color="purple">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg bg-purple-500/15 flex items-center justify-center flex-shrink-0">
              <span class="text-[10px] font-mono font-bold text-purple-300">{{ (b.type || '').substring(0, 3).toUpperCase() }}</span>
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ b.name }}</h3>
              <p class="text-[10px] text-arena-500 truncate">{{ b.url || b.type }}</p>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(b)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(b)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <div v-if="usedBy(b.id).length" class="flex flex-wrap gap-1">
          <Tooltip v-for="ref in usedBy(b.id)" :key="ref.name" :text="ref.tooltip">
            <Badge variant="muted">{{ ref.name }}</Badge>
          </Tooltip>
        </div>
        <p v-else class="text-[10px] text-arena-600">Not used by any agent</p>
      </Card>
    </div>

    <BackendDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { backendsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Tooltip from '../../components/Tooltip.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import BackendDialog from './BackendDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function openDialog(backend = null) {
  dialog.value?.open(backend)
}

function usedBy(id) {
  const refs = new Map()
  for (const a of store.agents) {
    const label = a.name || a.id
    if (refs.has(label)) continue
    const roles = []
    if (a.llm?.backend === id) roles.push('LLM')
    if (a.transcription?.backend === id) roles.push('STT')
    if (a.tts?.backend === id) roles.push('TTS')
    if (roles.length) {
      const prompt = a.systemPrompt ? a.systemPrompt.slice(0, 80) + (a.systemPrompt.length > 80 ? '...' : '') : ''
      refs.set(label, { name: label, tooltip: roles.join(' + ') + (prompt ? ' — ' + prompt : '') })
    }
  }
  return [...refs.values()]
}

function handleDelete(b) {
  requestDelete(`Delete backend "${b.name}"? This cannot be undone.`, async () => {
    try {
      await backendsApi.delete(b.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
