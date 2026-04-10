<template>
  <!-- AGENT NODE -->
  <div v-if="step.type === 'agent'" class="flow-agent group relative">
    <div class="bg-piedra-800 border border-sol-500/30 rounded-xl
                hover:border-sol-500/60 transition-all min-w-[130px] shadow-sm hover:shadow-md">
      <div class="flex items-center gap-2 px-3 py-2.5 cursor-grab active:cursor-grabbing">
        <div class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0"
             :class="agentNode?._isFlow ? 'bg-lava-500/15' : 'bg-sol-500/15'">
          <svg v-if="agentNode?._isFlow" class="w-3.5 h-3.5 text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h7" />
          </svg>
          <svg v-else class="w-3.5 h-3.5 text-sol-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </div>
        <button
          @click.stop="toggleAgentPicker"
          @mousedown.stop
          class="flex-1 flex items-center gap-1 text-xs font-medium outline-none cursor-pointer truncate min-w-0"
          :class="step.agentId ? 'text-arena-100' : 'text-arena-500 italic'"
        >
          <span class="truncate">{{ agentName }}</span>
          <svg class="w-3 h-3 flex-shrink-0 text-arena-500" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
          </svg>
        </button>
        <button v-if="!isRoot" @click.stop="$emit('remove')"
          class="p-0.5 rounded opacity-0 group-hover:opacity-100 hover:bg-lava-500/20 transition-all flex-shrink-0"
          title="Remove">
          <svg class="w-3 h-3 text-arena-500 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="flex items-center gap-1.5 px-3 pb-2 pt-0.5 border-t border-piedra-700/40" @mousedown.stop>
        <button @click.stop="toggleResponse"
          class="p-1 rounded-md transition-all select-none"
          :class="step.responseAgent
            ? 'bg-green-500/15 text-green-400 hover:bg-green-500/25'
            : 'text-arena-600 hover:text-arena-400 hover:bg-piedra-700/60'"
          :title="step.responseAgent ? 'This agent emits the final flow response' : 'Include this agent\'s output in the flow response'">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.076-4.076a1.526 1.526 0 0 1 1.037-.443 48.2 48.2 0 0 0 5.887-.512c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.4 48.4 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" />
          </svg>
        </button>
        <button v-if="agentNode?._isFlow" @click.stop="toggleInheritResponseAgents"
          class="p-1 rounded-md transition-all select-none ml-auto"
          :class="step.inheritResponseAgents !== false
            ? 'bg-lava-500/15 text-lava-400 hover:bg-lava-500/25'
            : 'text-arena-600 hover:text-arena-400 hover:bg-piedra-700/60'"
          :title="step.inheritResponseAgents !== false ? 'Inheriting sub-flow response agents' : 'Ignoring sub-flow response agents'">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z" />
          </svg>
        </button>
      </div>
    </div>

    <Transition name="dropdown">
      <div v-if="pickerOpen" class="absolute z-50 left-0 top-full mt-1 w-52 bg-piedra-800 border border-piedra-700/60 rounded-xl shadow-2xl overflow-hidden">
        <div v-if="agents.length" class="py-1 max-h-48 overflow-y-auto">
          <button
            v-for="a in agents" :key="a.id"
            @click.stop="pickAgent(a.id)"
            class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
            :class="a.id === step.agentId ? (a._isFlow ? 'bg-lava-500/10' : 'bg-sol-500/10') : 'hover:bg-piedra-700/60'"
          >
            <div class="w-5 h-5 rounded-md flex items-center justify-center flex-shrink-0"
              :class="a.id === step.agentId ? (a._isFlow ? 'bg-lava-500/20' : 'bg-sol-500/20') : 'bg-piedra-700'">
              <svg v-if="a._isFlow" class="w-3 h-3" :class="a.id === step.agentId ? 'text-lava-400' : 'text-arena-500'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h7" />
              </svg>
              <svg v-else class="w-3 h-3" :class="a.id === step.agentId ? 'text-sol-400' : 'text-arena-500'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <div class="text-xs font-medium truncate" :class="a.id === step.agentId ? 'text-sol-300' : 'text-arena-100'">{{ a.name || a.id }}</div>
              <div v-if="a.description" class="text-[9px] text-arena-500 truncate">{{ a.description }}</div>
            </div>
            <svg v-if="a.id === step.agentId" class="w-3.5 h-3.5 flex-shrink-0" :class="a._isFlow ? 'text-lava-400' : 'text-sol-400'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
        </div>
        <div v-else class="px-3 py-4 text-[10px] text-arena-500 italic text-center">No items available</div>
      </div>
    </Transition>
  </div>

  <!-- CONTAINER NODE -->
  <div
    v-else
    class="flow-container"
    :class="[containerClass, dropHighlight ? dropActiveClass : '']"
  >
    <div class="flow-container-header flex items-center gap-2 px-3 py-1.5 cursor-grab active:cursor-grabbing" :class="headerClass">
      <span class="text-[10px] font-bold uppercase tracking-wider select-none" :class="labelClass">{{ typeLabel }}</span>

      <template v-if="step.type === 'loop'">
        <button @click.stop="editIterations" @mousedown.stop
          class="text-[9px] px-1.5 py-0.5 rounded font-semibold hover:brightness-125 transition-all" :class="badgeClass">
          ×{{ step.maxIterations || '∞' }}
        </button>
      </template>

      <div class="flex-1" />

      <div class="flex items-center gap-0.5" @mousedown.stop>
        <button @click.stop="cycleType"
          class="text-[9px] px-1.5 py-0.5 rounded hover:bg-white/10 transition-colors select-none" :class="labelClass" title="Cycle type">
          ↻
        </button>
        <button v-if="!isRoot" @click.stop="$emit('remove')" @mousedown.stop
          class="text-[11px] px-1.5 py-0.5 rounded hover:bg-white/10 transition-colors text-lava-400/60 hover:text-lava-400 select-none">
          ✕
        </button>
      </div>
    </div>

    <div
      ref="dropZoneRef"
      class="flow-drop-zone p-4 min-h-[80px] transition-colors duration-200 rounded-b-xl"
      :class="dropHighlight ? 'bg-white/5' : ''"
      @dragenter.stop="onDragEnter"
      @dragover.prevent.stop="onDragOver"
      @dragleave.stop="onDragLeave"
      @drop.prevent.stop="onDrop"
    >
      <draggable
        :list="step.steps"
        :group="{ name: 'flow-steps' }"
        item-key="__key"
        :animation="200"
        ghost-class="flow-ghost"
        drag-class="flow-drag"
        :class="dragAreaClass"
        @change="onDragChange"
      >
        <template #item="{ element, index }">
          <div class="flow-item-wrap" :class="isHorizontal ? 'flow-item-wrap-h' : 'flow-item-wrap-v'">
            <div v-if="element.__placeholder" class="flow-ghost-wrap">
              <FlowBlock
                :step="ghostStep(element.__placeholderType)"
                :agents="agents"
                :is-root="false"
                :parent-type="step.type"
              />
            </div>
            <FlowBlock
              v-else
              :step="element"
              :agents="agents"
              :is-root="false"
              :parent-type="step.type"
              @update="updateChild(index, $event)"
              @remove="removeChild(index)"
            />
          </div>
        </template>
      </draggable>

      <div v-if="!step.steps?.length"
        class="flex items-center justify-center py-8 border-2 border-dashed rounded-lg text-[10px] italic select-none"
        :class="emptyClass">
        Drop here
      </div>
    </div>

    <div v-if="step.type === 'loop' && step.steps?.length" class="flex items-center gap-1.5 px-3 pb-1.5">
      <svg class="w-3 h-3 text-lava-400/50" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      <span class="text-[9px] text-lava-400/50 italic select-none">repeats {{ step.maxIterations ? `${step.maxIterations}×` : '∞' }}</span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch, onMounted, onBeforeUnmount } from 'vue'
