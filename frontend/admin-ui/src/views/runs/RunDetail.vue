<template>
  <div class="space-y-5">
    <!-- Header -->
    <RunHeader
      :run="run"
      :appName="appName"
      :appKind="appKind"
      :metaPills="metaPills"
      @back="$emit('back')"
    />

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
        <ActivationCard
          v-for="(act, idx) in run.activations"
          :key="idx"
          :activation="act"
          :expanded="!!expandedCards[idx]"
          :colors="getBranchColors(act.branch)"
          :events="activationEvents(act)"
          @toggle="toggleExpand(idx)"
        />
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
import SkeletonCard from '../../components/SkeletonCard.vue'
import RunHeader from './RunHeader.vue'
import ActivationCard from './ActivationCard.vue'

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

// metaPills flattens the client metadata the flow prefilter parked in state
// (state.magec_meta) into key/value pairs for the header. Platform prefixes
// are trimmed from keys for readability.
const metaPills = computed(() => {
  const acts = run.value?.activations
  if (!acts?.length) return []
  const meta = acts.find(a => a.stateAfter?.magec_meta)?.stateAfter?.magec_meta
  if (!meta || typeof meta !== 'object') return []
  return Object.entries(meta)
    .filter(([, v]) => v !== null && v !== '' && typeof v !== 'object')
    .map(([k, v]) => ({ key: k.replace(/^(telegram|discord|slack)_/, ''), value: String(v) }))
    .sort((a, b) => a.key.localeCompare(b.key))
})

const appName = computed(() => {
  const idOrName = run.value?.appName
  if (!idOrName) return 'App'
  const agent = store.agents?.find(a => a.id === idOrName || a.name === idOrName)
  if (agent) return agent.name
  const flow = store.flows?.find(f => f.id === idOrName || f.name === idOrName)
  if (flow) return flow.name
  return idOrName
})

const appKind = computed(() => {
  const idOrName = run.value?.appName
  if (store.flows?.find(f => f.id === idOrName || f.name === idOrName)) return 'flow'
  return 'agent'
})

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

// shortBranch keeps the last meaningful segments of a composite branch path
// so legend chips stay compact.
function shortBranch(branch) {
  const parts = branch.split('.')
  return parts.length > 2 ? parts.slice(-2).join('.') : branch
}

function toggleExpand(idx) {
  expandedCards.value[idx] = !expandedCards.value[idx]
}

// activationEvents slices the run's raw events for one activation, or null
// when the run predates raw event capture.
function activationEvents(act) {
  if (!run.value?.events || act.seq == null || act.events == null) return null
  return run.value.events.slice(act.seq, act.seq + act.events)
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

onMounted(loadRunDetail)
</script>
