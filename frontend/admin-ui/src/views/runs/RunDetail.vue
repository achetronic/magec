<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex flex-col lg:flex-row lg:items-center gap-3 mb-2">
      <div class="flex items-center gap-3 flex-1 min-w-0">
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

    <!-- Run error panel -->
    <div v-if="run?.error" class="bg-lava-500/10 border border-lava-500/30 rounded-xl p-3 text-lava-300 text-xs">
      <div class="font-semibold mb-1">Execution Error</div>
      <div class="font-mono whitespace-pre-wrap break-all">{{ run.error }}</div>
    </div>

    <!-- Run input -->
    <div v-if="runInput" class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-3 space-y-1">
      <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">Input</p>
      <p class="text-xs text-arena-300 whitespace-pre-wrap break-words">{{ runInput }}</p>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="space-y-4">
      <SkeletonCard />
    </div>

    <!-- Timeline -->
    <div v-else-if="run?.activations && run.activations.length" class="space-y-3">
      <div class="flex items-center justify-between">
        <h3 class="text-xs font-semibold text-arena-300 uppercase tracking-wider">Timeline</h3>
        <!-- Branch legend -->
        <div v-if="branches.length >= 2" class="flex flex-wrap items-center gap-2">
          <div
            v-for="br in branches" :key="br"
            class="flex items-center gap-1.5 rounded-full border px-2 py-0.5"
            :class="getBranchColors(br).chip"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="getBranchColors(br).dot" />
            <span class="font-mono text-[9px]" :class="getBranchColors(br).text">{{ shortBranch(br) }}</span>
          </div>
        </div>
      </div>

      <div class="relative pl-6 border-l border-piedra-800 space-y-3 ml-3">
        <div v-for="(act, idx) in run.activations" :key="idx" class="relative">
          <!-- Timeline dot -->
          <span
            class="absolute -left-[31px] top-[17px] w-2.5 h-2.5 rounded-full border-2 border-piedra-950"
            :class="act.error ? 'bg-lava-400' : getBranchColors(act.branch).dot"
          />

          <!-- Activation card -->
          <div
            class="bg-piedra-900 border border-piedra-800 rounded-xl overflow-hidden transition-colors"
            :class="expandedCards[idx] ? '' : 'hover:border-piedra-700'"
          >
            <!-- Header row (click to expand) -->
            <button
              type="button"
              @click="toggleExpand(idx)"
              class="w-full text-left flex items-center gap-2 px-3.5 py-2.5 cursor-pointer focus:outline-none select-none"
            >
              <span class="font-mono text-xs font-semibold text-arena-100 truncate">{{ act.node }}</span>
              <span
                v-if="act.branch"
                class="font-mono text-[9px] rounded px-1 flex-shrink-0"
                :class="getBranchColors(act.branch).text"
              >{{ shortBranch(act.branch) }}</span>
              <span
                v-for="r in normalizedRoutes(act)" :key="r"
                class="font-mono text-[9px] bg-atlantico-500/10 text-atlantico-300 rounded px-1.5 py-0.5 flex-shrink-0"
              >&rarr; {{ r }}</span>
              <span class="flex-1" />
              <span v-if="getActivationMs(act.startedAt, act.endedAt) > 0" class="text-[10px] text-arena-500 font-mono flex-shrink-0">
                {{ formatActivationDuration(act.startedAt, act.endedAt) }}
              </span>
              <span v-if="act.error" class="w-1.5 h-1.5 rounded-full bg-lava-400 flex-shrink-0" title="This node failed" />
              <Icon
                name="chevronDown"
                size="xs"
                class="text-arena-500 transition-transform duration-200 flex-shrink-0"
                :class="expandedCards[idx] ? 'rotate-180' : ''"
              />
            </button>

            <!-- Collapsed summary: output snippet + error line -->
            <div v-if="!expandedCards[idx] && (act.outputPreview || act.error)" class="px-3.5 pb-2.5 space-y-1.5">
              <p v-if="act.outputPreview" class="font-mono text-[10px] text-arena-400 truncate">{{ act.outputPreview }}</p>
              <p v-if="act.error" class="font-mono text-[10px] text-lava-300 truncate">{{ act.error }}</p>
            </div>

            <!-- Expanded panel -->
            <div v-if="expandedCards[idx]" class="border-t border-piedra-800 px-3.5 py-3 space-y-3 bg-piedra-950/30">
              <!-- Node error, front and center -->
              <div v-if="act.error" class="bg-lava-500/10 border border-lava-500/30 rounded-lg p-2.5 text-lava-300 text-[10px] font-mono whitespace-pre-wrap break-words">
                {{ act.error }}
              </div>

              <!-- Input / Output side by side -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <div class="space-y-1">
                  <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">Input</p>
                  <pre v-if="act.inputPreview" class="io-block">{{ act.inputPreview }}</pre>
                  <p v-else class="text-[10px] text-arena-600 italic">empty</p>
                </div>
                <div class="space-y-1">
                  <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">Output</p>
                  <pre v-if="act.outputPreview" class="io-block">{{ act.outputPreview }}</pre>
                  <p v-else class="text-[10px] text-arena-600 italic">empty</p>
                </div>
              </div>

              <!-- State -->
              <div class="space-y-1">
                <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">State</p>
                <div v-if="act.stateAfter && Object.keys(act.stateAfter).length" class="bg-piedra-800/60 rounded-lg divide-y divide-piedra-800/60">
                  <div
                    v-for="(val, key) in act.stateAfter" :key="key"
                    class="flex items-start gap-2 px-2 py-1.5"
                  >
                    <span class="font-mono text-[10px] text-sol-300 flex-shrink-0">{{ key }}</span>
                    <span class="font-mono text-[10px] text-arena-300 break-all flex-1">{{ formatStateValue(val) }}</span>
                    <span
                      v-if="act.stateDelta && key in act.stateDelta"
                      class="text-[8px] bg-emerald-500/10 text-emerald-300 rounded px-1 py-px flex-shrink-0"
                    >written here</span>
                  </div>
                </div>
                <p v-else class="text-[10px] text-arena-600 italic">empty</p>
              </div>

              <!-- Raw events of this activation -->
              <div class="space-y-1">
                <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">
                  Raw events ({{ act.events }})
                </p>
                <div v-if="!run.events" class="text-[10px] text-arena-600 italic">unavailable for this run</div>
                <div v-else class="space-y-1">
                  <details
                    v-for="(ev, evIdx) in getActivationEvents(act)" :key="evIdx"
                    class="bg-piedra-950 border border-piedra-800/60 rounded-lg overflow-hidden"
                  >
                    <summary class="cursor-pointer flex items-center gap-2 px-2 py-1.5 text-[10px] font-mono text-arena-400 hover:text-arena-200 select-none">
                      <span class="text-arena-600">#{{ act.seq + evIdx }}</span>
                      <span>{{ getEventAuthor(ev) }}</span>
                      <span v-if="getEventRoutes(ev)" class="text-atlantico-300">&rarr; {{ getEventRoutes(ev) }}</span>
                      <span class="flex-1 truncate text-arena-600">{{ getEventPreview(ev) }}</span>
                    </summary>
                    <pre class="border-t border-piedra-800/60 p-2 text-[9px] font-mono text-arena-400 whitespace-pre-wrap break-all max-h-64 overflow-y-auto">{{ prettyEvent(ev) }}</pre>
                  </details>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else-if="!loading" class="text-center py-8 text-arena-500 text-xs">
      No activations recorded for this run.
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted, computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { runsApi } from '../../lib/api/index.js'
import { stripMetadata } from '../../lib/metadata.js'
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
const expandedCards = ref({})

