<template>
  <div
    class="flow-canvas relative"
    :class="{ 'flow-canvas-fullscreen': fullscreen }"
    ref="canvasRef"
    @wheel.prevent="onWheel"
    @pointerdown="onCanvasPointerDown"
    @pointermove="onCanvasPointerMove"
    @pointerup="onCanvasPointerUp"
  >
    <!-- Toolbar -->
    <div class="flow-toolbar">
      <span class="text-[9px] text-arena-500 uppercase tracking-wider font-semibold px-1 select-none">Add node</span>
      <div v-for="(group, gi) in nodeGroups" :key="gi" class="flow-tool-group">
        <button v-for="t in group" :key="t.type" class="flow-tool-btn" :class="t.cls" @click="addNode(t.type)" :title="t.title">
          <span class="w-4 h-4 rounded flex items-center justify-center flex-shrink-0" :class="t.iconBg">
            <svg class="w-2.5 h-2.5" :class="t.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.7">
              <path stroke-linecap="round" stroke-linejoin="round" :d="t.icon" />
            </svg>
          </span>
          <span class="text-[10px] font-medium">{{ t.label }}</span>
        </button>
      </div>
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

        <!-- entry edge: START -> entry node (selectable + deletable) -->
        <g v-if="entryEdge.d">
          <!-- wide invisible hitbox so the thin line is easy to click -->
          <path :d="entryEdge.d" fill="none" stroke="transparent" stroke-width="18" class="flow-edge-hit" @pointerdown.stop="selectEdge('entry')" />
          <path
            :d="entryEdge.d"
            fill="none"
            :stroke="entryEdge.selected ? 'var(--color-sol-400)' : 'var(--color-arena-500)'"
            :stroke-width="entryEdge.selected ? 2.5 : 1.8"
            marker-end="url(#arrow)"
            class="flow-edge-line"
            @pointerdown.stop="selectEdge('entry')"
          />
          <g v-if="entryEdge.selected" :transform="`translate(${entryEdge.mid.x}, ${entryEdge.mid.y})`" class="flow-edge-delete" @pointerdown.stop="removeEdge('entry')">
            <circle r="8" fill="var(--color-sol-400)" />
            <path d="M-3,-3 L3,3 M3,-3 L-3,3" stroke="var(--color-piedra-950)" stroke-width="1.6" stroke-linecap="round" />
          </g>
        </g>

        <!-- real edges -->
        <g v-for="(e, i) in edgePaths" :key="'e' + i">
          <!-- wide invisible hitbox so the thin line is easy to click -->
          <path :d="e.d" fill="none" stroke="transparent" stroke-width="18" class="flow-edge-hit" @pointerdown.stop="selectEdge(e.index)" />
          <path
            :d="e.d" fill="none"
            :stroke="e.selected ? 'var(--color-sol-400)' : 'var(--color-arena-500)'"
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
          <g v-if="e.selected" :transform="`translate(${e.mid.x}, ${e.mid.y + (e.route ? 16 : 0)})`" class="flow-edge-delete" @pointerdown.stop="removeEdge(e.index)">
            <circle r="8" fill="var(--color-sol-400)" />
            <path d="M-3,-3 L3,3 M3,-3 L-3,3" stroke="var(--color-piedra-950)" stroke-width="1.6" stroke-linecap="round" />
          </g>
        </g>

        <!-- ghost edge while connecting -->
        <path v-if="ghostPath" :d="ghostPath" fill="none" stroke="var(--color-sol-400)" stroke-width="2" stroke-dasharray="4 4" opacity="0.85" />
      </svg>

      <!-- START sentinel -->
      <div class="flow-start" :style="{ left: startPos.x + 'px', top: startPos.y + 'px' }" @pointerdown.stop="onStartPointerDown" title="Drag to move. Drag from the dot to pick the entry node.">
        <div class="flow-start-card">
          <div class="flow-start-ic">
            <svg class="w-3.5 h-3.5 text-rose-400" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z" /></svg>
          </div>
          <span class="flow-start-label">Start</span>
        </div>
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
        <p class="text-[10px] text-rose-400/80 mt-1.5">Drag from the <span class="font-semibold">Start</span> dot to the node where execution begins.</p>
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
      <div class="w-px h-3.5 bg-piedra-700/40 mx-0.5"></div>
      <button @click="toggleFullscreen" class="flow-zoom-btn" :title="fullscreen ? 'Exit full screen (Esc)' : 'Full screen'">
        <svg v-if="!fullscreen" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 8V4h4m8 0h4v4m0 8v4h-4m-8 0H4v-4" /></svg>
        <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 4v4H5m10-4v4h4M9 20v-4H5m10 4v-4h4" /></svg>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
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
  // Preserve every graph field, not just the three big ones, so layout hints
  // like the Start box position survive a partial update.
  commit({ ...graph.value, entry: entry.value, nodes: nodes.value, edges: edges.value, ...partial })
}