import draggable from 'vuedraggable'

const props = defineProps({
  step:       { type: Object,  required: true },
  agents:     { type: Array,   default: () => [] },
  isRoot:     { type: Boolean, default: false },
  parentType: { type: String,  default: '' },
})

const emit = defineEmits(['update', 'remove'])

// ── helpers ─────────────────────────────────────────────────────────────────
function emitSteps(steps) {
  emit('update', { ...props.step, steps })
}

function realSteps() {
  return (props.step.steps || []).filter(s => !s.__placeholder)
}

function isToolbarDrag() {
  return 'toolbarDragType' in document.body.dataset
}

function ghostStep(type) {
  return type === 'agent'
    ? { type: 'agent', agentId: '' }
    : { type, steps: [] }
}

// ── toolbar drop (native HTML5 drag) ────────────────────────────────────────
const dropZoneRef      = ref(null)
const dropHighlight    = ref(false)
const placeholderIndex = ref(null)
const DEAD_ZONE        = 32

function makePlaceholder() {
  return {
    __key: '__toolbar_placeholder__',
    __placeholder: true,
    __placeholderType: document.body.dataset.toolbarDragType || 'agent',
  }
}

function computeDropIndex(e) {
  if (!dropZoneRef.value) return realSteps().length

  const directWraps = [...dropZoneRef.value.querySelectorAll('.flow-item-wrap')]
    .filter(el => el.closest('.flow-drop-zone') === dropZoneRef.value && !el.querySelector('.flow-ghost-wrap'))
  if (!directWraps.length) return 0

  const horizontal = isHorizontal.value
  const cursor = horizontal ? e.clientX : e.clientY
  for (const [i, wrap] of directWraps.entries()) {
    const rect = wrap.getBoundingClientRect()
    const mid  = horizontal ? rect.left + rect.width * 0.5 : rect.top + rect.height * 0.5
    if (cursor < mid) return i
  }
  return directWraps.length
}

