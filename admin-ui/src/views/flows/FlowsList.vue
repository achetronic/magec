<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Flows</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Flow
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.flows.length" />

    <EmptyState v-else-if="!store.flows.length" title="No flows configured" subtitle="Create multi-agent workflows with sequential, parallel, and loop steps" icon="flow" color="rose" actionLabel="+ New Flow" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="f in store.flows" :key="f.id" color="rose">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 bg-rose-500/15">
              <Icon name="flow" size="md" class="text-rose-400" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ f.name }}</h3>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(f)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(f)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="f.description" class="text-[10px] text-arena-400 mt-1">{{ f.description }}</p>
      </Card>
    </div>

    <FlowDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { flowsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import FlowDialog from './FlowDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function openDialog(flow = null) {
  dialog.value?.open(flow)
}

function handleDelete(f) {
  requestDelete(`Delete flow "${f.name}"? This cannot be undone.`, async () => {
    try {
      await flowsApi.delete(f.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