// ── node toolbar ─────────────────────────────────────────────────────────────
// Grouped by role: execution (agent, subflow), flow control (router, join,
// parallel), data (expression, template). Groups are separated by whitespace
// in the toolbar, no subheadings. Add new node types to the matching group.
const nodeGroups = [
  // Execution: things that run work.
  [
    { type: 'agent', label: 'Agent', title: 'An AI agent that processes input and produces output',
      cls: 'border-sol-500/30 hover:border-sol-500/60', iconBg: 'bg-sol-500/15', iconColor: 'text-sol-400',
      icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z' },
    { type: 'subflow', label: 'Subflow', title: 'Embed another flow as a nested workflow',
      cls: 'border-rose-500/30 hover:border-rose-500/60', iconBg: 'bg-rose-500/15', iconColor: 'text-rose-400',
      icon: 'M9 4H5a1 1 0 00-1 1v4m0 6v4a1 1 0 001 1h4m6-16h4a1 1 0 011 1v4m0 6v4a1 1 0 01-1 1h-4M9 9h6v6H9z' },
  ],
  // Flow control: routing, joining, fan-out.
  [
    { type: 'router', label: 'Router', title: 'Evaluates CEL rules and routes to one of several branches',
      cls: 'border-atlantico-500/30 hover:border-atlantico-500/60', iconBg: 'bg-atlantico-500/15', iconColor: 'text-atlantico-400',
      icon: 'M3 12h4l3-9 4 18 3-9h4' },
    { type: 'join', label: 'Join', title: 'Fan-in barrier: waits for all incoming branches',
      cls: 'border-purple-500/30 hover:border-purple-500/60', iconBg: 'bg-purple-500/15', iconColor: 'text-purple-400',
      icon: 'M7 4v5a5 5 0 005 5 5 5 0 005-5V4M12 14v6' },
    { type: 'parallel', label: 'Foreach', title: 'Runs an agent once per item of a list input, concurrently',
      cls: 'border-lava-500/30 hover:border-lava-500/60', iconBg: 'bg-lava-500/15', iconColor: 'text-lava-400',
      icon: 'M4 6h16M4 12h16M4 18h16' },
  ],
  // Data: shape and transform values between steps.
  [
    { type: 'expression', label: 'Expression', title: 'Transform the input with a CEL expression over input and state',
      cls: 'border-indigo-500/30 hover:border-indigo-500/60', iconBg: 'bg-indigo-500/15', iconColor: 'text-indigo-400',
      icon: 'M8 9l3 3-3 3m5 0h3M4 4h16a1 1 0 011 1v14a1 1 0 01-1 1H4a1 1 0 01-1-1V5a1 1 0 011-1z' },
    { type: 'template', label: 'Template', title: 'Render text with {{ input }} and {{ state.key }} placeholders',
      cls: 'border-teal-500/30 hover:border-teal-500/60', iconBg: 'bg-teal-500/15', iconColor: 'text-teal-400',
      icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
    { type: 'code', label: 'Code', title: 'Run a Starlark script over input and state',
      cls: 'border-emerald-500/30 hover:border-emerald-500/60', iconBg: 'bg-emerald-500/15', iconColor: 'text-emerald-400',
      icon: 'M8 9l3 3-3 3m5 0h3M5 5h14a1 1 0 011 1v12a1 1 0 01-1 1H5a1 1 0 01-1-1V6a1 1 0 011-1z' },
  ],
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
  // Foreach is wider than the default so its concurrency controls read well.
  if (type === 'parallel') { node.agentId = ''; node.maxConcurrency = 0; node.w = 280 }
  if (type === 'subflow') node.flowId = ''
  // Text nodes are born clearly larger than selection nodes: the size itself
  // signals that a lot of text goes inside.
  if (type === 'expression') { node.expression = ''; node.outputKey = ''; node.w = 360; node.h = 260 }
  if (type === 'template') { node.template = ''; node.outputKey = ''; node.w = 360; node.h = 280 }
  if (type === 'code') { node.script = ''; node.outputKey = ''; node.timeoutMs = 0; node.maxOutputBytes = 0; node.w = 400; node.h = 320 }
  if (type === 'router') {
    // Three empty rows so the rotating placeholders show a full example ladder
    // of score conditions, and a wider card so the expressions fit.
    node.rules = [{ when: '', route: '' }, { when: '', route: '' }, { when: '', route: '' }]
    node.defaultRoute = 'default'
    node.w = 280
  }
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
  if (i === 'entry') {
    // The Start -> Entry arrow is not a stored edge; "deleting" it clears the
    // entry so the operator can wire a new one by dragging from Start.
    patch({ entry: '' })
    selectedEdge.value = null
    nextTick(measurePorts)
    return
  }
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
  } else if (mode === 'drag-start') {
    const c = toCanvas(e)
    patch({ startX: Math.round(c.x - dragOffset.x), startY: Math.round(c.y - dragOffset.y) })
    measurePorts()
  }
}

function onCanvasPointerUp(e) {
  if (connecting.value) {
    // Tolerant drop: connect to whatever node sits under the cursor, not just
    // its 12px input port. The SVG edge layer is pointer-events:none, so
    // elementFromPoint returns the node beneath it.
    const targetId = nodeIdAtPoint(e.clientX, e.clientY)
    if (targetId) endEdge(targetId)
    else cancelEdge()
  }
  mode = null
  dragNodeId = null
}

// nodeIdAtPoint returns the FlowNode id under the given viewport point, or
// null. It scans the full stack (elementsFromPoint) so an interactive edge
// line or label sitting above the node does not shadow the drop target.
function nodeIdAtPoint(clientX, clientY) {
  const stack = document.elementsFromPoint(clientX, clientY)
  for (const el of stack) {
    const nodeEl = el.closest && el.closest('[data-node-id]')
    if (nodeEl) return nodeEl.dataset.nodeId
  }
  return null
}

// ── connecting edges ─────────────────────────────────────────────────────────
// connecting = { nodeId, route } | null   (nodeId '__start__' sets entry)
const connecting = ref(null)
const ghostCursor = ref({ x: 0, y: 0 })

function startEdge(src) {
  connecting.value = src
}
// onStartPointerDown drags the Start box around the canvas. The port sitting on
// its edge stops propagation and starts a connection instead, so a drag from
// the body moves the box and a drag from the port wires the entry.
function onStartPointerDown(e) {
  mode = 'drag-start'
  const c = toCanvas(e)
  dragOffset = { x: c.x - startPos.value.x, y: c.y - startPos.value.y }
}

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

// ── full screen ──────────────────────────────────────────────────────────────
// CSS-based full screen (a fixed overlay), not the native Fullscreen API, so we
// control Escape precisely: Esc exits full screen and never reaches the dialog.
// The dialog itself is `persistent`, so Esc outside full screen does nothing —
// a flow is only dismissed via Cancel/Save.
const fullscreen = ref(false)
function toggleFullscreen() {
  fullscreen.value = !fullscreen.value
  nextTick(() => { measurePorts(); fitView() })
}
function onKeydown(e) {
  if (e.key === 'Escape' && fullscreen.value) {
    e.preventDefault()
    e.stopPropagation()
    fullscreen.value = false
    nextTick(() => { measurePorts(); fitView() })
  }
}
// Capture phase so we intercept Escape before it bubbles anywhere else.
onMounted(() => window.addEventListener('keydown', onKeydown, true))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown, true))

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

