<template>
  <div
    class="flow-canvas relative"
    ref="canvasRef"
    @wheel.prevent="onWheel"
    @pointerdown="onCanvasPointerDown"
    @pointermove="onCanvasPointerMove"
    @pointerup="onCanvasPointerUp"
  >
    <!-- Toolbar -->
    <div class="flow-toolbar">
      <span class="text-[9px] text-arena-500 uppercase tracking-wider font-semibold px-1 select-none">Add node</span>
      <button v-for="t in nodeTypes" :key="t.type" class="flow-tool-btn" :class="t.cls" @click="addNode(t.type)" :title="t.title">
        <span class="w-4 h-4 rounded flex items-center justify-center flex-shrink-0" :class="t.iconBg">
          <svg class="w-2.5 h-2.5" :class="t.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.7">
            <path stroke-linecap="round" stroke-linejoin="round" :d="t.icon" />
          </svg>
        </span>
        <span class="text-[10px] font-medium">{{ t.label }}</span>
      </button>
    </div>

    <!-- Canvas content (panned + zoomed) -->
    <div class="flow-canvas-inner" ref="innerRef" :style="canvasTransform">
      <!-- edge layer -->
      <svg class="flow-edges" :width="CANVAS_SIZE" :height="CANVAS_SIZE">
        <defs>
          <marker id="arrow" markerWidth="10" markerHeight="10" refX="7" refY="3" orient="auto" markerUnits="strokeWidth">
            <path d="M0,0 L7,3 L0,6 Z" fill="var(--color-arena-500)" />
          </marker>
          <marker id="arrow-rose" markerWidth="10" markerHeight="10" refX="7" refY="3" orient="auto" markerUnits="strokeWidth">
            <path d="M0,0 L7,3 L0,6 Z" fill="#fb7185" />
          </marker>
        </defs>

        <!-- entry edge: START -> entry node -->
        <path
          v-if="entryPath"
          :d="entryPath"
          fill="none"
          stroke="#fb7185"
          stroke-width="2"
          stroke-dasharray="5 4"
          marker-end="url(#arrow-rose)"
          opacity="0.9"
        />

        <!-- real edges -->
        <g v-for="(e, i) in edgePaths" :key="'e' + i">
          <path
            :d="e.d" fill="none"
            :stroke="e.selected ? '#fb7185' : 'var(--color-arena-500)'"
            :stroke-width="e.selected ? 2.5 : 1.8"
            marker-end="url(#arrow)"
            class="flow-edge-line"
            @pointerdown.stop="selectEdge(e.index)"
          />
          <!-- route label -->
          <g v-if="e.route" :transform="`translate(${e.mid.x}, ${e.mid.y})`">
            <rect :x="-e.route.length * 3.4 - 6" y="-8" :width="e.route.length * 6.8 + 12" height="16" rx="8"
              fill="var(--color-piedra-950)" stroke="var(--color-atlantico-500)" stroke-opacity="0.5" />
            <text x="0" y="3" text-anchor="middle" class="flow-edge-label">{{ e.route }}</text>
          </g>
          <!-- delete button on selected edge -->
          <g v-if="e.selected" :transform="`translate(${e.mid.x}, ${e.mid.y + (e.route ? 16 : 0)})`" class="cursor-pointer" @pointerdown.stop="removeEdge(e.index)">
            <circle r="8" fill="var(--color-lava-500)" />
            <path d="M-3,-3 L3,3 M3,-3 L-3,3" stroke="white" stroke-width="1.5" stroke-linecap="round" />
          </g>
        </g>

        <!-- ghost edge while connecting -->
        <path v-if="ghostPath" :d="ghostPath" fill="none" stroke="#fb7185" stroke-width="2" stroke-dasharray="4 4" opacity="0.7" />
      </svg>

      <!-- START sentinel -->
      <div class="flow-start" :style="{ left: startPos.x + 'px', top: startPos.y + 'px' }" @pointerdown.stop="onStartPointerDown">
        <span class="flow-start-pill">
          <svg class="w-3 h-3 text-rose-300" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z" /></svg>
          Start
        </span>
        <span class="flow-port flow-port-out flow-start-port" data-port="out:__start__|" title="Drag to the entry node" @pointerdown.stop.prevent="startEdge({ nodeId: '__start__', route: '' })" />
      </div>

      <!-- nodes -->
      <FlowNode
        v-for="n in nodes" :key="n.id"
        :node="n"
        :agents="agents"
        :flows="flows"
        :is-entry="n.id === entry"
        :selected="n.id === selectedNode"
        :connecting-active="connecting !== null"
        @update="updateNode"
        @remove="removeNode(n.id)"
        @start-edge="startEdge"
        @end-edge="endEdge"
        @pointerdown-body="onNodePointerDown($event, n)"
      />
    </div>

    <!-- empty state -->
    <div v-if="!nodes.length" class="absolute inset-0 flex items-center justify-center pointer-events-none">
      <div class="text-center">
        <p class="text-sm font-medium text-arena-300">Build your workflow graph</p>
        <p class="text-[10px] text-arena-500 mt-1">Add nodes from the toolbar, then drag from a node's right port to another's left port to connect them.</p>
        <p class="text-[10px] text-rose-400/80 mt-1.5">Drag from <span class="font-semibold">Start</span> to mark the entry node.</p>
      </div>
    </div>

    <!-- bottom bar: zoom -->
    <div class="flow-bottom-bar">
      <button @click="zoomOut" class="flow-zoom-btn" title="Zoom out">−</button>
      <span class="text-[9px] text-arena-500 select-none w-8 text-center">{{ Math.round(scale * 100) }}%</span>
      <button @click="zoomIn" class="flow-zoom-btn" title="Zoom in">+</button>
      <div class="w-px h-3.5 bg-piedra-700/40 mx-0.5"></div>
      <button @click="fitView" class="flow-zoom-btn" title="Fit / center">
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8V5a2 2 0 012-2h3m8 0h3a2 2 0 012 2v3m0 8v3a2 2 0 01-2 2h-3m-8 0H5a2 2 0 01-2-2v-3" /></svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick, onMounted } from 'vue'
import FlowNode from './FlowNode.vue'

