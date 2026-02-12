<template>
  <div class="flow-editor" style="height: 500px;">
    <VueFlow
      id="flow-editor"
      :nodes="nodes"
      :edges="edges"
      :default-edge-options="{ type: 'smoothstep', animated: true, style: { stroke: '#6b7280' } }"
      :fit-view-on-init="true"
      @nodes-change="onNodesChange"
      @edges-change="onEdgesChange"
      @connect="onConnect"
    >
      <Background />
      <Controls position="bottom-left" />
      <MiniMap position="bottom-right" />

      <template #node-agent="nodeProps">
        <AgentNode v-bind="nodeProps" :agents="agents" @update-data="updateNodeData(nodeProps.id, $event)" @delete-node="deleteNode(nodeProps.id)" />
      </template>
      <template #node-sequential="nodeProps">
        <ContainerNode v-bind="nodeProps" label="Sequential" color="atlantico" @delete-node="deleteNode(nodeProps.id)" />
      </template>
      <template #node-parallel="nodeProps">
        <ContainerNode v-bind="nodeProps" label="Parallel" color="sol" @delete-node="deleteNode(nodeProps.id)" />
      </template>
      <template #node-loop="nodeProps">
        <LoopNode v-bind="nodeProps" @update-data="updateNodeData(nodeProps.id, $event)" @delete-node="deleteNode(nodeProps.id)" />
      </template>

      <Panel position="top-left" class="flex gap-1.5">
        <button
          v-for="t in stepTypes"
          :key="t.type"
          @click="addNode(t.type)"
          class="px-2.5 py-1.5 text-[10px] font-medium rounded-lg border transition-colors"
          :class="t.class"
        >
          + {{ t.label }}
        </button>
      </Panel>
    </VueFlow>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { VueFlow, Panel, useVueFlow, applyNodeChanges, applyEdgeChanges } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import AgentNode from './nodes/AgentNode.vue'
import ContainerNode from './nodes/ContainerNode.vue'
import LoopNode from './nodes/LoopNode.vue'

