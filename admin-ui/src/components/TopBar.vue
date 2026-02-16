<template>
  <header class="flex items-center justify-between px-5 py-4 border-b border-piedra-700/50 bg-piedra-900/60 flex-shrink-0">
    <!-- Left: menu + section title -->
    <div class="flex items-center gap-3 min-w-0">
      <button
        @click="$emit('menu')"
        class="md:hidden w-7 h-7 flex items-center justify-center rounded-lg text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80 transition-colors flex-shrink-0"
      >
        <Icon name="menu" size="md" />
      </button>
      <div
        class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0"
        :class="sectionIconBg"
      >
        <Icon :name="section.icon" size="sm" :class="sectionIconText" />
      </div>
      <div class="min-w-0">
        <h2 class="text-sm font-semibold text-arena-100 leading-tight truncate">{{ section.label }}</h2>
        <p class="text-[9px] text-arena-500 leading-tight">{{ section.group }}</p>
      </div>
    </div>

    <!-- Right: search + stats + refresh -->
    <div class="flex items-center gap-2">
      <button
        @click="$emit('search')"
        class="hidden sm:flex items-center gap-1.5 px-2.5 py-1 rounded-lg border border-piedra-700/50 text-arena-500 hover:text-arena-300 hover:border-piedra-600 transition-colors"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <span class="text-[10px]">Search</span>
        <kbd class="px-1 py-0.5 text-[9px] font-mono bg-piedra-800 border border-piedra-700/50 rounded">⌘K</kbd>
      </button>
      <div class="hidden sm:flex items-center gap-1.5">
        <span
          v-for="stat in stats"
          :key="stat.label"
          class="flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium"
          :class="stat.classes"
        >
          <span class="tabular-nums">{{ stat.count }}</span>
          <span class="text-arena-500">{{ stat.label }}</span>
        </span>
      </div>
      <div class="w-px h-5 bg-piedra-700/50 hidden sm:block" />
      <button
        @click="onRefresh"
        class="flex items-center justify-center w-7 h-7 rounded-lg text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80 transition-colors"
        title="Refresh data"
      >
        <Icon name="refresh" size="sm" :class="{ 'animate-spin': refreshing }" />
      </button>
    </div>
  </header>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useDataStore } from '../lib/stores/data.js'
import Icon from './Icon.vue'

const props = defineProps({
  activeTab: { type: String, required: true },
})

defineEmits(['search', 'menu'])

const store = useDataStore()
const refreshing = ref(false)

async function onRefresh() {
  refreshing.value = true
  try {
    await store.refresh()
  } finally {
    setTimeout(() => { refreshing.value = false }, 400)
  }
}

const sections = {
  backends:  { label: 'Backends',     icon: 'server',   color: 'purple',    group: 'Infrastructure' },
  memory:    { label: 'Memory',       icon: 'database',  color: 'green',     group: 'Infrastructure' },
  mcps:      { label: 'MCP Servers',  icon: 'bolt',      color: 'atlantico', group: 'Infrastructure' },
  agents:    { label: 'Agents',       icon: 'users',     color: 'sol',       group: 'Agents' },
  flows:     { label: 'Flows',        icon: 'flow',      color: 'rose',      group: 'Agents' },
  commands:  { label: 'Commands',     icon: 'command',   color: 'indigo',    group: 'Agents' },
  clients:   { label: 'Clients',      icon: 'phone',     color: 'lava',      group: 'Connections' },
  conversations: { label: 'Conversations', icon: 'chat', color: 'teal',      group: 'Audit' },
}

const section = computed(() => sections[props.activeTab] || sections.backends)

const sectionIconBg = computed(() => {
  const map = {
    purple: 'bg-purple-500/15',
    green: 'bg-green-500/15',
    atlantico: 'bg-atlantico-500/15',
    sol: 'bg-sol-500/15',
    rose: 'bg-rose-500/15',
    indigo: 'bg-indigo-500/15',
    teal: 'bg-teal-500/15',
    lava: 'bg-lava-500/15',
  }
  return map[section.value.color] || map.sol
})

const sectionIconText = computed(() => {
  const map = {
    purple: 'text-purple-400',
    green: 'text-green-400',
    atlantico: 'text-atlantico-400',
    sol: 'text-sol-400',
    rose: 'text-rose-400',
    indigo: 'text-indigo-400',
    teal: 'text-teal-400',
    lava: 'text-lava-400',
  }
  return map[section.value.color] || map.sol
})

const stats = computed(() => [
  { label: 'agents',   count: store.agents.length,   classes: 'bg-sol-500/10 text-sol-400' },
  { label: 'backends', count: store.backends.length,  classes: 'bg-purple-500/10 text-purple-400' },
  { label: 'clients',  count: store.clients.length,   classes: 'bg-lava-500/10 text-lava-400' },
])
</script>
