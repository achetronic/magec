<template>
  <div class="relative">
    <div
      class="bg-piedra-900 border border-piedra-800 rounded-xl overflow-hidden transition-colors"
      :class="expanded ? '' : 'hover:border-piedra-700'"
    >
      <!-- Header row (click to expand): only the node title, everything else
           lives in the expanded panel. -->
      <button
        type="button"
        @click="$emit('toggle')"
        class="w-full text-left flex items-center gap-2 px-3.5 py-2.5 cursor-pointer focus:outline-none select-none"
      >
        <span class="text-xs text-arena-500 flex-shrink-0">Node name:</span>
        <span class="font-mono text-xs font-semibold text-arena-200 truncate">{{ activation.node }}</span>
        <span class="flex-1" />
        <span v-if="activation.error" class="w-1.5 h-1.5 rounded-full bg-lava-400 flex-shrink-0" title="This node failed" />
        <Icon
          name="chevronDown"
          size="xs"
          class="text-arena-500 transition-transform duration-200 flex-shrink-0"
          :class="expanded ? 'rotate-180' : ''"
        />
      </button>

      <!-- Expanded panel -->
      <div v-if="expanded" class="border-t border-piedra-800 px-3.5 py-3 space-y-3 bg-piedra-950/30">
        <!-- Activation facts: duration, branch and emitted routes -->
        <div v-if="durationMs > 0 || activation.branch || routes.length" class="flex flex-wrap items-center gap-x-4 gap-y-1 font-mono text-[10px] text-arena-500">
          <span v-if="durationMs > 0">{{ durationText }}</span>
          <span v-if="activation.branch">branch {{ shortBranch }}</span>
          <span v-for="r in routes" :key="r">&rarr; {{ r }}</span>
        </div>

        <!-- Node error, front and center -->
        <div v-if="activation.error" class="bg-lava-500/10 border border-lava-500/30 rounded-lg p-2.5 text-lava-300 text-[10px] font-mono whitespace-pre-wrap break-words">
          {{ activation.error }}
        </div>

        <!-- Input / Output side by side -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div class="space-y-1">
            <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">Input</p>
            <pre v-if="activation.inputPreview" class="io-block">{{ activation.inputPreview }}</pre>
            <p v-else class="text-[10px] text-arena-600 italic">empty</p>
          </div>
          <div class="space-y-1">
            <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">Output</p>
            <pre v-if="activation.outputPreview" class="io-block">{{ activation.outputPreview }}</pre>
            <p v-else class="text-[10px] text-arena-600 italic">empty</p>
          </div>
        </div>

        <!-- State -->
        <div class="space-y-1">
          <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">State</p>
          <div v-if="activation.stateAfter && Object.keys(activation.stateAfter).length" class="bg-piedra-800/60 rounded-lg divide-y divide-piedra-800/60">
            <div
              v-for="(val, key) in activation.stateAfter" :key="key"
              class="flex items-start gap-2 px-2 py-1.5"
            >
              <span class="font-mono text-[10px] text-sol-300 flex-shrink-0">{{ key }}</span>
              <span class="font-mono text-[10px] text-arena-300 break-all flex-1">{{ formatStateValue(val) }}</span>
              <span
                v-if="activation.stateDelta && key in activation.stateDelta"
                class="text-[8px] bg-emerald-500/10 text-emerald-300 rounded px-1 py-px flex-shrink-0"
              >written here</span>
            </div>
          </div>
          <p v-else class="text-[10px] text-arena-600 italic">empty</p>
        </div>

        <!-- Raw events of this activation -->
        <div class="space-y-1">
          <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">
            Raw events ({{ activation.events }})
          </p>
          <div v-if="!events" class="text-[10px] text-arena-600 italic">unavailable for this run</div>
          <div v-else class="space-y-1">
            <details
              v-for="(ev, evIdx) in events" :key="evIdx"
              class="bg-piedra-950 border border-piedra-800/60 rounded-lg overflow-hidden"
            >
              <summary class="cursor-pointer flex items-center gap-2 px-2 py-1.5 text-[10px] font-mono text-arena-400 hover:text-arena-200 select-none">
                <span class="text-arena-600">#{{ activation.seq + evIdx }}</span>
                <span>{{ eventAuthor(ev) }}</span>
                <span v-if="eventRoutes(ev)" class="text-atlantico-300">&rarr; {{ eventRoutes(ev) }}</span>
                <span class="flex-1 truncate text-arena-600">{{ eventPreview(ev) }}</span>
              </summary>
              <pre class="border-t border-piedra-800/60 p-2 text-[9px] font-mono text-arena-400 whitespace-pre-wrap break-all max-h-64 overflow-y-auto">{{ prettyEvent(ev) }}</pre>
            </details>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Icon from '../../components/Icon.vue'

const props = defineProps({
  activation: { type: Object, required: true },
  expanded:   { type: Boolean, default: false },
  // events is this activation's slice of the run's raw events, or null when
  // the run predates raw event capture.
  events:     { type: Array, default: null },
})

defineEmits(['toggle'])

const routes = computed(() => {
  const r = props.activation.routes
  if (!r) return []
  return Array.isArray(r) ? r : [r]
})

// shortBranch keeps the last meaningful segments of a composite branch path
// so the facts row stays compact.
const shortBranch = computed(() => {
  const parts = props.activation.branch.split('.')
  return parts.length > 2 ? parts.slice(-2).join('.') : props.activation.branch
})

const durationMs = computed(() => {
  const a = props.activation
  if (!a.startedAt || !a.endedAt) return 0
  return new Date(a.endedAt).getTime() - new Date(a.startedAt).getTime()
})

const durationText = computed(() => {
  const ms = durationMs.value
  if (ms < 1000) return `${ms}ms`
  return `${(ms / 1000).toFixed(2)}s`
})

function eventAuthor(ev) {
  if (!ev) return 'unknown'
  return ev.Author || ev.author || 'unknown'
}

function eventRoutes(ev) {
  if (!ev) return ''
  const r = ev.Routes || ev.routes
  if (!r) return ''
  return Array.isArray(r) ? r.join(', ') : String(r)
}

function eventPreview(ev) {
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
