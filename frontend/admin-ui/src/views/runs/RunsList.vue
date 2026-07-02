<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Runs</h2>
      <div class="flex items-center gap-3">
        <!-- Auto-refresh segmented control -->
        <SegmentedControl :modelValue="refreshInterval" @update:modelValue="setAutoRefresh" :options="refreshOptions" />
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
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap items-center gap-2">
      <!-- App Select -->
      <select
        v-model="filterApp"
        class="bg-piedra-800 border border-piedra-700/50 text-arena-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-piedra-600"
      >
        <option value="">All apps</option>
        <option v-for="a in store.agents" :key="a.id" :value="a.id">{{ a.name }}</option>
        <option v-for="f in store.flows" :key="f.id" :value="f.id">{{ f.name }} (flow)</option>
      </select>

      <!-- Status Select -->
      <select
        v-model="filterStatus"
        class="bg-piedra-800 border border-piedra-700/50 text-arena-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-piedra-600"
      >
        <option value="">All status</option>
        <option value="completed">Completed</option>
        <option value="failed">Failed</option>
        <option value="interrupted">Interrupted</option>
      </select>

      <button
        v-if="filterApp || filterStatus"
        @click="filterApp = ''; filterStatus = ''"
        class="text-[10px] text-arena-500 hover:text-arena-300 px-2 py-1.5 transition-colors"
      >
        Clear filters
      </button>

      <div class="flex-1" />
      <span class="text-[10px] text-arena-500 tabular-nums">
        {{ totalCount }} run{{ totalCount !== 1 ? 's' : '' }}
      </span>
    </div>

    <!-- Loading skeleton -->
    <SkeletonCard v-if="loading && !runs.length" />

    <!-- Empty states -->
    <EmptyState
      v-else-if="!runs.length && !hasFilters"
      title="No runs yet"
      subtitle="Runs appear when agents or flows execute."
      icon="clock"
      color="sol"
    />

    <EmptyState
      v-else-if="!runs.length && hasFilters"
      title="No matching runs"
      subtitle="Try adjusting your filters"
      icon="clock"
      color="sol"
    />

    <!-- Runs List -->
    <div v-else class="space-y-2">
      <button
        v-for="r in runs"
        :key="r.runId"
        @click="$emit('select', r.runId)"
        class="w-full text-left focus:outline-none"
      >
        <div class="bg-piedra-900 border border-piedra-800 rounded-xl px-4 py-3 hover:border-piedra-700 transition-colors space-y-2">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2 min-w-0">
              <!-- Status dot -->
              <span class="w-2 h-2 rounded-full flex-shrink-0" :class="STATUS_DOT_CLASSES[r.status] || 'bg-arena-500'" />
              <!-- App Name -->
              <span class="text-sm font-medium text-arena-100 truncate">
                {{ getAppName(r.appName) }}
              </span>
              <!-- Source Badge -->
              <Badge v-if="r.source" variant="muted" class="!py-0">{{ r.source }}</Badge>
            </div>
            <!-- Duration -->
            <span class="text-[10px] text-arena-400 tabular-nums flex-shrink-0">
              {{ formatDuration(r.startedAt, r.endedAt) }}
            </span>
          </div>

          <div class="flex items-center justify-between gap-2 text-[10px] text-arena-500">
            <div class="flex items-center gap-1.5 min-w-0">
              <span class="truncate font-mono text-[9px] text-arena-600 max-w-[120px] sm:max-w-none">{{ r.runId }}</span>
              <span v-if="r.eventCount != null" class="flex items-center gap-1.5 flex-shrink-0">
                <span>•</span>
                <span>{{ r.eventCount }} event{{ r.eventCount !== 1 ? 's' : '' }}</span>
              </span>
            </div>
            <span class="tabular-nums text-arena-600 flex-shrink-0">{{ formatTime(r.startedAt) }}</span>
          </div>

          <!-- Error line -->
          <div v-if="r.status === 'failed' && r.error" class="text-lava-300 text-[10px] truncate pt-1.5 border-t border-piedra-800/40">
            {{ r.error }}
          </div>
        </div>
      </button>

      <!-- Load more button -->
      <button
        v-if="hasMore"
        @click="loadMore"
        :disabled="loading"
        class="w-full py-2.5 text-xs font-medium text-arena-400 hover:text-arena-200 bg-piedra-800/60 hover:bg-piedra-800 border border-piedra-700/50 rounded-xl transition-colors disabled:opacity-50"
      >
        <span v-if="loading">Loading…</span>
        <span v-else>Load more ({{ totalCount - runs.length }} remaining)</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject, onMounted, onBeforeUnmount } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { runsApi } from '../../lib/api/index.js'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import SegmentedControl from '../../components/SegmentedControl.vue'

const PAGE_SIZE = 30

const emit = defineEmits(['select'])
const store = useDataStore()
const toast = inject('toast', { error: console.error })

const runs = ref([])
const totalCount = ref(0)
const loading = ref(false)
const filterApp = ref('')
const filterStatus = ref('')
const refreshInterval = ref(0)
const refreshPulse = ref(false)
const refreshOptions = [
  { label: 'Off', value: 0 },
  { label: '5s', value: 5000 },
  { label: '30s', value: 30000 },
]

let refreshTimer = null

const hasFilters = computed(() => !!(filterApp.value || filterStatus.value))
const hasMore = computed(() => runs.value.length < totalCount.value)

// Full class strings per status so the Tailwind scanner sees them.
const STATUS_DOT_CLASSES = {
  completed: 'bg-green-400',
  failed: 'bg-lava-400',
  interrupted: 'bg-arena-500',
  running: 'bg-sol-400',
}

function getAppName(idOrName) {
  if (!idOrName) return 'App'
  const agent = store.agents?.find(a => a.id === idOrName || a.name === idOrName)
  if (agent) return agent.name
  const flow = store.flows?.find(f => f.id === idOrName || f.name === idOrName)
  if (flow) return `${flow.name} (flow)`
  return idOrName
}

function formatDuration(startedAt, endedAt) {
  if (!startedAt) return ''
  const start = new Date(startedAt).getTime()
  const end = endedAt ? new Date(endedAt).getTime() : Date.now()
  const ms = end - start
  if (ms < 0) return '0s'
  if (ms < 1000) return '<1s'
  const secs = Math.floor(ms / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  const remSecs = secs % 60
  if (remSecs === 0) return `${mins}m`
  return `${mins}m ${remSecs}s`
}

function formatTime(isoString) {
  if (!isoString) return ''
  return new Date(isoString).toLocaleString()
}

async function loadRuns(offset = 0) {
  loading.value = true
  try {
    const params = { limit: PAGE_SIZE, offset }
    if (filterApp.value) params.appName = filterApp.value
    if (filterStatus.value) params.status = filterStatus.value
    const result = await runsApi.list(params)
    if (offset === 0) {
      runs.value = result.items || []
    } else {
      runs.value = [...runs.value, ...(result.items || [])]
    }
    totalCount.value = result.total || 0
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

function resetAndLoad() {
  runs.value = []
  totalCount.value = 0
  loadRuns(0)
}

function silentRefresh() {
  loadRuns(0)
}

function autoRefreshTick() {
  refreshPulse.value = true
  silentRefresh()
  setTimeout(() => { refreshPulse.value = false }, 400)
}

function loadMore() {
  loadRuns(runs.value.length)
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

watch([filterApp, filterStatus], () => {
  resetAndLoad()
})

defineExpose({
  refresh() {
    filterApp.value = ''
    filterStatus.value = ''
    resetAndLoad()
  }
})

onMounted(() => {
  loadRuns(0)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>
