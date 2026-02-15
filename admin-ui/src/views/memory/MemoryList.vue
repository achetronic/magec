<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <h2 class="text-sm font-semibold text-arena-200">Memory Providers</h2>
      </div>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors cursor-pointer">
        + New Provider
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.memory.length" />
    <EmptyState v-else-if="!store.memory.length" title="No memory providers configured" subtitle="Add a Redis for sessions or Postgres for long-term memory" icon="database" color="green" actionLabel="+ New Provider" @action="openDialog()" />
    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <MemoryCard
        v-for="m in store.memory" :key="m.id"
        :provider="m"
        :active="isActive(m)"
        @edit="openDialog(m)"
        @delete="handleDelete(m)"
        @activate="toggleActive(m)"
      />
    </div>

    <MemoryDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import MemoryCard from './MemoryCard.vue'
import MemoryDialog from './MemoryDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function settingsKey(m) {
  return m.category === 'session' ? 'sessionProvider' : 'longTermProvider'
}

function isActive(m) {
  return store.settings[settingsKey(m)] === m.id
}

async function toggleActive(m) {
  const key = settingsKey(m)
  const newId = isActive(m) ? '' : m.id
  try {
    await store.saveSettings({ ...store.settings, [key]: newId })
  } catch (e) {
    toast.error(e.message)
  }
}

function openDialog(mem = null) {
  dialog.value?.open(mem)
}

function handleDelete(m) {
  requestDelete(`Delete memory provider "${m.name}"? This cannot be undone.`, async () => {
    try {
      await memoryApi.delete(m.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
