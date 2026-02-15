<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">MCP Servers</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New MCP
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.mcps.length" />

    <EmptyState v-else-if="!store.mcps.length" title="No MCP servers configured" subtitle="Connect external tools via Model Context Protocol" icon="bolt" color="atlantico" actionLabel="+ New MCP" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="m in store.mcps" :key="m.id" color="atlantico">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg bg-atlantico-500/15 flex items-center justify-center flex-shrink-0">
              <Icon name="bolt" size="md" class="text-atlantico-400" />
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-1.5">
                <h3 class="font-medium text-arena-100 text-sm">{{ m.name }}</h3>
                <Badge variant="muted">{{ m.type || 'http' }}</Badge>
              </div>
              <p class="text-[10px] text-arena-500 truncate">{{ m.endpoint || m.command || '' }}</p>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(m)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(m)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="m.systemPrompt" class="text-[10px] text-arena-400 mb-2 line-clamp-2">{{ m.systemPrompt }}</p>
        <div v-if="usedBy(m.id).length" class="flex flex-wrap gap-1">
          <Tooltip v-for="ref in usedBy(m.id)" :key="ref.name" :text="ref.tooltip">
            <Badge variant="muted">{{ ref.name }}</Badge>
          </Tooltip>
        </div>
        <p v-else class="text-[10px] text-arena-600">Not linked to any agent</p>
      </Card>
    </div>

    <McpDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { mcpsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Tooltip from '../../components/Tooltip.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import McpDialog from './McpDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function openDialog(mcp = null) {
  dialog.value?.open(mcp)
}

function usedBy(id) {
  const refs = []
  for (const a of store.agents) {
    if ((a.mcpServers || []).includes(id)) {
      const name = a.name || a.id
      const prompt = a.systemPrompt ? a.systemPrompt.slice(0, 80) + (a.systemPrompt.length > 80 ? '...' : '') : ''
      refs.push({ name, tooltip: a.description || prompt })
    }
  }
  return refs
}

function handleDelete(m) {
  requestDelete(`Delete MCP server "${m.name}"? This cannot be undone.`, async () => {
    try {
      await mcpsApi.delete(m.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
