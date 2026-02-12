<template>
  <div class="h-full flex flex-col">
    <!-- Header -->
    <header class="border-b border-piedra-700/50 bg-piedra-900/90 flex-shrink-0">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 py-3 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <img src="/assets/logo.svg" alt="Magec" class="w-7 h-7" />
          <div>
            <h1 class="text-base font-semibold text-arena-50 leading-tight">Magec</h1>
            <p class="text-[10px] text-arena-500">Admin</p>
          </div>
        </div>
        <OverviewBadges />
      </div>
    </header>

    <!-- Tabs -->
    <nav class="border-b border-piedra-700/50 bg-piedra-900/50 flex-shrink-0 overflow-x-auto">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 flex gap-1">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          class="flex items-center gap-1.5 px-3 py-2.5 text-sm font-medium border-b-2 transition-colors whitespace-nowrap"
          :class="activeTab === tab.id
            ? 'border-sol-500 text-arena-50'
            : 'border-transparent text-arena-400 hover:text-arena-300 hover:border-piedra-700'"
        >
          <Icon :name="tab.icon" size="sm" />
          {{ tab.label }}
        </button>
      </div>
    </nav>

    <!-- Content -->
    <main class="flex-1 overflow-y-auto">
      <div class="max-w-5xl mx-auto px-4 sm:px-6 py-5">
        <BackendsList v-if="activeTab === 'backends'" />
        <MemoryList v-else-if="activeTab === 'memory'" />
        <McpsList v-else-if="activeTab === 'mcps'" />
        <AgentsList v-else-if="activeTab === 'agents'" />
        <ClientsList v-else-if="activeTab === 'clients'" />
        <CronsList v-else-if="activeTab === 'crons'" />
        <FlowsList v-else-if="activeTab === 'flows'" />
      </div>
    </main>

    <ConfirmDialog
      ref="confirmDialog"
      :message="confirmMessage"
      @confirm="onConfirmDelete"
    />
  </div>
</template>

<script setup>
import { ref, provide, onMounted, watch } from 'vue'
import { useDataStore } from './lib/stores/data.js'
import Icon from './components/Icon.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import OverviewBadges from './components/OverviewBadges.vue'
import BackendsList from './views/backends/BackendsList.vue'
import MemoryList from './views/memory/MemoryList.vue'
import McpsList from './views/mcps/McpsList.vue'
import AgentsList from './views/agents/AgentsList.vue'
import ClientsList from './views/clients/ClientsList.vue'
import CronsList from './views/crons/CronsList.vue'
import FlowsList from './views/flows/FlowsList.vue'

const store = useDataStore()

const validTabs = ['backends', 'memory', 'mcps', 'agents', 'clients', 'crons', 'flows']
const saved = location.hash.slice(1)
const activeTab = ref(validTabs.includes(saved) ? saved : 'backends')

watch(activeTab, (tab) => {
  location.hash = tab
})

const tabs = [
  { id: 'backends', label: 'Backends', icon: 'server' },
  { id: 'memory', label: 'Memory', icon: 'database' },
  { id: 'mcps', label: 'MCP Servers', icon: 'bolt' },
  { id: 'agents', label: 'Agents', icon: 'users' },
  { id: 'clients', label: 'Clients', icon: 'phone' },
  { id: 'crons', label: 'Crons', icon: 'clock' },
  { id: 'flows', label: 'Flows', icon: 'flow' },
]

const confirmDialog = ref(null)
const confirmMessage = ref('')
let confirmCallback = null

function requestDelete(message, callback) {
  confirmMessage.value = message
  confirmCallback = callback
  confirmDialog.value?.open()
}

function onConfirmDelete() {
  if (confirmCallback) {
    confirmCallback()
    confirmCallback = null
  }
}

provide('requestDelete', requestDelete)

onMounted(() => store.init())
</script>
