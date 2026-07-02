<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center gap-3 mb-2">
      <div class="flex items-center gap-3 flex-1 min-w-0">
        <!-- Back button -->
        <button
          @click="$emit('back')"
          class="w-7 h-7 rounded-lg flex items-center justify-center text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80 transition-colors flex-shrink-0"
        >
          <Icon name="back" size="sm" />
        </button>
        <div class="flex-1 min-w-0 space-y-1.5">
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-semibold text-arena-200 truncate">
              {{ getAppName(run?.appName) }}
            </h2>
            <Badge v-if="run?.status" :variant="STATUS_BADGE_VARIANTS[run.status] || 'default'">
              {{ STATUS_TEXT[run.status] || run.status }}
            </Badge>
          </div>
          <!-- Meta row -->
          <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[10px] text-arena-500 font-mono">
            <span>ID: <span class="text-arena-400">{{ run?.runId }}</span></span>
            <span v-if="run?.sessionId">Session: <span class="text-arena-400">{{ run.sessionId }}</span></span>
            <span v-if="run?.source">Source: <span class="text-arena-400">{{ run.source }}</span></span>
            <span v-if="run?.startedAt">Started: <span class="text-arena-400">{{ formatTime(run.startedAt) }}</span></span>
            <span v-if="run?.startedAt">Duration: <span class="text-arena-400">{{ formatDuration(run.startedAt, run.endedAt) }}</span></span>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Panel -->
    <div v-if="run?.error" class="bg-lava-500/10 border border-lava-500/30 rounded-xl p-3 text-lava-300 text-xs">
      <div class="font-semibold mb-1">Execution Error</div>
      <div class="font-mono whitespace-pre-wrap break-all">{{ run.error }}</div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-4">
      <SkeletonCard />
    </div>

    <!-- Activation Timeline -->
    <div v-else-if="run?.activations && run.activations.length" class="space-y-3">
      <h3 class="text-xs font-semibold text-arena-300 uppercase tracking-wider mb-2">Activation Timeline</h3>
      
      <div class="relative pl-6 border-l border-piedra-800 space-y-4 ml-3">
        <div v-for="(act, idx) in run.activations" :key="idx" class="relative">
          <!-- Circle timeline indicator aligned with node card header -->
          <span
            class="absolute -left-[30.5px] top-[15px] w-2 h-2 rounded-full border bg-piedra-950 transition-colors"
            :class="act.error ? 'border-lava-500' : 'border-piedra-500'"
          />
          
          <!-- Activation card -->
          <div class="bg-piedra-900 border border-piedra-800 rounded-xl p-3.5 space-y-2">
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2 min-w-0">
                <span class="font-mono text-xs font-semibold text-arena-100 truncate">
                  {{ act.node }}
                </span>
                <span v-if="act.seq != null" class="font-mono text-[9px] text-arena-600 bg-piedra-850 px-1 rounded">
                  #{{ act.seq }}
                </span>
              </div>
              
              <div class="flex items-center gap-1.5 flex-shrink-0">
                <!-- Duration Badge if > 0ms -->
                <Badge v-if="getActivationMs(act.startedAt, act.endedAt) > 0" variant="muted" class="!py-0">
                  {{ formatActivationDuration(act.startedAt, act.endedAt) }}
                </Badge>
                <!-- Event count -->
                <span v-if="act.events != null" class="text-[10px] text-arena-500 font-mono">
                  {{ act.events }} event{{ act.events !== 1 ? 's' : '' }}
                </span>
              </div>
            </div>

            <!-- Branch -->
            <div v-if="act.branch" class="text-[9px] text-arena-600 font-mono">
              Branch: <span class="text-arena-400">{{ act.branch }}</span>
            </div>

            <!-- Routes pills -->
            <div v-if="act.routes && (Array.isArray(act.routes) ? act.routes.length : act.routes)" class="flex flex-wrap gap-1">
              <span
                v-for="r in (Array.isArray(act.routes) ? act.routes : [act.routes])"
                :key="r"
                class="font-mono text-[9px] bg-atlantico-500/10 text-atlantico-300 rounded px-1.5 py-0.5"
              >
                {{ r }}
              </span>
            </div>

            <!-- Output Preview -->
            <div v-if="act.outputPreview" class="mt-1">
              <pre class="bg-piedra-800/60 rounded-lg p-2 text-[10px] font-mono text-arena-300 whitespace-pre-wrap break-words max-h-24 overflow-y-auto">{{ act.outputPreview }}</pre>
            </div>

            <!-- Activation Error -->
            <div v-if="act.error" class="bg-lava-500/5 border border-lava-500/20 rounded-lg p-2 text-lava-300 text-[10px] font-mono mt-1 whitespace-pre-wrap break-words">
              {{ act.error }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty activations -->
    <div v-else-if="!loading" class="text-center py-8 text-arena-500 text-xs">
      No activations recorded for this run.
    </div>

    <!-- Raw Events Section -->
    <div v-if="run" class="mt-6 border-t border-piedra-800 pt-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-xs font-semibold text-arena-300">Raw Events</h3>
        <button
          @click="toggleRaw"
          class="flex items-center gap-1 px-2.5 py-1 hover:bg-piedra-800 rounded-lg text-[10px] font-medium text-arena-400 hover:text-arena-200 transition-colors"
        >
          <Icon name="eye" size="xs" />
          <span>{{ showRaw ? 'Hide raw events' : 'Show raw events' }}</span>
        </button>
      </div>

      <div v-if="showRaw" class="space-y-2">
        <div v-if="rawLoading" class="text-xs text-arena-500 italic">
          Loading raw events…
        </div>
        <div v-else-if="!run.events || !run.events.length" class="text-xs text-arena-500 italic">
          No raw events found for this run.
        </div>
        <div v-else class="space-y-1">
          <details
            v-for="(evt, idx) in run.events"
            :key="idx"
            class="bg-piedra-900 border border-piedra-800/50 rounded-lg overflow-hidden group"
          >
            <summary class="flex items-center justify-between px-3 py-1.5 text-[10px] text-arena-400 font-mono cursor-pointer hover:bg-piedra-800/50 select-none">
              <span>{{ getEventSummary(evt, idx) }}</span>
              <Icon name="chevronDown" size="xs" class="text-arena-600 group-open:rotate-180 transition-transform" />
            </summary>
            <div class="p-3 border-t border-piedra-800/30 bg-piedra-950/40">
              <pre class="font-mono text-[9px] text-arena-400 max-h-48 overflow-auto whitespace-pre-wrap break-all">{{ safeFormatJson(evt) }}</pre>
            </div>
          </details>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { runsApi } from '../../lib/api/index.js'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'

const props = defineProps({
  runId: { type: String, required: true }
})

defineEmits(['back'])

const store = useDataStore()
const toast = inject('toast', { error: console.error })

const run = ref(null)
const loading = ref(false)
const showRaw = ref(false)
const rawLoading = ref(false)

const STATUS_BADGE_VARIANTS = {
  completed: 'green',
  failed: 'lava',
  interrupted: 'default'
}

const STATUS_TEXT = {
  completed: 'Completed',
  failed: 'Failed',
  interrupted: 'Interrupted'
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

function getActivationMs(startedAt, endedAt) {
  if (!startedAt || !endedAt) return 0
  return new Date(endedAt).getTime() - new Date(startedAt).getTime()
}

function formatActivationDuration(startedAt, endedAt) {
  const ms = getActivationMs(startedAt, endedAt)
  if (ms <= 0) return ''
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

async function loadRunDetail() {
  loading.value = true
  try {
    const data = await runsApi.get(props.runId)
    run.value = data
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

async function toggleRaw() {
  showRaw.value = !showRaw.value
  if (showRaw.value && run.value && !run.value.events) {
    rawLoading.value = true
    try {
      const data = await runsApi.get(props.runId, { raw: true })
      run.value = data
    } catch (e) {
      toast.error('Failed to load raw events: ' + e.message)
    } finally {
      rawLoading.value = false
    }
  }
}

function safeFormatJson(evt) {
  if (typeof evt === 'string') {
    try {
      return JSON.stringify(JSON.parse(evt), null, 2)
    } catch (e) {
      return evt
    }
  }
  return JSON.stringify(evt, null, 2)
}

function getEventSummary(evt, idx) {
  let data = evt
  if (typeof evt === 'string') {
    try {
      data = JSON.parse(evt)
    } catch (e) {
      return `Event #${idx}`
    }
  }
  if (!data) return `Event #${idx}`
  const type = data.type || data.event || 'Event'
  const source = data.source || data.author || ''
  return `[${idx}] ${type}${source ? ' - ' + source : ''}`
}

onMounted(() => {
  loadRunDetail()
})
</script>
