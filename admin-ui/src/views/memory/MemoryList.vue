<template>
  <div class="space-y-6">
    <!-- Session Memory -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h2 class="text-sm font-semibold text-arena-200">Session Memory</h2>
          <Badge variant="green" v-for="t in sessionTypes" :key="t.type">{{ t.displayName }}</Badge>
        </div>
        <button @click="openDialog(null, 'session')" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
          + New Provider
        </button>
      </div>
      <p class="text-[11px] text-arena-500 -mt-1">Short-lived conversation state with TTL-based expiration.</p>

      <SkeletonCard v-if="store.loading && !sessionProviders.length" />
      <EmptyState v-else-if="!sessionProviders.length" title="No session providers configured" subtitle="Short-lived conversation state with TTL-based expiration" icon="database" color="green" actionLabel="+ New Provider" @action="openDialog(null, 'session')" />
      <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
        <MemoryCard v-for="m in sessionProviders" :key="m.id" :provider="m" @edit="openDialog(m)" @delete="handleDelete(m)" />
      </div>
    </div>

    <!-- Long-Term Memory -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <h2 class="text-sm font-semibold text-arena-200">Long-Term Memory</h2>
          <Badge variant="green" v-for="t in longTermTypes" :key="t.type">{{ t.displayName }}</Badge>
        </div>
        <button @click="openDialog(null, 'longterm')" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
          + New Provider
        </button>
      </div>
      <p class="text-[11px] text-arena-500 -mt-1">Persistent facts and preferences with vector embeddings for semantic recall.</p>

      <SkeletonCard v-if="store.loading && !longTermProviders.length" />
      <EmptyState v-else-if="!longTermProviders.length" title="No long-term providers configured" subtitle="Persistent facts and preferences with vector embeddings" icon="database" color="green" actionLabel="+ New Provider" @action="openDialog(null, 'longterm')" />
      <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
        <MemoryCard v-for="m in longTermProviders" :key="m.id" :provider="m" @edit="openDialog(m)" @delete="handleDelete(m)" />
      </div>
    </div>

    <MemoryDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { computed, inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
import Badge from '../../components/Badge.vue'
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

const sessionTypes = computed(() => store.memoryTypes.filter(t => t.categories?.includes('session')))
const longTermTypes = computed(() => store.memoryTypes.filter(t => t.categories?.includes('longterm')))
const sessionProviders = computed(() => store.memory.filter(m => m.category === 'session'))
const longTermProviders = computed(() => store.memory.filter(m => m.category === 'longterm'))

function openDialog(mem = null, category = null) {
  dialog.value?.open(mem, category)
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