const CANVAS_SIZE = 4000

const props = defineProps({
  // { entry: string, nodes: FlowNode[], edges: FlowEdge[] }
  modelValue: { type: Object, default: null },
  agents:     { type: Array, default: () => [] },
  flows:      { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue'])

// ── model accessors ──────────────────────────────────────────────────────────
const graph = computed(() => props.modelValue || { entry: '', nodes: [], edges: [] })
const nodes = computed(() => graph.value.nodes || [])
const edges = computed(() => graph.value.edges || [])
const entry = computed(() => graph.value.entry || '')

function commit(next) {
  emit('update:modelValue', next)
}
function patch(partial) {
  commit({ entry: entry.value, nodes: nodes.value, edges: edges.value, ...partial })
}

// ── node toolbar ─────────────────────────────────────────────────────────────
const nodeTypes = [
  { type: 'agent', label: 'Agent', title: 'An AI agent that processes input and produces output',
    cls: 'border-sol-500/30 hover:border-sol-500/60', iconBg: 'bg-sol-500/15', iconColor: 'text-sol-400',
    icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z' },
  { type: 'router', label: 'Router', title: 'Evaluates CEL rules and routes to one of several branches',
    cls: 'border-atlantico-500/30 hover:border-atlantico-500/60', iconBg: 'bg-atlantico-500/15', iconColor: 'text-atlantico-400',
    icon: 'M3 12h4l3-9 4 18 3-9h4' },
  { type: 'join', label: 'Join', title: 'Fan-in barrier: waits for all incoming branches',
    cls: 'border-purple-500/30 hover:border-purple-500/60', iconBg: 'bg-purple-500/15', iconColor: 'text-purple-400',
    icon: 'M7 4v5a5 5 0 005 5 5 5 0 005-5V4M12 14v6' },
  { type: 'parallel', label: 'Parallel', title: 'Runs an agent once per item of a list input, concurrently',
    cls: 'border-lava-500/30 hover:border-lava-500/60', iconBg: 'bg-lava-500/15', iconColor: 'text-lava-400',
    icon: 'M4 6h16M4 12h16M4 18h16' },
  { type: 'subflow', label: 'Subflow', title: 'Embed another flow as a nested workflow',
    cls: 'border-rose-500/30 hover:border-rose-500/60', iconBg: 'bg-rose-500/15', iconColor: 'text-rose-400',
    icon: 'M9 4H5a1 1 0 00-1 1v4m0 6v4a1 1 0 001 1h4m6-16h4a1 1 0 011 1v4m0 6v4a1 1 0 01-1 1h-4M9 9h6v6H9z' },
]

let idCounter = 0
function genId(type) {
  // Unique, pattern-safe id. Bump past any existing numeric suffix collisions.
  const existing = new Set(nodes.value.map(n => n.id))
  do { idCounter++ } while (existing.has(`${type}_${idCounter}`))
  return `${type}_${idCounter}`
}

function addNode(type) {
  const id = genId(type)
  // Place near the centre of the current viewport, in canvas coords.
  const rect = canvasRef.value.getBoundingClientRect()
  const cx = (rect.width / 2 - panX.value) / scale.value - 105
  const cy = (rect.height / 2 - panY.value) / scale.value - 30
  const node = { id, type, x: Math.round(cx), y: Math.round(cy) }
  if (type === 'agent') node.agentId = ''
  if (type === 'parallel') { node.agentId = ''; node.maxConcurrency = 0 }
  if (type === 'subflow') node.flowId = ''
  if (type === 'router') { node.rules = [{ when: '', route: '' }]; node.defaultRoute = 'default' }
  const nextNodes = [...nodes.value, node]
  // First node added becomes the entry by default.
  patch({ nodes: nextNodes, entry: entry.value || id })
  nextTick(measurePorts)
}

function updateNode(updated) {
  patch({ nodes: nodes.value.map(n => n.id === updated.id ? updated : n) })
  nextTick(measurePorts)
}

function removeNode(id) {
  const nextNodes = nodes.value.filter(n => n.id !== id)
  const nextEdges = edges.value.filter(e => e.from !== id && e.to !== id)
  patch({
    nodes: nextNodes,
    edges: nextEdges,
    entry: entry.value === id ? (nextNodes[0]?.id || '') : entry.value,
  })
  if (selectedNode.value === id) selectedNode.value = null
  nextTick(measurePorts)
}

// ── selection ────────────────────────────────────────────────────────────────
const selectedNode = ref(null)
const selectedEdge = ref(null)
function selectEdge(i) { selectedEdge.value = i; selectedNode.value = null }
function removeEdge(i) {
  patch({ edges: edges.value.filter((_, idx) => idx !== i) })
  selectedEdge.value = null
  nextTick(measurePorts)
}

// ── pan / zoom ───────────────────────────────────────────────────────────────
const canvasRef = ref(null)
const innerRef = ref(null)
const panX = ref(0)
const panY = ref(0)
const scale = ref(1)
const canvasTransform = computed(() => ({
  transform: `translate(${panX.value}px, ${panY.value}px) scale(${scale.value})`,
  transformOrigin: '0 0',
}))

function onWheel(e) {
  if (e.ctrlKey || e.metaKey) {
    const delta = e.deltaY > 0 ? 0.9 : 1.1
    const ns = Math.min(2, Math.max(0.3, scale.value * delta))
    const rect = canvasRef.value.getBoundingClientRect()
    const mx = e.clientX - rect.left
    const my = e.clientY - rect.top
    panX.value = mx - (mx - panX.value) * (ns / scale.value)
    panY.value = my - (my - panY.value) * (ns / scale.value)
    scale.value = ns
  } else {
    panX.value -= e.deltaX
    panY.value -= e.deltaY
  }
}
function zoomIn()  { scale.value = Math.min(2, scale.value * 1.2) }
function zoomOut() { scale.value = Math.max(0.3, scale.value / 1.2) }

function fitView() {
  if (!nodes.value.length || !canvasRef.value) { scale.value = 1; panX.value = 40; panY.value = 40; return }
  const xs = nodes.value.map(n => n.x), ys = nodes.value.map(n => n.y)
  const minX = Math.min(...xs) - 80, minY = Math.min(...ys) - 80
  const maxX = Math.max(...xs) + 290, maxY = Math.max(...ys) + 220
  const rect = canvasRef.value.getBoundingClientRect()
  const s = Math.min(1.2, Math.max(0.3, Math.min(rect.width / (maxX - minX), rect.height / (maxY - minY))))
  scale.value = s
  panX.value = (rect.width - (maxX - minX) * s) / 2 - minX * s
  panY.value = (rect.height - (maxY - minY) * s) / 2 - minY * s
  nextTick(measurePorts)
}

// canvas-space cursor from a pointer event
function toCanvas(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  return {
    x: (e.clientX - rect.left - panX.value) / scale.value,
    y: (e.clientY - rect.top - panY.value) / scale.value,
  }
}

// ── canvas panning + node dragging ───────────────────────────────────────────
let mode = null // 'pan' | 'drag-node'
let dragNodeId = null
let dragOffset = { x: 0, y: 0 }
let panStart = { x: 0, y: 0 }

function onCanvasPointerDown(e) {
  // background click: deselect + start panning
  selectedNode.value = null
  selectedEdge.value = null
  mode = 'pan'
  panStart = { x: e.clientX - panX.value, y: e.clientY - panY.value }
}

function onNodePointerDown(e, node) {
  // don't start a drag from interactive controls (they stop propagation themselves)
  selectedNode.value = node.id
  selectedEdge.value = null
  mode = 'drag-node'
  dragNodeId = node.id
  const c = toCanvas(e)
  dragOffset = { x: c.x - node.x, y: c.y - node.y }
}

function onCanvasPointerMove(e) {
  if (connecting.value) {
    ghostCursor.value = toCanvas(e)
    return
  }
  if (mode === 'pan') {
    panX.value = e.clientX - panStart.x
    panY.value = e.clientY - panStart.y
  } else if (mode === 'drag-node' && dragNodeId) {
    const c = toCanvas(e)
    const n = nodes.value.find(n => n.id === dragNodeId)
    if (n) {
      patch({ nodes: nodes.value.map(x => x.id === dragNodeId ? { ...x, x: Math.round(c.x - dragOffset.x), y: Math.round(c.y - dragOffset.y) } : x) })
      measurePorts()
    }
  }
}

function onCanvasPointerUp() {
  mode = null
  dragNodeId = null
  if (connecting.value) cancelEdge()
}

// ── connecting edges ─────────────────────────────────────────────────────────
// connecting = { nodeId, route } | null   (nodeId '__start__' sets entry)
const connecting = ref(null)
const ghostCursor = ref({ x: 0, y: 0 })

function startEdge(src) {
  connecting.value = src
}
function onStartPointerDown() { /* allow pan from pill body; port handles connect */ }

function endEdge(targetId) {
  const src = connecting.value
  connecting.value = null
  if (!src || targetId === src.nodeId) return
  if (src.nodeId === '__start__') {
    patch({ entry: targetId })
  } else {
    // replace any existing edge with the same (from, route) so a router route
    // points at exactly one target, matching the validation rule
    const filtered = edges.value.filter(e => !(e.from === src.nodeId && (e.route || '') === (src.route || '')))
    const edge = { from: src.nodeId, to: targetId }
    if (src.route) edge.route = src.route
    patch({ edges: [...filtered, edge] })
  }
  nextTick(measurePorts)
}
function cancelEdge() { connecting.value = null }

// ── port geometry (measured from DOM, converted to canvas coords) ────────────
const portPos = reactive({})   // key -> {x, y}
const measureTick = ref(0)

function portKeyOut(nodeId, route) { return `out:${nodeId}|${route || ''}` }
function portKeyIn(nodeId) { return `in:${nodeId}` }

function measurePorts() {
  if (!innerRef.value) return
  const innerRect = innerRef.value.getBoundingClientRect()
  const s = scale.value
  const next = {}
  innerRef.value.querySelectorAll('[data-port]').forEach(el => {
    const r = el.getBoundingClientRect()
    next[el.dataset.port] = {
      x: (r.left + r.width / 2 - innerRect.left) / s,
      y: (r.top + r.height / 2 - innerRect.top) / s,
    }
  })
  Object.keys(portPos).forEach(k => delete portPos[k])
  Object.assign(portPos, next)
  measureTick.value++
}

// Recompute when the structure changes or the view transforms.
watch([nodes, edges, scale, panX, panY], () => nextTick(measurePorts), { deep: true })
onMounted(() => { nextTick(() => { measurePorts(); fitView() }) })

// bezier between two canvas points, horizontal tangents
function bezier(a, b) {
  if (!a || !b) return ''
  const dx = Math.max(40, Math.abs(b.x - a.x) * 0.5)
  return `M ${a.x},${a.y} C ${a.x + dx},${a.y} ${b.x - dx},${b.y} ${b.x},${b.y}`
}
function mid(a, b) { return { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 } }

const edgePaths = computed(() => {
  measureTick.value // reactive dependency
  return edges.value.map((e, index) => {
    const a = portPos[portKeyOut(e.from, e.route)] || portPos[portKeyOut(e.from, '')]
    const b = portPos[portKeyIn(e.to)]
    return {
      index,
      route: e.route || '',
      d: bezier(a, b),
      mid: a && b ? mid(a, b) : { x: 0, y: 0 },
      selected: selectedEdge.value === index,
    }
  }).filter(e => e.d)
})

const entryPath = computed(() => {
  measureTick.value
  if (!entry.value) return ''
  const a = portPos['out:__start__|']
  const b = portPos[portKeyIn(entry.value)]
  return bezier(a, b)
})

const ghostPath = computed(() => {
  measureTick.value
  if (!connecting.value) return ''
  const a = portPos[portKeyOut(connecting.value.nodeId, connecting.value.route)]
  return bezier(a, ghostCursor.value)
})

// START sentinel position: left of the entry node, else a default spot.
const startPos = computed(() => {
  const entryNode = nodes.value.find(n => n.id === entry.value)
  if (entryNode) return { x: entryNode.x - 150, y: entryNode.y + 4 }
  return { x: 40, y: 60 }
})
</script>

<style scoped>
.flow-canvas {
  position: relative;
  width: 100%;
  height: 520px;
  overflow: hidden;
  border-radius: 0.75rem;
  border: 1px solid rgba(120, 113, 108, 0.25);
  background-color: #121214;
  background-image: radial-gradient(circle, rgba(120, 113, 108, 0.13) 1px, transparent 1px);
  background-size: 22px 22px;
  cursor: grab;
  touch-action: none;
}
.flow-canvas:active { cursor: grabbing; }

.flow-canvas-inner {
  position: absolute;
  top: 0;
  left: 0;
  width: max-content;
  will-change: transform;
}

.flow-edges {
  position: absolute;
  top: 0;
  left: 0;
  pointer-events: none;
  overflow: visible;
}
.flow-edge-line { pointer-events: stroke; cursor: pointer; transition: stroke 0.12s ease; }
.flow-edge-line:hover { stroke: var(--color-arena-300); }
.flow-edge-label {
  fill: var(--color-atlantico-300);
  font-size: 9px;
  font-weight: 600;
  font-family: ui-monospace, monospace;
}

/* START sentinel */
.flow-start { position: absolute; }
.flow-start-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px 4px 8px;
  border-radius: 9999px;
  background: rgba(251, 113, 133, 0.12);
  border: 1px solid rgba(251, 113, 133, 0.4);
  color: var(--color-rose-300, #fda4af);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  white-space: nowrap;
  box-shadow: 0 2px 10px -4px rgba(251, 113, 133, 0.5);
}
.flow-start-port { right: -7px; top: 50%; transform: translateY(-50%); }
.flow-start-port:hover { transform: translateY(-50%) scale(1.25); }

/* toolbar */
.flow-toolbar {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 9px;
  width: 132px;
  background: rgba(26, 26, 29, 0.95);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(120, 113, 108, 0.15);
  border-radius: 12px;
}
.flow-tool-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 9px;
  border: 1px solid;
  border-radius: 8px;
  background: var(--color-piedra-800);
  color: var(--color-arena-200);
  transition: all 0.13s ease;
  cursor: pointer;
}
.flow-tool-btn:hover { transform: translateY(-1px); }

/* bottom bar */
.flow-bottom-bar {
  position: absolute;
  bottom: 12px;
  left: 12px;
  z-index: 21;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 10px;
  background: rgba(26, 26, 29, 0.95);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(120, 113, 108, 0.15);
  border-radius: 10px;
}
.flow-zoom-btn {
  height: 24px;
  min-width: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--color-arena-400);
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s;
}
.flow-zoom-btn:hover { background: rgba(120, 113, 108, 0.15); color: var(--color-arena-200); }
</style>
