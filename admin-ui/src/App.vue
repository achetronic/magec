<template>
  <div class="h-full flex">
    <!-- Sidebar -->
    <Sidebar
      :active="activeTab"
      :collapsed="sidebarCollapsed"
      @navigate="activeTab = $event"
      @toggle="sidebarCollapsed = !sidebarCollapsed"
    />

    <!-- Main area -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top bar -->
      <TopBar :activeTab="activeTab" />

      <!-- Content -->
      <main class="flex-1 overflow-y-auto">
        <div class="max-w-5xl mx-auto px-4 sm:px-6 py-5">
          <BackendsList v-if="activeTab === 'backends'" />
          <MemoryList v-else-if="activeTab === 'memory'" />
          <McpsList v-else-if="activeTab === 'mcps'" />
          <AgentsList v-else-if="activeTab === 'agents'" />
          <FlowsList v-else-if="activeTab === 'flows'" />
          <CommandsList v-else-if="activeTab === 'commands'" />
          <TriggersList v-else-if="activeTab === 'triggers'" />
          <ClientsList v-else-if="activeTab === 'clients'" />
        </div>
      </main>
    </div>

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
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import BackendsList from './views/backends/BackendsList.vue'
import MemoryList from './views/memory/MemoryList.vue'
import McpsList from './views/mcps/McpsList.vue'
import AgentsList from './views/agents/AgentsList.vue'
import FlowsList from './views/flows/FlowsList.vue'
import CommandsList from './views/commands/CommandsList.vue'
import TriggersList from './views/triggers/TriggersList.vue'
import ClientsList from './views/clients/ClientsList.vue'

const store = useDataStore()

const validTabs = ['backends', 'memory', 'mcps', 'agents', 'flows', 'commands', 'triggers', 'clients']
const saved = location.hash.slice(1)
const activeTab = ref(validTabs.includes(saved) ? saved : 'backends')
const sidebarCollapsed = ref(localStorage.getItem('sidebar-collapsed') === 'true')

watch(sidebarCollapsed, (v) => {
  localStorage.setItem('sidebar-collapsed', v)
})

watch(activeTab, (tab) => {
  location.hash = tab
})

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
