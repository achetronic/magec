<template>
  <span v-if="!step" class="text-arena-500">empty</span>
  <span v-else-if="step.type === 'agent'" class="inline-flex px-1.5 py-0.5 bg-sol-500/10 text-sol-300 text-[10px] rounded">
    {{ agentName(step.agentId) }}
  </span>
  <span v-else class="inline-flex items-center gap-0.5 flex-wrap">
    <span class="text-arena-300">{{ label }}</span>
    <span v-if="step.type === 'loop' && step.maxIterations" class="text-arena-500">&times;{{ step.maxIterations }}</span>
    <span class="text-arena-500">(</span>
    <template v-for="(child, i) in step.steps" :key="i">
      <span v-if="i > 0" class="text-arena-600 mx-0.5">&rarr;</span>
      <StepSummary :step="child" :agents="agents" />
    </template>
    <span class="text-arena-500">)</span>
  </span>
</template>

<script setup>
const props = defineProps({
  step: { type: Object, default: null },
  agents: { type: Array, default: () => [] },
})

const label = (() => {
  const map = { sequential: 'Sequential', parallel: 'Parallel', loop: 'Loop' }
  return map[props.step?.type] || props.step?.type || ''
})()

function agentName(id) {
  if (!id) return '?'
  const a = props.agents.find(a => a.id === id)
  return a?.name || a?.id || id
}
</script>