function cursorInsidePlaceholder(e) {
  if (placeholderIndex.value === null || !dropZoneRef.value) return false
  const wrap = dropZoneRef.value.querySelector('.flow-ghost-wrap')?.closest('.flow-item-wrap')
  if (!wrap) return false
  const rect = wrap.getBoundingClientRect()
  const horizontal = isHorizontal.value
  const cursor = horizontal ? e.clientX : e.clientY
  const start  = horizontal ? rect.left  : rect.top
  const end    = horizontal ? rect.right : rect.bottom
  return cursor >= start - DEAD_ZONE && cursor <= end + DEAD_ZONE
}

function movePlaceholder(e) {
  const index = computeDropIndex(e)
  if (placeholderIndex.value === index) return
  if (cursorInsidePlaceholder(e)) return
  const steps = realSteps()
  steps.splice(index, 0, makePlaceholder())
  placeholderIndex.value = index
  emitSteps(steps)
}

function removePlaceholder() {
  if (placeholderIndex.value === null) return
  placeholderIndex.value = null
  emitSteps(realSteps())
}

function onDragEnter(e) {
  if (!isToolbarDrag()) return
  dropHighlight.value = true
  movePlaceholder(e)
}

function onDragOver(e) {
  if (!isToolbarDrag()) return
  e.dataTransfer.dropEffect = 'copy'
  movePlaceholder(e)
}

function onDragLeave(e) {
  if (!isToolbarDrag()) return
  if (e.relatedTarget && dropZoneRef.value?.contains(e.relatedTarget)) return
  dropHighlight.value = false
  removePlaceholder()
}

function onDrop(e) {
  dropHighlight.value = false
  const insertAt = placeholderIndex.value ?? realSteps().length
  placeholderIndex.value = null

  try {
    const data = JSON.parse(e.dataTransfer.getData('text/plain') || 'null')
    if (!data?.fromToolbar) return

    const newStep = data.type === 'agent'
      ? { type: 'agent', agentId: '' }
      : { type: data.type, steps: [], ...(data.type === 'loop' ? { maxIterations: 3 } : {}) }

    addStepAt(newStep, insertAt)
  } catch { /* invalid payload — ignore */ }
}

// ── step management (shared by both drag systems) ───────────────────────────
let keyCounter = 0

function ensureKeys(steps) {
  steps?.forEach(s => {
    if (s.__placeholder) return
    if (!s.__key) s.__key = `k${++keyCounter}`
    if (s.steps) ensureKeys(s.steps)
  })
}

watch(() => props.step.steps, ensureKeys, { immediate: true, deep: true })

function addStepAt(newStep, index) {
  newStep.__key = `k${++keyCounter}`
  const steps = realSteps()
  steps.splice(index, 0, newStep)
  emitSteps(steps)
}

function updateChild(index, newChild) {
  const steps = [...props.step.steps]
  steps[index] = newChild
  emitSteps(steps)
}

function removeChild(index) {
  emitSteps(props.step.steps.filter((s, i) => i !== index && !s.__placeholder))
}

function onDragChange() {
  emitSteps([...realSteps()])
}

// ── agent picker ─────────────────────────────────────────────────────────────
const pickerOpen = ref(false)

function toggleAgentPicker() { pickerOpen.value = !pickerOpen.value }
function pickAgent(id) { 
  pickerOpen.value = false
  const picked = props.agents.find(a => a.id === id)
  emit('update', { 
    ...props.step, 
    agentId: id, 
    ...(picked?._isFlow ? { inheritResponseAgents: true } : {}) 
  }) 
}
function toggleResponse()    { emit('update', { ...props.step, responseAgent: !props.step.responseAgent }) }
function toggleInheritResponseAgents() { emit('update', { ...props.step, inheritResponseAgents: props.step.inheritResponseAgents === false }) }

