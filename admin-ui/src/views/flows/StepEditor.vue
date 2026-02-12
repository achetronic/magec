<template>
  <div class="space-y-2">
    <div class="flex items-center gap-2">
      <select
        :value="step.type"
        @change="changeType($event.target.value)"
        class="bg-piedra-800 border border-piedra-700 rounded-lg px-2 py-1.5 text-xs focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
      >
        <option value="agent">Agent</option>
        <option value="sequential">Sequential</option>
        <option value="parallel">Parallel</option>
        <option value="loop">Loop</option>
      </select>

      <select
        v-if="step.type === 'agent'"
        :value="step.agentId || ''"
        @change="updateField('agentId', $event.target.value)"
        class="flex-1 bg-piedra-800 border border-piedra-700 rounded-lg px-2 py-1.5 text-xs focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
      >
        <option value="" disabled>Select agent</option>
        <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name || a.id }}</option>
      </select>

      <div v-if="step.type === 'loop'" class="flex items-center gap-1.5">
        <label class="text-[10px] text-arena-500 whitespace-nowrap">Max iter</label>
        <input
          type="number"
          min="0"
          :value="step.maxIterations || 0"
          @input="updateField('maxIterations', parseInt($event.target.value) || 0)"
          class="w-16 bg-piedra-800 border border-piedra-700 rounded-lg px-2 py-1.5 text-xs focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
        />
        <span class="text-[10px] text-arena-600">0 = infinite</span>
      </div>

      <button
        v-if="!isRoot"
        @click="$emit('remove')"
        class="p-1 hover:bg-piedra-700 rounded-lg flex-shrink-0"
        title="Remove step"
      >
        <svg class="w-3.5 h-3.5 text-arena-400 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div v-if="isContainer" class="ml-4 pl-3 border-l-2 border-piedra-700/40 space-y-2">
      <div v-for="(child, i) in step.steps" :key="i">
        <StepEditor
          :step="child"
          :agents="agents"
          :is-root="false"
          @update="updateChild(i, $event)"
          @remove="removeChild(i)"
        />
      </div>
      <button
        @click="addChild"
        class="text-[10px] text-sol-400 hover:text-sol-300 flex items-center gap-1 py-1"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add step
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  step: { type: Object, required: true },
  agents: { type: Array, default: () => [] },
  isRoot: { type: Boolean, default: false },
})

const emit = defineEmits(['update', 'remove'])

const isContainer = computed(() =>
  ['sequential', 'parallel', 'loop'].includes(props.step.type)
)

function changeType(newType) {
  if (newType === 'agent') {
    emit('update', { type: 'agent', agentId: '' })
  } else {
    const updated = { type: newType, steps: props.step.steps || [] }
    if (newType === 'loop') {
      updated.maxIterations = props.step.maxIterations || 0
    }
    emit('update', updated)
  }
}

function updateField(field, value) {
  emit('update', { ...props.step, [field]: value })
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

function addChild() {
  const steps = [...(props.step.steps || []), { type: 'agent', agentId: '' }]
  emit('update', { ...props.step, steps })
}
</script>
