<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Flow' : 'New Flow'" size="2xl" @save="save">
    <div class="space-y-4">
      <div class="grid grid-cols-3 gap-4">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="my-workflow" :required="true" />
        </div>
        <div class="col-span-2">
          <FormLabel label="Description" />
          <FormInput v-model="form.description" placeholder="What this flow does..." />
        </div>
      </div>
      <div class="border border-piedra-700/40 rounded-xl px-4 py-3">
        <div class="flex items-center justify-between">
          <div>
            <span class="text-xs font-medium text-arena-400">A2A Protocol</span>
            <p class="text-[10px] text-arena-500 mt-0.5">Expose this flow via the Agent-to-Agent protocol for external discovery and invocation</p>
          </div>
          <FormToggle v-model="form.a2aEnabled" />
        </div>
      </div>

      <FlowCanvas v-model="form.graph" :agents="store.agents" :flows="availableSubflows" />

      <details class="group text-arena-500">
        <summary class="text-[10px] font-medium cursor-pointer select-none hover:text-arena-300 transition-colors">
          How does the flow editor work?
        </summary>
        <div class="mt-2 text-[10px] leading-relaxed space-y-2 text-arena-500/80">
          <p>A flow is a <span class="text-arena-300 font-semibold">graph</span>: add nodes from the toolbar, then drag from a node's right port to another node's left port to connect them. Drag from <span class="text-rose-400 font-semibold">Start</span> to the node where execution begins.</p>
          <div class="grid grid-cols-3 gap-x-4 gap-y-1.5">
            <div><span class="text-sol-400 font-semibold">Agent</span> — an AI agent that processes input and produces output.</div>
            <div><span class="text-atlantico-400 font-semibold">Router</span> — evaluates CEL rules in order and routes to one branch.</div>
            <div><span class="text-purple-400 font-semibold">Join</span> — waits for all incoming branches, then forwards their combined output.</div>
          </div>
          <div class="flex items-start gap-1.5 pt-1 border-t border-piedra-700/30">
            <svg class="w-3.5 h-3.5 text-green-400 flex-shrink-0 mt-px" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.076-4.076a1.526 1.526 0 0 1 1.037-.443 48.2 48.2 0 0 0 5.887-.512c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.4 48.4 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" />
            </svg>
            <span>Each agent has a <span class="text-green-400 font-semibold">response</span> toggle. Only agents with this active are included in the flow output. If none are marked, all agent outputs are returned.</span>
          </div>
          <div class="pt-1 border-t border-piedra-700/30 space-y-1">
            <p>Agents in a flow share a <span class="text-arena-300 font-semibold">state</span> map via <code class="text-arena-300">set_state(key, value)</code> and <code class="text-arena-300">get_state(key)</code>. A <span class="text-atlantico-400 font-semibold">router</span>'s CEL rules read that state as <code class="text-arena-300">state.key</code> and the loop counter as <code class="text-arena-300">iterations</code>, e.g. <code class="text-arena-300">state.done || iterations >= 5</code>.</p>
            <p>To build a <span class="text-atlantico-400 font-semibold">loop</span>, draw an edge from a router's route back to an earlier node; the other route leaves the loop.</p>
          </div>
        </div>
      </details>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject, computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { flowsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import FormToggle from '../../components/FormToggle.vue'
import FlowCanvas from './FlowCanvas.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

// Flows available as subflow targets: every flow except the one being edited
// (a flow cannot embed itself). Deeper cycles are caught server-side by the
// topological sort at build time.
const availableSubflows = computed(() => (store.flows || []).filter(f => f.id !== editId.value))

const form = reactive({
  name: '',
  description: '',
  graph: null,
  a2aEnabled: false,
})

function open(flow = null) {
  isEdit.value = !!flow
  editId.value = flow?.id || null
  form.name = flow?.name || ''
  form.description = flow?.description || ''
  form.graph = flow
    ? { entry: flow.entry || '', nodes: deepClone(flow.nodes || []), edges: deepClone(flow.edges || []) }
    : { entry: '', nodes: [], edges: [] }
  form.a2aEnabled = flow?.a2a?.enabled || false
  dialogRef.value?.open()
}

function deepClone(v) { return JSON.parse(JSON.stringify(v)) }

async function save() {
  const g = form.graph || { entry: '', nodes: [], edges: [] }
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    entry: g.entry,
    nodes: (g.nodes || []).map(cleanNode),
    edges: (g.edges || []).map(cleanEdge),
    a2a: form.a2aEnabled ? { enabled: true } : undefined,
  }
  try {
    if (isEdit.value) {
      await flowsApi.update(editId.value, data)
    } else {
      await flowsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

// cleanNode drops empty/irrelevant fields per node type so the saved JSON is tidy.
function cleanNode(n) {
  const clean = { id: n.id, type: n.type }
  if (typeof n.x === 'number') clean.x = Math.round(n.x)
  if (typeof n.y === 'number') clean.y = Math.round(n.y)
  if (n.type === 'agent') {
    clean.agentId = n.agentId || ''
    if (n.responseAgent) clean.responseAgent = true
  } else if (n.type === 'router') {
    clean.rules = (n.rules || []).map(r => ({ when: r.when || '', route: r.route || '' }))
    clean.defaultRoute = n.defaultRoute || ''
  } else if (n.type === 'parallel') {
    clean.agentId = n.agentId || ''
    if (n.maxConcurrency) clean.maxConcurrency = n.maxConcurrency
  } else if (n.type === 'subflow') {
    clean.flowId = n.flowId || ''
  }
  return clean
}

function cleanEdge(e) {
  const clean = { from: e.from, to: e.to }
  if (e.route) clean.route = e.route
  return clean
}

defineExpose({ open })
</script>
