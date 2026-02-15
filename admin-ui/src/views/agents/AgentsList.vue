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

    <template v-else>
      <!-- Search + Tag Filter Bar -->
      <div class="space-y-3">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search agents..."
          class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm text-arena-200 placeholder-arena-600 focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
        />

        <div v-if="allTags.length" class="space-y-2">
          <button
            type="button"
            @click="showTags = !showTags"
            class="flex items-center gap-1.5 text-[11px] text-arena-500 hover:text-arena-300 transition-colors cursor-pointer"
          >
            <Icon name="chevronRight" size="sm" class="transition-transform" :class="{ 'rotate-90': showTags }" />
            <span>Tags</span>
            <span v-if="selectedTags.length" class="px-1.5 py-0.5 text-[9px] font-medium rounded bg-sol-500/15 text-sol-300">{{ selectedTags.length }}</span>
          </button>
          <div v-if="showTags" class="flex flex-wrap gap-1.5">
            <button
              v-for="tag in allTags" :key="tag"
              type="button"
              @click="toggleTag(tag)"
              class="px-2.5 py-1 text-[11px] font-medium rounded-lg border transition-all cursor-pointer"
              :class="selectedTags.includes(tag)
                ? 'bg-sol-500/15 text-sol-300 border-sol-500/30'
                : 'bg-piedra-800 text-arena-500 border-piedra-700/40 hover:border-piedra-600 hover:text-arena-300'"
            >
              {{ tag }}
            </button>
            <button
              v-if="selectedTags.length"
              type="button"
              @click="selectedTags = []"
              class="px-2.5 py-1 text-[11px] font-medium rounded-lg text-arena-500 hover:text-arena-300 transition-colors cursor-pointer"
            >
              Clear
            </button>
          </div>
        </div>
      </div>

      <!-- Agent count + pagination info -->
      <div v-if="filteredAgents.length !== store.agents.length || totalPages > 1" class="flex items-center justify-between">
        <span class="text-[11px] text-arena-500">
          {{ filteredAgents.length }} of {{ store.agents.length }} agents
        </span>
        <span v-if="totalPages > 1" class="text-[11px] text-arena-500">
          Page {{ currentPage }} / {{ totalPages }}
        </span>
      </div>

      <!-- Cards -->
      <div class="space-y-3">
        <Card v-for="a in paginatedAgents" :key="a.id" color="sol">
          <div class="flex items-center gap-3 cursor-pointer" @click="toggle(a.id)">
            <div class="w-9 h-9 rounded-lg bg-sol-500/15 flex items-center justify-center flex-shrink-0">
              <span class="text-sm font-semibold text-sol-400">{{ (a.name || a.id).charAt(0).toUpperCase() }}</span>
            </div>
            <div class="min-w-0 flex-1">
              <h3 class="font-medium text-arena-100 text-sm">{{ a.name || a.id }}</h3>
              <div class="flex items-center gap-1.5 mt-2 flex-wrap">
                <Badge variant="muted">{{ store.backendLabel(a.llm?.backend) }} / {{ a.llm?.model || '?' }}</Badge>
                <Badge v-if="a.transcription?.backend" variant="muted">STT</Badge>
                <Badge v-if="a.tts?.backend" variant="muted">TTS</Badge>
                <Badge v-if="(a.mcpServers||[]).length" variant="muted">{{ (a.mcpServers||[]).length }} MCP{{ (a.mcpServers||[]).length > 1 ? 's' : '' }}</Badge>
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

      <!-- Pagination Controls -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 pt-2">
        <button
          @click="currentPage = Math.max(1, currentPage - 1)"
          :disabled="currentPage <= 1"
          class="px-3 py-1.5 text-xs font-medium rounded-lg border transition-all cursor-pointer"
          :class="currentPage <= 1
            ? 'border-piedra-700/20 text-arena-600 cursor-not-allowed'
            : 'border-piedra-700/40 text-arena-300 hover:border-piedra-600 hover:bg-piedra-800'"
        >
          &larr; Prev
        </button>
        <button
          v-for="p in pageNumbers" :key="p"
          @click="currentPage = p"
          class="w-8 h-8 text-xs font-medium rounded-lg border transition-all cursor-pointer"
          :class="p === currentPage
            ? 'bg-sol-500/15 text-sol-300 border-sol-500/30'
            : 'border-piedra-700/40 text-arena-400 hover:border-piedra-600 hover:text-arena-300'"
        >
          {{ p }}
        </button>
        <button
          @click="currentPage = Math.min(totalPages, currentPage + 1)"
          :disabled="currentPage >= totalPages"
          class="px-3 py-1.5 text-xs font-medium rounded-lg border transition-all cursor-pointer"
          :class="currentPage >= totalPages
            ? 'border-piedra-700/20 text-arena-600 cursor-not-allowed'
            : 'border-piedra-700/40 text-arena-300 hover:border-piedra-600 hover:bg-piedra-800'"
        >
          Next &rarr;
        </button>
      </div>
    </template>

    <AgentDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
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
const expandedTags = reactive({})
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

const searchQuery = ref('')
const selectedTags = ref([])
const showTags = ref(false)
const currentPage = ref(1)
const pageSize = 10

const allTags = computed(() => {
  const tagSet = new Set()
  for (const a of store.agents) {
    for (const t of a.tags || []) tagSet.add(t)
  }
  return [...tagSet].sort()
})

const filteredAgents = computed(() => {
  let list = store.agents
  if (selectedTags.value.length) {
    list = list.filter(a =>
      (a.tags || []).some(t => selectedTags.value.includes(t))
    )
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter(a =>
      (a.name || '').toLowerCase().includes(q) ||
      (a.description || '').toLowerCase().includes(q) ||
      a.id.toLowerCase().includes(q)
    )
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredAgents.value.length / pageSize)))

const pageNumbers = computed(() => {
  const pages = []
  for (let i = 1; i <= totalPages.value; i++) pages.push(i)
  return pages
})

const paginatedAgents = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredAgents.value.slice(start, start + pageSize)
})

watch([searchQuery, selectedTags], () => { currentPage.value = 1 })

function toggleTag(tag) {
  const idx = selectedTags.value.indexOf(tag)
  if (idx === -1) selectedTags.value.push(tag)
  else selectedTags.value.splice(idx, 1)
}

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