const props = defineProps({
  modelValue: { type: Object, default: () => ({ type: 'sequential', steps: [] }) },
  agents: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:modelValue'])

const stepTypes = [
  { type: 'agent', label: 'Agent', class: 'bg-sol-500/10 border-sol-500/30 text-sol-300 hover:bg-sol-500/20' },
  { type: 'sequential', label: 'Sequential', class: 'bg-atlantico-500/10 border-atlantico-500/30 text-atlantico-300 hover:bg-atlantico-500/20' },
  { type: 'parallel', label: 'Parallel', class: 'bg-sol-500/10 border-sol-500/30 text-sol-300 hover:bg-sol-500/20' },
  { type: 'loop', label: 'Loop', class: 'bg-lava-500/10 border-lava-500/30 text-lava-300 hover:bg-lava-500/20' },
]

const { fitView } = useVueFlow({ id: 'flow-editor' })

const nodes = ref([])
const edges = ref([])
let nodeIdCounter = 0
let suppressSync = false

function generateId() {
  return `node_${++nodeIdCounter}_${Date.now()}`
}

function treeToGraph(step, parentId = null, depth = 0, index = 0) {
  const id = generateId()
  const x = 250 * depth
  const y = 120 * index
  const resultNodes = []
  const resultEdges = []

  const node = {
    id,
    type: step.type,
    position: { x, y },
    data: { ...step },
  }
  delete node.data.steps
  resultNodes.push(node)

  if (parentId) {
    resultEdges.push({
      id: `e-${parentId}-${id}`,
      source: parentId,
      target: id,
      type: 'smoothstep',
      animated: true,
      style: { stroke: '#6b7280' },
    })
  }

  if (step.steps?.length) {
    step.steps.forEach((child, i) => {
      const { nodes: cn, edges: ce } = treeToGraph(child, id, depth + 1, i)
      resultNodes.push(...cn)
      resultEdges.push(...ce)
    })
  }

  return { nodes: resultNodes, edges: resultEdges }
}

function graphToTree(nodesArr, edgesArr) {
  if (!nodesArr.length) return { type: 'sequential', steps: [] }

  const childMap = {}
  for (const e of edgesArr) {
    if (!childMap[e.source]) childMap[e.source] = []
    childMap[e.source].push(e.target)
  }

  const targetIds = new Set(edgesArr.map(e => e.target))
  const roots = nodesArr.filter(n => !targetIds.has(n.id))

  function buildStep(node) {
    const step = { type: node.type }
    if (node.type === 'agent') {
      step.agentId = node.data?.agentId || ''
    }
    if (node.type === 'loop') {
      step.maxIterations = node.data?.maxIterations || 0
    }
    if (['sequential', 'parallel', 'loop'].includes(node.type)) {
      const childIds = childMap[node.id] || []
      const childNodes = childIds.map(cid => nodesArr.find(n => n.id === cid)).filter(Boolean)
      childNodes.sort((a, b) => a.position.y - b.position.y)
      step.steps = childNodes.map(buildStep)
    }
    return step
  }

  if (roots.length === 1) {
    return buildStep(roots[0])
  }

  roots.sort((a, b) => a.position.y - b.position.y)
  return {
    type: 'sequential',
    steps: roots.map(buildStep),
  }
}

function loadFromTree(step) {
  nodeIdCounter = 0
  const graph = treeToGraph(step)
  suppressSync = true
  nodes.value = graph.nodes
  edges.value = graph.edges
  suppressSync = false
  nextTick(() => {
    setTimeout(() => fitView({ padding: 0.5, maxZoom: 0.85 }), 50)
  })
}

function syncToTree() {
  syncToTreeGuarded()
}

function onNodesChange(changes) {
  nodes.value = applyNodeChanges(changes, nodes.value)
  syncToTree()
}

function onEdgesChange(changes) {
  edges.value = applyEdgeChanges(changes, edges.value)
  syncToTree()
}

function onConnect(connection) {
  const newEdge = {
    id: `e-${connection.source}-${connection.target}`,
    source: connection.source,
    target: connection.target,
    type: 'smoothstep',
    animated: true,
    style: { stroke: '#6b7280' },
  }
  edges.value = [...edges.value, newEdge]
  syncToTree()
}

function addNode(type) {
  const id = generateId()
  const data = { type }
  if (type === 'agent') data.agentId = ''
  if (type === 'loop') data.maxIterations = 0

  const newNode = {
    id,
    type,
    position: { x: 100 + Math.random() * 200, y: 100 + Math.random() * 200 },
    data,
  }
  nodes.value = [...nodes.value, newNode]
  syncToTree()
}

function updateNodeData(nodeId, newData) {
  nodes.value = nodes.value.map(n =>
    n.id === nodeId ? { ...n, data: { ...n.data, ...newData } } : n
  )
  syncToTree()
}

function deleteNode(nodeId) {
  nodes.value = nodes.value.filter(n => n.id !== nodeId)
  edges.value = edges.value.filter(e => e.source !== nodeId && e.target !== nodeId)
  syncToTree()
}

let lastEmittedJson = ''

function syncToTreeGuarded() {
  if (suppressSync) return
  const tree = graphToTree(nodes.value, edges.value)
  const json = JSON.stringify(tree)
  if (json === lastEmittedJson) return
  lastEmittedJson = json
  emit('update:modelValue', tree)
}

watch(() => props.modelValue, (val) => {
  if (!val) return
  const json = JSON.stringify(val)
  if (json === lastEmittedJson) return
  loadFromTree(val)
}, { immediate: true, deep: true })

defineExpose({ loadFromTree })
</script>

<style scoped>
.flow-editor {
  border-radius: 0.75rem;
  overflow: hidden;
  border: 1px solid rgba(120, 113, 108, 0.4);
}

.flow-editor :deep(.vue-flow__background) {
  background-color: #1c1917;
}

.flow-editor :deep(.vue-flow__minimap) {
  background-color: #292524;
  border-radius: 0.5rem;
  overflow: hidden;
}

.flow-editor :deep(.vue-flow__controls) {
  background-color: #292524;
  border-radius: 0.5rem;
  overflow: hidden;
  border: 1px solid rgba(120, 113, 108, 0.3);
}

.flow-editor :deep(.vue-flow__controls-button) {
  background-color: #292524;
  border-color: rgba(120, 113, 108, 0.3);
  fill: #a8a29e;
}

.flow-editor :deep(.vue-flow__controls-button:hover) {
  background-color: #44403c;
}
</style>
