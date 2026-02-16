<template>
  <aside
    class="flex flex-col h-full border-r border-piedra-700/50 bg-piedra-900 transition-all duration-200 flex-shrink-0"
    :class="[
      collapsed ? 'w-[52px]' : 'w-56',
      mobileOpen
        ? 'fixed inset-y-0 left-0 z-40 w-56 shadow-2xl shadow-black/50'
        : 'hidden md:flex',
    ]"
  >
    <!-- Logo -->
    <div class="flex items-center gap-3 px-3.5 py-4 border-b border-piedra-700/50">
      <img src="/assets/logo.svg" alt="Magec" class="w-7 h-7 flex-shrink-0" />
      <transition name="fade">
        <div v-if="!collapsed" class="min-w-0">
          <h1 class="text-sm font-bold text-arena-50 leading-tight tracking-tight">Magec</h1>
          <p class="text-[9px] text-arena-500 font-medium">Admin Console</p>
        </div>
      </transition>
    </div>

    <!-- Nav groups -->
    <nav class="flex-1 overflow-y-auto py-3 space-y-4 px-2">
      <div v-for="group in groups" :key="group.label">
        <p
          v-if="!collapsed"
          class="px-2 mb-1.5 text-[9px] font-bold uppercase tracking-widest text-arena-600"
        >
          {{ group.label }}
        </p>
        <div v-else class="mx-auto mb-2 mt-1 w-4 border-t border-piedra-700/40" />

        <div class="space-y-0.5">
          <button
            v-for="item in group.items"
            :key="item.id"
            @click="$emit('navigate', item.id)"
            class="flex items-center gap-2.5 w-full py-2 text-[13px] font-medium rounded-lg transition-all duration-150"
            :class="[
              active === item.id
                ? activeClasses(item.color)
                : 'text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80',
              collapsed ? 'justify-center px-0' : 'px-2.5',
            ]"
            :title="collapsed ? item.label : undefined"
          >
            <div
              class="w-7 h-7 rounded-md flex items-center justify-center flex-shrink-0 transition-colors duration-150"
              :class="active === item.id ? iconBgClasses(item.color) : 'bg-piedra-800/60'"
            >
              <Icon :name="item.icon" size="sm" class="flex-shrink-0" />
            </div>
            <template v-if="!collapsed">
              <span class="truncate flex-1 text-left">{{ item.label }}</span>
              <span
                v-if="itemCount(item.id) > 0"
                class="text-[10px] tabular-nums px-1.5 py-0.5 rounded-full"
                :class="active === item.id ? countActiveClasses(item.color) : 'text-arena-500 bg-piedra-800/80'"
              >
                {{ itemCount(item.id) }}
              </span>
            </template>
          </button>
        </div>
      </div>
    </nav>

    <!-- Collapse toggle -->
    <div class="border-t border-piedra-700/50 px-2 py-2">
      <button
        @click="$emit('toggle')"
        class="flex items-center justify-center w-full py-1.5 rounded-lg text-arena-500 hover:text-arena-300 hover:bg-piedra-800/60 transition-colors"
      >
        <Icon :name="collapsed ? 'chevronRight' : 'chevronLeft'" size="sm" />
      </button>
    </div>
  </aside>
</template>

<script setup>
import Icon from './Icon.vue'
import { useDataStore } from '../lib/stores/data.js'

const props = defineProps({
  active: { type: String, required: true },
  collapsed: { type: Boolean, default: false },
  mobileOpen: { type: Boolean, default: false },
})

defineEmits(['navigate', 'toggle'])

const store = useDataStore()

const groups = [
  {
    label: 'Infrastructure',
    items: [
      { id: 'backends', label: 'Backends', icon: 'server', color: 'purple' },
      { id: 'memory', label: 'Memory', icon: 'database', color: 'green' },
      { id: 'mcps', label: 'MCP Servers', icon: 'bolt', color: 'atlantico' },
    ],
  },
  {
    label: 'Agents',
    items: [
      { id: 'agents', label: 'Agents', icon: 'users', color: 'sol' },
      { id: 'flows', label: 'Flows', icon: 'flow', color: 'rose' },
      { id: 'commands', label: 'Commands', icon: 'command', color: 'indigo' },
    ],
  },
  {
    label: 'Connections',
    items: [
      { id: 'clients', label: 'Clients', icon: 'phone', color: 'lava' },
    ],
  },
  {
    label: 'Audit',
    items: [
      { id: 'conversations', label: 'Conversations', icon: 'chat', color: 'teal' },
    ],
  },
]

function itemCount(id) {
  const map = {
    backends: store.backends,
    memory: store.memory,
    mcps: store.mcps,
    agents: store.agents,
    flows: store.flows,
    commands: store.commands,
    clients: store.clients,
  }
  return map[id]?.length || 0
}

function activeClasses(color) {
  const map = {
    purple: 'bg-purple-500/10 text-purple-200',
    green: 'bg-green-500/10 text-green-200',
    atlantico: 'bg-atlantico-500/10 text-atlantico-200',
    sol: 'bg-sol-500/10 text-sol-200',
    rose: 'bg-rose-500/10 text-rose-200',
    indigo: 'bg-indigo-500/10 text-indigo-200',
    teal: 'bg-teal-500/10 text-teal-200',
    lava: 'bg-lava-500/10 text-lava-200',
    arena: 'bg-arena-500/10 text-arena-200',
  }
  return map[color] || map.sol
}

function iconBgClasses(color) {
  const map = {
    purple: 'bg-purple-500/20 text-purple-300',
    green: 'bg-green-500/20 text-green-300',
    atlantico: 'bg-atlantico-500/20 text-atlantico-300',
    sol: 'bg-sol-500/20 text-sol-300',
    rose: 'bg-rose-500/20 text-rose-300',
    indigo: 'bg-indigo-500/20 text-indigo-300',
    teal: 'bg-teal-500/20 text-teal-300',
    lava: 'bg-lava-500/20 text-lava-300',
    arena: 'bg-arena-500/20 text-arena-300',
  }
  return map[color] || map.sol
}

function countActiveClasses(color) {
  const map = {
    purple: 'text-purple-300 bg-purple-500/15',
    green: 'text-green-300 bg-green-500/15',
    atlantico: 'text-atlantico-300 bg-atlantico-500/15',
    sol: 'text-sol-300 bg-sol-500/15',
    rose: 'text-rose-300 bg-rose-500/15',
    indigo: 'text-indigo-300 bg-indigo-500/15',
    teal: 'text-teal-300 bg-teal-500/15',
    lava: 'text-lava-300 bg-lava-500/15',
    arena: 'text-arena-300 bg-arena-500/15',
  }
  return map[color] || map.sol
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
