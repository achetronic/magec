<template>
  <div class="space-y-5">
    <!-- Header -->
    <RunHeader
      :run="run"
      :appName="appName"
      :appKind="appKind"
      :metaPills="metaPills"
      @back="$emit('back')"
      @delete="handleDelete"
    />

    <!-- Run error panel: the client-pill palette (soft lava wash, strong
         lava title, arena text), roomy inner padding. -->
    <div v-if="run?.error" class="bg-lava-500/10 rounded-xl p-4 space-y-1.5">
      <p class="text-[9px] font-semibold text-lava-400 uppercase tracking-wider">Execution Error</p>
      <p class="font-mono text-xs text-arena-500 whitespace-pre-wrap break-all">{{ run.error }}</p>
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
    <div v-else-if="run?.activations && run.activations.length" class="space-y-3 pt-3">
      <h3 class="text-xs font-semibold text-arena-300 uppercase tracking-wider">Timeline</h3>

      <!-- The vertical line is only a reading hint, kept deliberately faint. -->
      <div class="relative pl-6 border-l border-piedra-800/50 space-y-4 ml-3">
        <ActivationCard
          v-for="(act, idx) in run.activations"
          :key="idx"
          :activation="act"
          :expanded="!!expandedCards[idx]"
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

const emit = defineEmits(['back'])

const store = useDataStore()
const requestDelete = inject('requestDelete')
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

function handleDelete() {
  requestDelete('Delete this run? This cannot be undone.', async () => {
    try {
      await runsApi.delete(props.runId)
      toast.success('Run deleted')
      emit('back')
    } catch (e) {
      toast.error(e.message)
    }
  })
}

onMounted(loadRunDetail)
</script>