// entryEdge describes the Start -> Entry arrow: its bezier path, midpoint (for
// the delete button) and whether it is currently selected. Empty d when there
// is no entry yet.
const entryEdge = computed(() => {
  measureTick.value
  if (!entry.value) return { d: '' }
  const a = portPos['out:__start__|']
  const b = portPos[portKeyIn(entry.value)]
  return {
    d: bezier(a, b),
    mid: a && b ? mid(a, b) : { x: 0, y: 0 },
    selected: selectedEdge.value === 'entry',
  }
})

const ghostPath = computed(() => {
  measureTick.value
  if (!connecting.value) return ''
  const a = portPos[portKeyOut(connecting.value.nodeId, connecting.value.route)]
  return bezier(a, ghostCursor.value)
})

// START sentinel position: the operator's placement (startX/startY) if set,
// else pinned to the left of the entry node, else a default spot.
const startPos = computed(() => {
  const g = graph.value
  if (typeof g.startX === 'number' && typeof g.startY === 'number') {
    return { x: g.startX, y: g.startY }
  }
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

/* CSS full screen: a fixed overlay above the dialog (z ~50). */
.flow-canvas-fullscreen {
  position: fixed;
  inset: 0;
  width: 100vw;
  height: 100vh;
  z-index: 60;
  border-radius: 0;
}

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
.flow-edge-line { pointer-events: none; transition: stroke 0.12s ease; }
/* wide transparent path under each edge that captures clicks */
.flow-edge-hit { pointer-events: stroke; cursor: pointer; }
.flow-edge-hit:hover + .flow-edge-line { stroke: var(--color-arena-300); }
/* the SVG layer is pointer-events:none; the delete hitbox opts back in */
.flow-edge-delete { pointer-events: all; cursor: pointer; }
.flow-edge-label {
  fill: var(--color-atlantico-300);
  font-size: 9px;
  font-weight: 600;
  font-family: ui-monospace, monospace;
}

/* START sentinel */
.flow-start { position: absolute; cursor: grab; }
.flow-start:active { cursor: grabbing; }
.flow-start-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 12px;
  background: rgba(26, 26, 29, 0.95);
  border: 1px solid rgba(251, 113, 133, 0.35);
  box-shadow: 0 8px 24px -10px rgba(0, 0, 0, 0.6);
  transition: border-color 0.12s ease;
}
.flow-start:hover .flow-start-card { border-color: rgba(251, 113, 133, 0.6); }
.flow-start-ic {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: rgba(251, 113, 133, 0.15);
}
.flow-start-label {
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-rose-400, #fb7185);
}
.flow-start-port {
  /* .flow-port lives in FlowNode.vue's scoped styles, so the Start box's own
     port needs the circle defined here too. */
  position: absolute;
  right: -7px;
  top: 50%;
  transform: translateY(-50%);
  width: 12px;
  height: 12px;
  border-radius: 9999px;
  background: var(--color-piedra-700);
  border: 2px solid var(--color-rose-400, #fb7185);
  cursor: crosshair;
  transition: all 0.12s ease;
  z-index: 6;
}
.flow-start-port:hover {
  transform: translateY(-50%) scale(1.3);
  background: var(--color-rose-400, #fb7185);
}

/* toolbar */
.flow-toolbar {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding: 9px;
  width: 132px;
  background: rgba(26, 26, 29, 0.95);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(120, 113, 108, 0.15);
  border-radius: 12px;
}
/* a group of related node buttons; the toolbar's larger gap separates groups,
   this smaller gap keeps buttons within a group tight */
.flow-tool-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
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
