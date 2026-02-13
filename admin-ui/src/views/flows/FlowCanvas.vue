<template>
  <div class="flow-canvas" ref="canvasRef" @wheel.prevent="onWheel" @pointerdown="onCanvasPointerDown">
    <!-- Empty state: pick root type -->
    <div v-if="!modelValue" class="absolute inset-0 z-10 flex items-center justify-center">
      <div class="flex flex-col items-center gap-5">
        <div class="text-center">
          <p class="text-sm font-medium text-arena-200">Choose a root type</p>
          <p class="text-[10px] text-arena-500 mt-1">This defines how the top-level steps execute</p>
        </div>
        <div class="flex gap-3">
          <button v-for="rt in rootTypes" :key="rt.type" @click="pickRoot(rt.type)"
            class="flex flex-col items-center gap-2 px-5 py-4 rounded-xl border transition-all hover:scale-105"
            :class="rt.cls">
            <div class="w-9 h-9 rounded-lg flex items-center justify-center" :class="rt.iconBg">
              <svg class="w-5 h-5" :class="rt.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="rt.icon" />
              </svg>
            </div>
            <span class="text-xs font-semibold" :class="rt.labelColor">{{ rt.label }}</span>
            <span class="text-[9px] text-arena-500 max-w-[90px] text-center leading-tight">{{ rt.desc }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Canvas content -->
    <template v-else>
      <div class="flow-canvas-inner" ref="innerRef" :style="canvasTransform">
        <FlowBlock
          :step="modelValue"
          :agents="agents"
          :is-root="true"
          @update="$emit('update:modelValue', $event)"
        />
      </div>

      <!-- Sidebar -->
      <Transition name="sidebar">
        <div v-if="sidebarOpen" class="flow-sidebar">
          <span class="text-[9px] text-arena-500 uppercase tracking-wider font-semibold px-1">Agents</span>
          <div
            class="flow-toolbar-item"
            :class="agentItem.cls"
            :title="agentItem.title"
            draggable="true"
            @dragstart="onToolbarDragStart($event, agentItem)"
            @dragend="onToolbarDragEnd"
          >
            <div class="w-5 h-5 rounded flex items-center justify-center flex-shrink-0" :class="agentItem.iconBg">
              <svg class="w-3 h-3" :class="agentItem.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="agentItem.icon" />
              </svg>
            </div>
            <span class="text-[10px] font-medium">{{ agentItem.label }}</span>
          </div>

          <div class="border-t border-piedra-700/40 my-0.5"></div>

          <span class="text-[9px] text-arena-500 uppercase tracking-wider font-semibold px-1">Flow</span>
          <div
            v-for="item in flowItems"
            :key="item.subtype"
            class="flow-toolbar-item"
            :class="item.cls"
            :title="item.title"
            draggable="true"
            @dragstart="onToolbarDragStart($event, item)"
            @dragend="onToolbarDragEnd"
          >
            <div class="w-5 h-5 rounded flex items-center justify-center flex-shrink-0" :class="item.iconBg">
              <svg class="w-3 h-3" :class="item.iconColor" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="item.icon" />
              </svg>
            </div>
            <span class="text-[10px] font-medium">{{ item.label }}</span>
          </div>
        </div>
      </Transition>

      <!-- Toggle sidebar button -->
      <button
        class="flow-sidebar-toggle"
        :style="{ left: sidebarOpen ? '166px' : '12px' }"
        @click="toggleSidebar"
        :title="sidebarOpen ? 'Hide toolbar' : 'Show toolbar'"
      >
        <svg class="w-3.5 h-3.5 text-arena-400 transition-transform" :class="sidebarOpen ? '' : 'rotate-180'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <!-- Bottom bar: zoom + center -->
      <div class="flow-bottom-bar">
        <button @click="zoomOut" class="flow-zoom-btn" title="Zoom out">−</button>
        <span class="text-[9px] text-arena-500 select-none w-8 text-center">{{ Math.round(scale * 100) }}%</span>
        <button @click="zoomIn" class="flow-zoom-btn" title="Zoom in">+</button>
        <div class="w-px h-3.5 bg-piedra-700/40 mx-0.5"></div>
        <button @click="centerView" class="flow-zoom-btn" title="Center view">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 8V5a2 2 0 012-2h3m8 0h3a2 2 0 012 2v3m0 8v3a2 2 0 01-2 2h-3m-8 0H5a2 2 0 01-2-2v-3" />
          </svg>
        </button>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, onMounted, watch } from 'vue'
import FlowBlock from './FlowBlock.vue'

