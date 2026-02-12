<template>
  <div class="container-node px-3 py-2 border-2 rounded-xl shadow-lg min-w-[140px]"
    :class="borderClass"
    :style="{ backgroundColor: '#1c1917' }"
  >
    <Handle type="target" :position="Position.Left" />
    <div class="flex items-center justify-between gap-2">
      <span class="text-[10px] font-medium uppercase tracking-wider" :class="textClass">{{ label }}</span>
      <button @click="$emit('delete-node')" class="p-0.5 hover:bg-piedra-800 rounded text-arena-500 hover:text-lava-400">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <p class="text-[9px] text-arena-500 mt-0.5">Connect child steps &rarr;</p>
    <Handle type="source" :position="Position.Right" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps({
  id: { type: String, required: true },
  data: { type: Object, default: () => ({}) },
  label: { type: String, default: 'Container' },
  color: { type: String, default: 'atlantico' },
})

defineEmits(['delete-node'])

const borderClass = computed(() => {
  const map = {
    atlantico: 'border-atlantico-500/40',
    sol: 'border-sol-500/40',
    lava: 'border-lava-500/40',
  }
  return map[props.color] || map.atlantico
})

const textClass = computed(() => {
  const map = {
    atlantico: 'text-atlantico-400',
    sol: 'text-sol-400',
    lava: 'text-lava-400',
  }
  return map[props.color] || map.atlantico
})
</script>

<style scoped>
.container-node :deep(.vue-flow__handle) {
  width: 8px;
  height: 8px;
  background-color: #6b7280;
  border: 2px solid #1c1917;
}
</style>