function onClickOutside() { if (pickerOpen.value) pickerOpen.value = false }
onMounted(()        => document.addEventListener('click', onClickOutside))
onBeforeUnmount(()  => document.removeEventListener('click', onClickOutside))

// ── container appearance ─────────────────────────────────────────────────────
const COLORS = {
  sequential: {
    border:     'border-atlantico-500/25',
    dropActive: 'border-atlantico-500/70 shadow-[0_0_0_2px_rgba(8,145,178,0.15)]',
    header:     'bg-atlantico-500/8 rounded-t-xl',
    label:      'text-atlantico-400',
    badge:      'bg-atlantico-500/20 text-atlantico-300',
    empty:      'text-atlantico-400/30 border-atlantico-500/15',
  },
  parallel: {
    border:     'border-sol-500/25',
    dropActive: 'border-sol-500/70 shadow-[0_0_0_2px_rgba(234,179,8,0.15)]',
    header:     'bg-sol-500/8 rounded-t-xl',
    label:      'text-sol-400',
    badge:      'bg-sol-500/20 text-sol-300',
    empty:      'text-sol-400/30 border-sol-500/15',
  },
  loop: {
    border:     'border-lava-500/25',
    dropActive: 'border-lava-500/70 shadow-[0_0_0_2px_rgba(239,68,68,0.15)]',
    header:     'bg-lava-500/8 rounded-t-xl',
    label:      'text-lava-400',
    badge:      'bg-lava-500/20 text-lava-300',
    empty:      'text-lava-400/30 border-lava-500/15',
  },
}

const agentNode = computed(() => props.agents.find(a => a.id === props.step.agentId))
const agentName = computed(() => {
  const a = agentNode.value
  return a?.name || props.step.agentId || 'Select item...'
})

const typeLabel       = computed(() => ({ sequential: 'Sequential', parallel: 'Parallel', loop: 'Loop' })[props.step.type] || props.step.type)
const colors          = computed(() => COLORS[props.step.type] || COLORS.sequential)
const containerClass  = computed(() => `rounded-xl border transition-all duration-150 bg-piedra-900/60 ${colors.value.border}`)
const dropActiveClass = computed(() => colors.value.dropActive)
const headerClass     = computed(() => colors.value.header)
const labelClass      = computed(() => colors.value.label)
const badgeClass      = computed(() => colors.value.badge)
const emptyClass      = computed(() => colors.value.empty)
const isHorizontal    = computed(() => props.step.type === 'sequential' || props.step.type === 'loop')
const dragAreaClass   = computed(() => isHorizontal.value ? 'flex flex-row flex-nowrap items-center gap-0' : 'flex flex-col gap-2.5')

// ── container controls ───────────────────────────────────────────────────────
const TYPE_ORDER = ['sequential', 'parallel', 'loop']

function cycleType() {
  const next = TYPE_ORDER[(TYPE_ORDER.indexOf(props.step.type) + 1) % TYPE_ORDER.length]
  emit('update', { ...props.step, type: next, ...(next === 'loop' && !props.step.maxIterations ? { maxIterations: 3 } : {}) })
}

function editIterations() {
  const val = prompt('Max iterations (0 = infinite):', props.step.maxIterations || 0)
  if (val !== null) emit('update', { ...props.step, maxIterations: parseInt(val) || 0 })
}
</script>

<style scoped>
.flow-ghost { 
  opacity: 0.4; 
  border-radius: 0.75rem; 
  outline: 2px dashed rgba(250,204,21,0.5); 
  outline-offset: 2px; 
}
.flow-drag  { 
  opacity: 0.95; 
  transform: scale(1.02) rotate(1deg); 
  cursor: grabbing !important; 
  box-shadow: 0 10px 25px -5px rgba(0,0,0,0.5); 
}

.flow-agent,
.flow-container { position: relative; }

.flow-item-wrap   { position: relative; flex-shrink: 0; }
.flow-item-wrap-h { display: flex; align-items: center; }
.flow-item-wrap-v { display: flex; flex-direction: column; }

.flow-item-wrap-h:not(:last-child)::after {
  content: '›';
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  flex-shrink: 0;
  font-size: 18px;
  line-height: 1;
  color: var(--color-arena-600);
  opacity: 0.7;
}

.flow-ghost-wrap {
  pointer-events: none;
  opacity: 0.25;
  border-radius: 0.75rem;
}

.dropdown-enter-active { transition: all 0.15s ease-out; }
.dropdown-leave-active { transition: all 0.1s ease-in; }
.dropdown-enter-from,
.dropdown-leave-to     { opacity: 0; transform: translateY(-4px) scale(0.97); }
</style>
