<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Flow' : 'New Flow'" size="xl" persistent :footer-divider="false" @save="save">
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
    ? { entry: flow.entry || '', nodes: deepClone(flow.nodes || []), edges: deepClone(flow.edges || []), startX: flow.startX, startY: flow.startY }
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
  if (typeof g.startX === 'number') data.startX = Math.round(g.startX)
  if (typeof g.startY === 'number') data.startY = Math.round(g.startY)
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
  if (typeof n.w === 'number' && n.w > 0) clean.w = Math.round(n.w)
  if (typeof n.h === 'number' && n.h > 0) clean.h = Math.round(n.h)
  if (n.type === 'agent') {
    clean.agentId = n.agentId || ''
    if (n.responseAgent) clean.responseAgent = true
  } else if (n.type === 'router') {
    // Untouched placeholder rows (both fields empty) are dropped so they do
    // not trip the backend validation that requires a route per rule. The
    // otherwise route is fixed server-side, nothing to serialise.
    clean.rules = (n.rules || [])
      .filter(r => (r.when || '').trim() !== '' || (r.route || '').trim() !== '')
      .map(r => ({ when: r.when || '', route: r.route || '' }))
  } else if (n.type === 'parallel') {
    clean.agentId = n.agentId || ''
    if (n.maxConcurrency) clean.maxConcurrency = n.maxConcurrency
  } else if (n.type === 'subflow') {
    clean.flowId = n.flowId || ''
  } else if (n.type === 'expression') {
    clean.expression = n.expression || ''
    if (n.outputKey) clean.outputKey = n.outputKey
  } else if (n.type === 'template') {
    clean.template = n.template || ''
    if (n.outputKey) clean.outputKey = n.outputKey
  } else if (n.type === 'code') {
    clean.script = n.script || ''
    if (n.outputKey) clean.outputKey = n.outputKey
    if (n.timeoutMs) clean.timeoutMs = n.timeoutMs
    if (n.maxOutputBytes) clean.maxOutputBytes = n.maxOutputBytes
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
