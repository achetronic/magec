<template>
  <div class="loop-node px-3 py-2 bg-piedra-900 border-2 border-lava-500/40 rounded-xl shadow-lg min-w-[160px]">
    <Handle type="target" :position="Position.Left" />
    <div class="flex items-center justify-between gap-2 mb-1.5">
      <span class="text-[10px] font-medium text-lava-400 uppercase tracking-wider">Loop</span>
      <button @click="$emit('delete-node')" class="p-0.5 hover:bg-piedra-800 rounded text-arena-500 hover:text-lava-400">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <div class="flex items-center gap-1.5">
      <label class="text-[10px] text-arena-500 whitespace-nowrap">Max iter</label>
      <input
        type="number"
        min="0"
        :value="data?.maxIterations || 0"
        @input="$emit('update-data', { maxIterations: parseInt($event.target.value) || 0 })"
        class="w-14 bg-piedra-800 border border-piedra-700 rounded-lg px-2 py-1 text-xs text-arena-200 focus:ring-1 focus:ring-lava-500 outline-none"
      />
      <span class="text-[9px] text-arena-600">0 = &infin;</span>
    </div>
    <p class="text-[9px] text-arena-500 mt-1">Connect child steps &rarr;</p>
    <Handle type="source" :position="Position.Right" />
  </div>
</template>

<script setup>
import { Handle, Position } from '@vue-flow/core'

defineProps({
  id: { type: String, required: true },
  data: { type: Object, default: () => ({}) },
})

defineEmits(['update-data', 'delete-node'])
</script>

<style scoped>
.loop-node :deep(.vue-flow__handle) {
  width: 8px;
  height: 8px;
  background-color: #ef4444;
  border: 2px solid #1c1917;
}
</style>
