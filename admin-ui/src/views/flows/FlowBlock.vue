<template>
  <!-- AGENT NODE -->
  <div v-if="step.type === 'agent'" class="flow-agent group relative">
    <div class="bg-piedra-800 border border-sol-500/30 rounded-xl
                hover:border-sol-500/60 transition-all min-w-[130px] shadow-sm hover:shadow-md">
      <div class="flex items-center gap-2 px-3 py-2.5 cursor-grab active:cursor-grabbing">
        <div class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 bg-sol-500/15">
          <svg class="w-3.5 h-3.5 text-sol-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
      </div>
    </div>

    <Transition name="dropdown">
      <div v-if="pickerOpen" class="absolute z-50 left-0 top-full mt-1 w-52 bg-piedra-800 border border-piedra-700/60 rounded-xl shadow-2xl overflow-hidden">
        <div v-if="agents.length" class="py-1 max-h-48 overflow-y-auto">
          <button
            v-for="a in agents" :key="a.id"
            @click.stop="pickAgent(a.id)"
            class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
            :class="a.id === step.agentId ? 'bg-sol-500/10' : 'hover:bg-piedra-700/60'"
          >
            <div class="w-5 h-5 rounded-md flex items-center justify-center flex-shrink-0"
              :class="a.id === step.agentId ? 'bg-sol-500/20' : 'bg-piedra-700'">
              <svg class="w-3 h-3" :class="a.id === step.agentId ? 'text-sol-400' : 'text-arena-500'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <div class="text-xs font-medium truncate" :class="a.id === step.agentId ? 'text-sol-300' : 'text-arena-100'">{{ a.name || a.id }}</div>
              <div v-if="a.description" class="text-[9px] text-arena-500 truncate">{{ a.description }}</div>
            </div>
            <svg v-if="a.id === step.agentId" class="w-3.5 h-3.5 text-sol-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </button>
        </div>
        <div v-else class="px-3 py-4 text-[10px] text-arena-500 italic text-center">No agents available</div>
      </div>
    </Transition>
  </div>

  <!-- CONTAINER NODE -->
  <div v-else class="flow-container" :class="containerClass">
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
          class="text-[9px] px-1 py-0.5 rounded hover:bg-white/10 transition-colors text-lava-400/60 hover:text-lava-400 select-none">
          ✕
        </button>
      </div>
    </div>

    <div
      class="flow-drop-zone p-2 min-h-[52px] transition-colors"
      :class="[dropHighlight ? 'flow-drop-active' : '']"
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
          <div class="flow-item" :class="itemClass(index)">
            <FlowBlock
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
        class="flex items-center justify-center py-6 border-2 border-dashed rounded-lg text-[10px] italic select-none"
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
  step: { type: Object, required: true },
  agents: { type: Array, default: () => [] },
  isRoot: { type: Boolean, default: false },
  parentType: { type: String, default: '' },
})

const emit = defineEmits(['update', 'remove'])

const dropHighlight = ref(false)
const pickerOpen = ref(false)

function toggleAgentPicker() {
  pickerOpen.value = !pickerOpen.value
}

function pickAgent(id) {
  pickerOpen.value = false
  emit('update', { ...props.step, agentId: id })
}

function toggleResponse() {
  emit('update', { ...props.step, responseAgent: !props.step.responseAgent })
}

