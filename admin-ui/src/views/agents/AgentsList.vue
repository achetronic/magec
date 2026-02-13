<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Agents</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Agent
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.agents.length" :grid="false" />

    <EmptyState v-else-if="!store.agents.length" title="No agents configured" subtitle="Create your first agent to get started" icon="users" color="sol" actionLabel="+ New Agent" @action="openDialog()" />

    <div v-else class="space-y-3">
      <Card v-for="a in store.agents" :key="a.id" :active="expandedId === a.id">
        <div class="flex items-center gap-3 cursor-pointer" @click="toggle(a.id)">
          <div class="w-9 h-9 rounded-lg bg-sol-500/15 flex items-center justify-center flex-shrink-0">
            <span class="text-sm font-semibold text-sol-400">{{ (a.name || a.id).charAt(0).toUpperCase() }}</span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-arena-100 text-sm">{{ a.name || a.id }}</h3>
              <span class="text-[10px] text-arena-500 font-mono">{{ a.id }}</span>
            </div>
            <div class="flex items-center gap-1.5 mt-1 flex-wrap">
              <Badge>{{ store.backendLabel(a.llm?.backend) }} / {{ a.llm?.model || '?' }}</Badge>
              <Badge v-if="a.transcription?.backend" variant="muted">STT</Badge>
              <Badge v-if="a.tts?.backend" variant="lava">TTS</Badge>
              <Badge v-if="(a.mcpServers||[]).length" variant="atlantico">{{ (a.mcpServers||[]).length }} MCP{{ (a.mcpServers||[]).length > 1 ? 's' : '' }}</Badge>
            </div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <button @click.stop="openDialog(a)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="md" class="text-arena-400" />
            </button>
            <button @click.stop="handleDelete(a)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="md" class="text-arena-400 hover:text-lava-400" />
            </button>
            <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform" :class="{ 'rotate-180': expandedId === a.id }" />
          </div>
        </div>
        <AgentDetail v-if="expandedId === a.id" :agent="a" />
      </Card>
    </div>

    <AgentDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { agentsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import AgentDetail from './AgentDetail.vue'
import AgentDialog from './AgentDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const expandedId = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function toggle(id) {
  expandedId.value = expandedId.value === id ? null : id
}

function openDialog(agent = null) {
  dialog.value?.open(agent)
}

function handleDelete(a) {
  requestDelete(`Delete agent "${a.name || a.id}"? This cannot be undone.`, async () => {
    try {
      await agentsApi.delete(a.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