const props = defineProps({
  modelValue: { type: Object, default: null },
  agents: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:modelValue'])

const sidebarOpen = ref(true)

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}

const rootTypes = [
  {
    type: 'sequential', label: 'Sequential', desc: 'Steps run one after another',
    cls: 'border-atlantico-500/30 hover:border-atlantico-500/60 bg-piedra-800/80',
    iconBg: 'bg-atlantico-500/15', iconColor: 'text-atlantico-400', labelColor: 'text-atlantico-300',
    icon: 'M13 5l7 7-7 7M5 5l7 7-7 7',
  },
  {
    type: 'parallel', label: 'Parallel', desc: 'Steps run at the same time',
    cls: 'border-sol-500/30 hover:border-sol-500/60 bg-piedra-800/80',
    iconBg: 'bg-sol-500/15', iconColor: 'text-sol-400', labelColor: 'text-sol-300',
    icon: 'M9 4v16M15 4v16',
  },
  {
    type: 'loop', label: 'Loop', desc: 'Steps repeat N times',
    cls: 'border-lava-500/30 hover:border-lava-500/60 bg-piedra-800/80',
    iconBg: 'bg-lava-500/15', iconColor: 'text-lava-400', labelColor: 'text-lava-300',
    icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15',
  },
]

function pickRoot(type) {
  const root = { type, steps: [] }
  if (type === 'loop') root.maxIterations = 3
  emit('update:modelValue', root)
  nextTick(() => setTimeout(fitView, 50))
}

const canvasRef = ref(null)
const innerRef = ref(null)
const panX = ref(0)
const panY = ref(0)
const scale = ref(1)
let isPanning = false
let panStartX = 0
let panStartY = 0

const canvasTransform = computed(() => ({
  transform: `translate(${panX.value}px, ${panY.value}px) scale(${scale.value})`,
  transformOrigin: '0 0',
}))

function centerView() {
  if (!canvasRef.value || !innerRef.value) return
  const canvas = canvasRef.value.getBoundingClientRect()
  const inner = innerRef.value

  const contentW = inner.scrollWidth * scale.value
  const contentH = inner.scrollHeight * scale.value

  panX.value = (canvas.width - contentW) / 2
  panY.value = (canvas.height - contentH) / 2
}

function fitView() {
  if (!canvasRef.value || !innerRef.value) return
  const canvas = canvasRef.value.getBoundingClientRect()
  const inner = innerRef.value

  const contentW = inner.scrollWidth
  const contentH = inner.scrollHeight

  if (contentW === 0 || contentH === 0) {
    scale.value = 1
    panX.value = canvas.width / 2
    panY.value = canvas.height / 2
    return
  }

  const padding = 60
  const scaleX = (canvas.width - padding * 2) / contentW
  const scaleY = (canvas.height - padding * 2) / contentH
  scale.value = Math.min(1.2, Math.max(0.3, Math.min(scaleX, scaleY)))

  const scaledW = contentW * scale.value
  const scaledH = contentH * scale.value
  panX.value = (canvas.width - scaledW) / 2
  panY.value = (canvas.height - scaledH) / 2
}

function onWheel(e) {
  if (e.ctrlKey || e.metaKey) {
    const delta = e.deltaY > 0 ? 0.9 : 1.1
    const newScale = Math.min(2, Math.max(0.3, scale.value * delta))
    const rect = canvasRef.value.getBoundingClientRect()
    const mx = e.clientX - rect.left
    const my = e.clientY - rect.top
    panX.value = mx - (mx - panX.value) * (newScale / scale.value)
    panY.value = my - (my - panY.value) * (newScale / scale.value)
    scale.value = newScale
  } else {
    panX.value -= e.deltaX
    panY.value -= e.deltaY
  }
}

function onCanvasPointerDown(e) {
  if (e.target !== canvasRef.value && !e.target.classList.contains('flow-canvas-inner')) return
  isPanning = true
  panStartX = e.clientX - panX.value
  panStartY = e.clientY - panY.value
  canvasRef.value.setPointerCapture(e.pointerId)
  canvasRef.value.addEventListener('pointermove', onCanvasPointerMove)
  canvasRef.value.addEventListener('pointerup', onCanvasPointerUp)
}

function onCanvasPointerMove(e) {
  if (!isPanning) return
  panX.value = e.clientX - panStartX
  panY.value = e.clientY - panStartY
}

function onCanvasPointerUp() {
  isPanning = false
  canvasRef.value?.removeEventListener('pointermove', onCanvasPointerMove)
  canvasRef.value?.removeEventListener('pointerup', onCanvasPointerUp)
}

function zoomIn() {
  scale.value = Math.min(2, scale.value * 1.2)
  nextTick(centerView)
}

function zoomOut() {
  scale.value = Math.max(0.3, scale.value / 1.2)
  nextTick(centerView)
}

const agentItem = {
  type: 'agent', label: 'Agent',
  title: 'An AI agent that processes input and produces output',
  cls: 'border-sol-500/30 hover:border-sol-500/60 bg-piedra-800',
  iconBg: 'bg-sol-500/15', iconColor: 'text-sol-400',
  icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z',
}

const flowItems = [
  {
    type: 'container', subtype: 'sequential', label: 'Sequential',
    title: 'Runs steps one after another, in order',
    cls: 'border-atlantico-500/30 hover:border-atlantico-500/60 bg-piedra-800',
    iconBg: 'bg-atlantico-500/15', iconColor: 'text-atlantico-400',
    icon: 'M13 5l7 7-7 7M5 5l7 7-7 7',
  },
  {
    type: 'container', subtype: 'parallel', label: 'Parallel',
    title: 'Runs all steps at the same time',
    cls: 'border-sol-500/30 hover:border-sol-500/60 bg-piedra-800',
    iconBg: 'bg-sol-500/15', iconColor: 'text-sol-400',
    icon: 'M9 4v16M15 4v16',
  },
  {
    type: 'container', subtype: 'loop', label: 'Loop',
    title: 'Repeats steps N times',
    cls: 'border-lava-500/30 hover:border-lava-500/60 bg-piedra-800',
    iconBg: 'bg-lava-500/15', iconColor: 'text-lava-400',
    icon: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15',
  },
]

function onToolbarDragStart(e, item) {
  e.dataTransfer.effectAllowed = 'copy'
  e.dataTransfer.setData('text/plain', JSON.stringify({
    type: item.type === 'agent' ? 'agent' : item.subtype,
    fromToolbar: true,
  }))
  document.body.classList.add('flow-dragging-from-toolbar')
}

function onToolbarDragEnd() {
  document.body.classList.remove('flow-dragging-from-toolbar')
}

onMounted(() => {
  nextTick(() => setTimeout(fitView, 50))
})

watch(() => props.modelValue, () => {
  nextTick(() => setTimeout(centerView, 30))
}, { deep: true })
</script>

<style scoped>
.flow-canvas {
  position: relative;
  width: 100%;
  height: 480px;
  overflow: hidden;
  border-radius: 0.75rem;
  border: 1px solid rgba(120, 113, 108, 0.25);
  background-color: #121214;
  background-image:
    radial-gradient(circle, rgba(120, 113, 108, 0.12) 1px, transparent 1px);
  background-size: 24px 24px;
  cursor: grab;
}

.flow-canvas:active {
  cursor: grabbing;
}

.flow-canvas-inner {
  position: absolute;
  top: 0;
  left: 0;
  width: max-content;
}

.flow-sidebar {
  position: absolute;
  top: 12px;
  left: 12px;
  bottom: 58px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px;
  background: rgba(26, 26, 29, 0.95);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(120, 113, 108, 0.15);
  border-radius: 12px;
  z-index: 20;
  width: 140px;
}

.flow-sidebar-toggle {
  position: absolute;
  top: 12px;
  z-index: 21;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgba(26, 26, 29, 0.95);
  backdrop-filter: blur(16px);
  border: 1px solid rgba(120, 113, 108, 0.15);
  transition: left 0.2s ease;
  cursor: pointer;
}

.flow-sidebar-toggle:hover {
  background: rgba(40, 40, 44, 0.95);
}

.flow-toolbar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border: 1px solid;
  border-radius: 8px;
  cursor: grab;
  transition: all 0.15s;
  color: var(--color-arena-200);
  user-select: none;
}

.flow-toolbar-item:active {
  cursor: grabbing;
  transform: scale(0.97);
}

.flow-bottom-bar {
  position: absolute;
  bottom: 12px;
  left: 12px;
  width: 140px;
  z-index: 21;
  display: flex;
  align-items: center;
  justify-content: center;
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

.flow-zoom-btn:hover {
  background: rgba(120, 113, 108, 0.15);
  color: var(--color-arena-200);
}

.sidebar-enter-active {
  transition: all 0.2s ease-out;
}
.sidebar-leave-active {
  transition: all 0.15s ease-in;
}
.sidebar-enter-from {
  opacity: 0;
  transform: translateX(-12px);
}
.sidebar-leave-to {
  opacity: 0;
  transform: translateX(-12px);
}
</style>