// runInput is the user message that started the run, with client metadata
// blocks stripped for display.
const runInput = computed(() => run.value?.input ? stripMetadata(run.value.input).trim() : '')

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

// Full literal class strings per branch lane so the Tailwind scanner sees
// them. Branch identity is expressed through the timeline dot, the branch
// pill and the legend chip.
const BRANCH_COLORS = [
  { dot: 'bg-atlantico-400', text: 'text-atlantico-300', chip: 'border-atlantico-500/30 bg-atlantico-500/10' },
  { dot: 'bg-teal-400', text: 'text-teal-300', chip: 'border-teal-500/30 bg-teal-500/10' },
  { dot: 'bg-purple-400', text: 'text-purple-300', chip: 'border-purple-500/30 bg-purple-500/10' },
  { dot: 'bg-rose-400', text: 'text-rose-300', chip: 'border-rose-500/30 bg-rose-500/10' },
  { dot: 'bg-emerald-400', text: 'text-emerald-300', chip: 'border-emerald-500/30 bg-emerald-500/10' },
]

const NEUTRAL_BRANCH = { dot: 'bg-arena-500', text: 'text-arena-500', chip: 'border-piedra-700 bg-piedra-800/60' }

const branches = computed(() => {
  if (!run.value?.activations) return []
  const list = []
  for (const act of run.value.activations) {
    if (act.branch && !list.includes(act.branch)) list.push(act.branch)
  }
  return list
})

