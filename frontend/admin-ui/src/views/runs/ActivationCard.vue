<template>
  <div class="relative">
    <div
      class="bg-piedra-900 border border-piedra-800 rounded-xl overflow-hidden transition-colors"
      :class="expanded ? '' : 'hover:border-piedra-700'"
    >
      <!-- Header row (click to expand): type icon + type label, a vertical
           separator, then the node name. Everything else lives in the
           expanded panel. -->
      <button
        type="button"
        @click="$emit('toggle')"
        class="w-full text-left flex items-center gap-2 px-3.5 py-2.5 cursor-pointer focus:outline-none select-none"
      >
        <template v-if="nodeType">
          <span class="w-24 flex items-center gap-2 flex-shrink-0">
            <svg class="w-3.5 h-3.5 text-arena-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.6">
              <path stroke-linecap="round" stroke-linejoin="round" :d="nodeType.icon" />
            </svg>
            <span class="text-xs text-arena-500">{{ nodeType.label }}</span>
          </span>
          <span class="w-px self-stretch bg-piedra-700 flex-shrink-0 mr-3" />
        </template>
        <span class="text-xs font-semibold text-arena-200 truncate">{{ activation.node }}</span>
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

        <!-- State: a two-column grid so keys stay aligned. Values are
             hidden behind a per-key toggle; a sol dot marks keys written by
             this activation. -->
        <div class="space-y-1">
          <p class="text-[9px] font-semibold text-arena-500 uppercase tracking-wider">State</p>
          <div v-if="activation.stateAfter && Object.keys(activation.stateAfter).length" class="bg-piedra-800/60 rounded-lg px-2 py-1 grid grid-cols-[auto_1fr] gap-x-3">
            <template v-for="(val, key) in activation.stateAfter" :key="key">
              <div class="flex items-start gap-2 py-1">
                <span
                  class="w-1.5 h-1.5 rounded-full flex-shrink-0 mt-[3px]"
                  :class="activation.stateDelta && key in activation.stateDelta ? 'bg-sol-400' : 'bg-transparent'"
                  :title="activation.stateDelta && key in activation.stateDelta ? 'Written by this node' : ''"
                />
                <span class="font-mono text-[10px] text-sol-300 leading-tight">{{ key }}</span>
              </div>
              <div class="py-1 min-w-0 self-start">
                <p v-if="shownState[key]" class="font-mono text-[10px] text-arena-400 leading-tight break-all whitespace-pre-wrap">{{ formatStateValue(val) }}</p>
                <button
                  v-else
                  type="button"
                  @click="shownState[key] = true"
                  class="block text-[10px] text-arena-500 hover:text-arena-300 leading-tight transition-colors cursor-pointer"
                >Show content</button>
              </div>
            </template>
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
import { computed, ref } from 'vue'
import Icon from '../../components/Icon.vue'

const props = defineProps({
  activation: { type: Object, required: true },
  expanded:   { type: Boolean, default: false },
  // events is this activation's slice of the run's raw events, or null when
  // the run predates raw event capture.
  events:     { type: Array, default: null },
})

defineEmits(['toggle'])

// shownState tracks which state keys have their value revealed; values start
// hidden so large payloads do not flood the panel.
const shownState = ref({})

// NODE_TYPES mirrors the flow editor's per-type labels and icon paths (see
// FlowNode.vue), rendered here in neutral gray: the timeline is an audit
// surface, entity colors stay in the editor.
const NODE_TYPES = {
  agent: {
    label: 'Agent',
    icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z',
  },
  router: {
    label: 'Router',
    icon: 'M3 12h4l3-9 4 18 3-9h4',
  },
  join: {
    label: 'Join',
    icon: 'M7 4v5a5 5 0 005 5 5 5 0 005-5V4M12 14v6',
  },
  parallel: {
    label: 'Foreach',
    icon: 'M4 6h16M4 12h16M4 18h16',
  },
  subflow: {
    label: 'Subflow',
    icon: 'M9 4H5a1 1 0 00-1 1v4m0 6v4a1 1 0 001 1h4m6-16h4a1 1 0 011 1v4m0 6v4a1 1 0 01-1 1h-4M9 9h6v6H9z',
  },
  expression: {
    label: 'Expression',
    icon: 'M8 9l3 3-3 3m5 0h3M4 4h16a1 1 0 011 1v14a1 1 0 01-1 1H4a1 1 0 01-1-1V5a1 1 0 011-1z',
  },
  template: {
    label: 'Template',
    icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
  },
  code: {
    label: 'Code',
    icon: 'M8 9l3 3-3 3m5 0h3M5 5h14a1 1 0 011 1v12a1 1 0 01-1 1H5a1 1 0 01-1-1V6a1 1 0 011-1z',
  },
}

// nodeType is null for runs recorded before node-type snapshots existed and
// for agent-only runs; the card then shows just the name.
const nodeType = computed(() => NODE_TYPES[props.activation.nodeType] || null)

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
  color: var(--color-arena-400);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 7rem;
  overflow-y: auto;
}
</style>
