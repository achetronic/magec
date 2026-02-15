<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Conversations</h2>
      <div class="flex items-center gap-3">
        <!-- Auto-refresh segmented control -->
        <div class="flex items-center bg-piedra-800 rounded-lg border border-piedra-700/50 p-0.5">
          <button
            v-for="opt in refreshOptions" :key="opt.value"
            @click="setAutoRefresh(opt.value)"
            class="px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
            :class="refreshInterval === opt.value
              ? 'bg-piedra-700 text-arena-100'
              : 'text-arena-500 hover:text-arena-300'"
          >{{ opt.label }}</button>
        </div>
        <!-- Actions -->
        <div class="flex items-center gap-1">
          <button
            @click="silentRefresh"
            class="p-1.5 rounded-lg transition-colors group/btn"
            :class="refreshPulse ? 'bg-piedra-800' : 'hover:bg-piedra-800'"
            title="Refresh now"
          >
            <Icon name="refresh" size="sm"
              class="transition-all duration-300"
              :class="refreshPulse
                ? 'text-arena-200 rotate-180'
                : 'text-arena-500 group-hover/btn:text-arena-300'"
            />
          </button>
          <button
            v-if="conversations.length"
            @click="handleClearAll"
            class="flex items-center gap-1 p-1.5 hover:bg-piedra-800 rounded-lg transition-colors group/btn"
            title="Clear all conversations"
          >
            <Icon name="trash" size="sm" class="text-arena-500 group-hover/btn:text-arena-300 transition-colors" />
            <span class="text-[10px] font-medium text-arena-500 group-hover/btn:text-arena-300 transition-colors">All</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div v-if="grouped.length || hasFilters" class="flex flex-wrap items-center gap-2">
      <select
        v-model="filterAgent"
        class="bg-piedra-800 border border-piedra-700/50 text-arena-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-piedra-600"
      >
        <option value="">All agents</option>
        <option v-for="a in store.agents" :key="a.id" :value="a.id">{{ a.name }}</option>
        <option v-for="f in store.flows" :key="f.id" :value="f.id">{{ f.name }} (flow)</option>
      </select>
      <select
        v-model="filterSource"
        class="bg-piedra-800 border border-piedra-700/50 text-arena-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-piedra-600"
      >
        <option value="">All sources</option>
        <option value="voice-ui">Voice UI</option>
        <option value="telegram">Telegram</option>
        <option value="executor">Executor</option>
        <option value="flow">Flow</option>
        <option value="direct">Direct</option>
      </select>
      <button
        v-if="hasFilters"
        @click="filterAgent = ''; filterSource = ''"
        class="text-[10px] text-arena-500 hover:text-arena-300 px-2 py-1.5 transition-colors"
      >
        Clear filters
      </button>
      <div class="flex-1" />
      <span class="text-[10px] text-arena-500 tabular-nums">
        {{ grouped.length }} conversation{{ grouped.length !== 1 ? 's' : '' }}
      </span>
    </div>

    <!-- Loading -->
    <SkeletonCard v-if="loading && !conversations.length" />

    <!-- Empty state -->
    <EmptyState
      v-else-if="!grouped.length && !hasFilters"
      title="No conversations yet"
      subtitle="Conversations will appear here as agents interact with users"
      icon="chat"
      color="teal"
    />

    <EmptyState
      v-else-if="!grouped.length && hasFilters"
      title="No matching conversations"
      subtitle="Try adjusting your filters"
      icon="chat"
      color="teal"
    />

    <!-- Conversation list -->
    <div v-else class="space-y-2">
      <button
        v-for="c in grouped"
        :key="c.id"
        @click="$emit('select', c.id)"
        class="w-full text-left"
      >
        <Card color="teal" class="cursor-pointer group">
          <div class="flex items-start gap-3">
            <!-- Source icon -->
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0" :class="sourceBg(c.source)">
              <Icon :name="sourceIcon(c.source)" size="sm" :class="sourceText(c.source)" />
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0 space-y-2">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-arena-100 truncate">
                  {{ c.agentName || c.agentId }}
                </span>
                <Badge v-if="c.summary" variant="green">summarized</Badge>
              </div>
              <p v-if="c.preview" class="text-[11px] text-arena-500 italic truncate">"{{ c.preview }}"</p>
              <div class="flex items-center gap-1.5">
                <Badge variant="muted" class="!py-0">{{ formatSource(c.source) }}</Badge>
                <Badge v-if="c.flowId" variant="muted" class="!py-0">Flow</Badge>
                <Badge v-if="c.clientName" variant="muted" class="!py-0">{{ c.clientName }}</Badge>
              </div>
              <span class="text-[10px] text-arena-600 tabular-nums">{{ formatTime(c.startedAt) }}</span>
            </div>

            <!-- Arrow -->
            <Icon name="chevronRight" size="sm" class="text-arena-600 group-hover:text-arena-400 flex-shrink-0 mt-2 transition-colors" />
          </div>
        </Card>
      </button>

      <!-- Load more -->
      <button
        v-if="hasMore"
        @click="loadMore"
        :disabled="loading"
        class="w-full py-2.5 text-xs font-medium text-arena-400 hover:text-arena-200 bg-piedra-800/60 hover:bg-piedra-800 border border-piedra-700/50 rounded-xl transition-colors disabled:opacity-50"
      >
        <span v-if="loading">Loading…</span>
        <span v-else>Load more ({{ totalCount - conversations.length }} remaining)</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject, onMounted, onBeforeUnmount } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { conversationsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'

const PAGE_SIZE = 60

const emit = defineEmits(['select'])
const store = useDataStore()
const requestDelete = inject('requestDelete')
const toast = inject('toast')

const conversations = ref([])
const totalCount = ref(0)
const loading = ref(false)
const filterAgent = ref('')
const filterSource = ref('')
const refreshInterval = ref(0)
const refreshPulse = ref(false)
const refreshOptions = [
  { label: 'Off', value: 0 },
  { label: '5s', value: 5000 },
  { label: '30s', value: 30000 },
]

let refreshTimer = null

const hasFilters = computed(() => filterAgent.value || filterSource.value)
const hasMore = computed(() => conversations.value.length < totalCount.value)

const grouped = computed(() => {
  const seen = new Map()
  const result = []
  for (const c of conversations.value) {
    const key = `${c.sessionId}::${c.agentId}`
    const existing = seen.get(key)
    if (!existing) {
      seen.set(key, { ...c, hasPair: false })
      result.push(seen.get(key))
    } else {
      existing.hasPair = true
      if (c.perspective === 'user' && existing.perspective !== 'user') {
        existing.id = c.id
        existing.perspective = c.perspective
        existing.preview = c.preview || existing.preview
        existing.summary = c.summary || existing.summary
      }
    }
  }
  return result
})

async function loadConversations(offset = 0) {
  loading.value = true
  try {
    const params = { limit: PAGE_SIZE, offset }
    if (filterAgent.value) params.agentId = filterAgent.value
    if (filterSource.value) params.source = filterSource.value
    const result = await conversationsApi.list(params)
    if (offset === 0) {
      conversations.value = result.items || []
    } else {
      conversations.value = [...conversations.value, ...(result.items || [])]
    }
    totalCount.value = result.total || 0
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

function resetAndLoad() {
  conversations.value = []
  totalCount.value = 0
  loadConversations(0)
}

function silentRefresh() {
  loadConversations(0)
}

function autoRefreshTick() {
  refreshPulse.value = true
  silentRefresh()
  setTimeout(() => { refreshPulse.value = false }, 400)
}

function loadMore() {
  loadConversations(conversations.value.length)
}

function setAutoRefresh(ms) {
  refreshInterval.value = ms
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  if (ms > 0) {
    autoRefreshTick()
    refreshTimer = setInterval(autoRefreshTick, ms)
  }
}

watch([filterAgent, filterSource], () => resetAndLoad())
onMounted(() => loadConversations(0))
onBeforeUnmount(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

function handleClearAll() {
  requestDelete('Clear ALL conversation logs? This cannot be undone.', async () => {
    try {
      await conversationsApi.clear()
      conversations.value = []
      totalCount.value = 0
      toast.success('All conversations cleared')
    } catch (e) {
      toast.error(e.message)
    }
  })
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  const diff = now - d

  if (diff < 60000) return 'just now'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`

  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) +
    ' ' + d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
}

function sourceIcon(source) {
  const map = {
    'voice-ui': 'phone',
    telegram: 'phone',
    executor: 'command',
    flow: 'flow',
    direct: 'phone',
    cron: 'clock',
    webhook: 'bolt',
  }
  return map[source] || 'chat'
}

function sourceBg(source) {
  const map = {
    'voice-ui': 'bg-teal-500/15',
    telegram: 'bg-atlantico-500/15',
    executor: 'bg-indigo-500/15',
    flow: 'bg-rose-500/15',
    direct: 'bg-teal-500/15',
    cron: 'bg-sol-500/15',
    webhook: 'bg-purple-500/15',
  }
  return map[source] || 'bg-piedra-800/60'
}

function sourceText(source) {
  const map = {
    'voice-ui': 'text-teal-400',
    telegram: 'text-atlantico-400',
    executor: 'text-indigo-400',
    flow: 'text-rose-400',
    direct: 'text-teal-400',
    cron: 'text-sol-400',
    webhook: 'text-purple-400',
  }
  return map[source] || 'text-arena-500'
}

function formatSource(source) {
  const map = {
    'voice-ui': 'Voice UI',
    telegram: 'Telegram',
    executor: 'Executor',
    flow: 'Flow',
    direct: 'Direct',
    cron: 'Cron',
    webhook: 'Webhook',
  }
  return map[source] || source
}

defineExpose({ refresh: resetAndLoad })
</script>
