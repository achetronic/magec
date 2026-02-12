<template>
  <div class="agent-node px-3 py-2 bg-piedra-900 border-2 border-sol-500/40 rounded-xl shadow-lg min-w-[160px]">
    <Handle type="target" :position="Position.Left" />
    <div class="flex items-center justify-between gap-2 mb-1.5">
      <span class="text-[10px] font-medium text-sol-400 uppercase tracking-wider">Agent</span>
      <button @click="$emit('delete-node')" class="p-0.5 hover:bg-piedra-800 rounded text-arena-500 hover:text-lava-400">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
    <select
      :value="data?.agentId || ''"
      @change="$emit('update-data', { agentId: $event.target.value })"
      class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-2 py-1 text-xs text-arena-200 focus:ring-1 focus:ring-sol-500 outline-none"
    >
      <option value="" disabled>Select agent</option>
      <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name || a.id }}</option>
    </select>
    <Handle type="source" :position="Position.Right" />
  </div>
</template>

<script setup>
import { Handle, Position } from '@vue-flow/core'

defineProps({
  id: { type: String, required: true },
  data: { type: Object, default: () => ({}) },
  agents: { type: Array, default: () => [] },
})

defineEmits(['update-data', 'delete-node'])
</script>

<style scoped>
.agent-node :deep(.vue-flow__handle) {
  width: 8px;
  height: 8px;
  background-color: #eab308;
  border: 2px solid #1c1917;
}
</style>