function getBranchColors(branch) {
  if (!branch) return NEUTRAL_BRANCH
  const idx = branches.value.indexOf(branch)
  if (idx === -1) return NEUTRAL_BRANCH
  return BRANCH_COLORS[idx % BRANCH_COLORS.length]
}

// shortBranch keeps the last meaningful segment of a composite branch path so
// legend chips stay compact ("flow.writer@1" instead of the full dotted path).
function shortBranch(branch) {
  const parts = branch.split('.')
  return parts.length > 2 ? parts.slice(-2).join('.') : branch
}

function normalizedRoutes(act) {
  if (!act.routes) return []
  return Array.isArray(act.routes) ? act.routes : [act.routes]
}

function toggleExpand(idx) {
  expandedCards.value[idx] = !expandedCards.value[idx]
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
    run.value = await runsApi.get(props.runId, { raw: true })
  } catch (e) {
    toast.error(e.message)
  } finally {
    loading.value = false
  }
}

function getActivationEvents(act) {
  if (!run.value?.events || act.seq == null || act.events == null) return []
  return run.value.events.slice(act.seq, act.seq + act.events)
}

function getEventAuthor(ev) {
  if (!ev) return 'unknown'
  return ev.Author || ev.author || 'unknown'
}

function getEventRoutes(ev) {
  if (!ev) return ''
  const r = ev.Routes || ev.routes
  if (!r) return ''
  return Array.isArray(r) ? r.join(', ') : String(r)
}

function getEventPreview(ev) {
  if (!ev) return ''
  try {
    const str = JSON.stringify(ev)
    return str.length > 80 ? str.slice(0, 80) + '...' : str
  } catch (e) {
    return String(ev).slice(0, 80)
  }
}

function prettyEvent(ev) {
  try {
    return JSON.stringify(ev, null, 2)
  } catch (e) {
    return String(ev)
  }
}

function formatStateValue(val) {
  if (val === undefined) return 'undefined'
  try {
    return JSON.stringify(val)
  } catch (e) {
    return String(val)
  }
}

onMounted(loadRunDetail)
</script>

<style scoped>
/* Shared look for the input/output blocks in the expanded panel. */
.io-block {
  background: rgba(33, 33, 37, 0.6);
  border-radius: 0.5rem;
  padding: 0.5rem;
  font-family: ui-monospace, monospace;
  font-size: 10px;
  color: var(--color-arena-300);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 7rem;
  overflow-y: auto;
}
</style>
