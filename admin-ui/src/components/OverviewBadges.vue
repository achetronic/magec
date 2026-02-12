<template>
  <div class="flex items-center gap-2 text-[11px]">
    <span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">{{ store.backends.length }} backend{{ store.backends.length !== 1 ? 's' : '' }}</span>
    <span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">{{ store.mcps.length }} MCP{{ store.mcps.length !== 1 ? 's' : '' }}</span>
    <span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">{{ store.agents.length }} agent{{ store.agents.length !== 1 ? 's' : '' }}</span>
    <span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">{{ store.clients.length }} client{{ store.clients.length !== 1 ? 's' : '' }}</span>
    <span class="px-2 py-0.5 bg-piedra-800 rounded-full text-arena-300">{{ store.flows.length }} flow{{ store.flows.length !== 1 ? 's' : '' }}</span>
  </div>
</template>

<script setup>
import { useDataStore } from '../lib/stores/data.js'
const store = useDataStore()
</script>