function onClickOutside(e) {
  if (pickerOpen.value) pickerOpen.value = false
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

let keyCounter = 0
function ensureKeys(steps) {
  if (!steps) return
  for (const s of steps) {
    if (!s.__key) s.__key = `k${++keyCounter}`
    if (s.steps) ensureKeys(s.steps)
  }
}

watch(() => props.step.steps, (steps) => ensureKeys(steps), { immediate: true, deep: true })

const agentName = computed(() => {
  if (props.step.type !== 'agent') return ''
  const a = props.agents.find(a => a.id === props.step.agentId)
  return a?.name || props.step.agentId || 'Select agent...'
})

const typeLabel = computed(() => ({
  sequential: 'Sequential',
  parallel: 'Parallel',
  loop: 'Loop',
})[props.step.type] || props.step.type)

const containerColors = {
  sequential: {
    border: 'border-atlantico-500/25',
    header: 'bg-atlantico-500/8 rounded-t-xl',
    label: 'text-atlantico-400',
    badge: 'bg-atlantico-500/20 text-atlantico-300',
    empty: 'text-atlantico-400/30 border-atlantico-500/15',
  },
  parallel: {
    border: 'border-sol-500/25',
    header: 'bg-sol-500/8 rounded-t-xl',
    label: 'text-sol-400',
    badge: 'bg-sol-500/20 text-sol-300',
    empty: 'text-sol-400/30 border-sol-500/15',
  },
  loop: {
    border: 'border-lava-500/25',
    header: 'bg-lava-500/8 rounded-t-xl',
    label: 'text-lava-400',
    badge: 'bg-lava-500/20 text-lava-300',
    empty: 'text-lava-400/30 border-lava-500/15',
  },
}

const colors = computed(() => containerColors[props.step.type] || containerColors.sequential)
const containerClass = computed(() => `rounded-xl border ${colors.value.border} bg-piedra-900/60`)
const headerClass = computed(() => colors.value.header)
const labelClass = computed(() => colors.value.label)
const badgeClass = computed(() => colors.value.badge)
const emptyClass = computed(() => colors.value.empty)

const isHorizontal = computed(() => props.step.type === 'sequential' || props.step.type === 'loop')

const dragAreaClass = computed(() =>
  isHorizontal.value
    ? 'flex flex-row flex-nowrap items-center gap-0'
    : 'flex flex-col gap-1.5'
)

function itemClass(index) {
  if (!isHorizontal.value) return ''
  const isLast = index === (props.step.steps?.length || 0) - 1
  return isLast ? '' : 'seq-item'
}

function addStep(newStep) {
  newStep.__key = `k${++keyCounter}`
  const steps = [...(props.step.steps || []), newStep]
  emit('update', { ...props.step, steps })
}

function updateChild(index, newChild) {
  const steps = [...props.step.steps]
  steps[index] = newChild
  emit('update', { ...props.step, steps })
}

function removeChild(index) {
  const steps = props.step.steps.filter((_, i) => i !== index)
  emit('update', { ...props.step, steps })
}

function onDragChange() {
  emit('update', { ...props.step, steps: [...props.step.steps] })
}

const typeOrder = ['sequential', 'parallel', 'loop']

function cycleType() {
  const idx = typeOrder.indexOf(props.step.type)
  const nextType = typeOrder[(idx + 1) % typeOrder.length]
  const updated = { ...props.step, type: nextType }
  if (nextType === 'loop' && !updated.maxIterations) {
    updated.maxIterations = 3
  }
  emit('update', updated)
}

function editIterations() {
  const val = prompt('Max iterations (0 = infinite):', props.step.maxIterations || 0)
  if (val !== null) {
    emit('update', { ...props.step, maxIterations: parseInt(val) || 0 })
  }
}

function onDragOver(e) {
  try {
    const raw = e.dataTransfer.types.includes('text/plain')
    if (raw) {
      e.dataTransfer.dropEffect = 'copy'
      dropHighlight.value = true
    }
  } catch { /* */ }
}

function onDragLeave(e) {
  if (e.currentTarget.contains(e.relatedTarget)) return
  dropHighlight.value = false
}

function onDrop(e) {
  dropHighlight.value = false
  try {
    const raw = e.dataTransfer.getData('text/plain')
    if (!raw) return
    const data = JSON.parse(raw)
    if (!data.fromToolbar) return

    if (data.type === 'agent') {
      addStep({ type: 'agent', agentId: '' })
    } else {
      const newContainer = { type: data.type, steps: [] }
      if (data.type === 'loop') newContainer.maxIterations = 3
      addStep(newContainer)
    }
  } catch { /* */ }
}
</script>

<style scoped>
.flow-ghost {
  opacity: 0.25;
  border-radius: 0.75rem;
}

.flow-drag {
  opacity: 0.95;
  transform: rotate(1deg);
}

.flow-agent,
.flow-container {
  position: relative;
}

.seq-item {
  display: flex;
  align-items: center;
}

.seq-item::after {
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

.flow-drop-active {
  background: rgba(8, 145, 178, 0.06);
  border-radius: 0.5rem;
  outline: 2px dashed rgba(8, 145, 178, 0.3);
  outline-offset: -2px;
}

.dropdown-enter-active {
  transition: all 0.15s ease-out;
}
.dropdown-leave-active {
  transition: all 0.1s ease-in;
}
.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-4px) scale(0.97);
}
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.97);
}
</style>
