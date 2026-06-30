<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <LoginScreen v-if="showLogin" @authenticated="onAuthenticated" />

  <div v-else class="h-full flex">
    <!-- Mobile backdrop -->
    <Transition name="fade">
      <div v-if="mobileOpen" class="fixed inset-0 bg-black/50 z-30 md:hidden" @click="mobileOpen = false" />
    </Transition>

    <!-- Sidebar -->
    <Sidebar
      :active="activeTab"
      :collapsed="sidebarCollapsed"
      :mobile-open="mobileOpen"
      @navigate="onNavigate"
      @toggle="sidebarCollapsed = !sidebarCollapsed"
    />

    <!-- Main area -->
    <div class="flex-1 flex flex-col min-w-0">
      <!-- Top bar -->
      <TopBar :activeTab="activeTab" @search="searchRef?.show()" @menu="mobileOpen = !mobileOpen" />

      <!-- Content -->
      <main class="flex-1 overflow-y-auto">
        <Transition name="section" mode="out-in">
          <div :key="activeTab" class="max-w-5xl mx-auto px-4 sm:px-6 py-5">
            <BackendsList v-if="activeTab === 'backends'" />
            <MemoryList v-else-if="activeTab === 'memory'" />
            <McpsList v-else-if="activeTab === 'mcps'" />
            <AgentsList v-else-if="activeTab === 'agents'" />
            <FlowsList v-else-if="activeTab === 'flows'" />
            <CommandsList v-else-if="activeTab === 'commands'" />
            <SkillsList v-else-if="activeTab === 'skills'" />
            <ClientsList v-else-if="activeTab === 'clients'" />
            <SecretsList v-else-if="activeTab === 'secrets'" />
            <ConversationsView v-else-if="activeTab === 'conversations'" />
            <SettingsView v-else-if="activeTab === 'settings'" />
          </div>
        </Transition>
      </main>
    </div>

    <ConfirmDialog
      ref="confirmDialog"
      :message="confirmMessage"
      @confirm="onConfirmDelete"
    />
    <Toast ref="toastRef" />
    <SearchPalette ref="searchRef" @navigate="activeTab = $event" />
  </div>
</template>

<script setup>
import { ref, provide, onMounted, onUnmounted, watch } from 'vue'
import { useDataStore } from './lib/stores/data.js'
import { checkAuth, isAuthRequired, isAuthenticated } from './lib/auth.js'
import LoginScreen from './components/LoginScreen.vue'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import Toast from './components/Toast.vue'
import SearchPalette from './components/SearchPalette.vue'
import BackendsList from './views/backends/BackendsList.vue'
import MemoryList from './views/memory/MemoryList.vue'
import McpsList from './views/mcps/McpsList.vue'
import AgentsList from './views/agents/AgentsList.vue'
import FlowsList from './views/flows/FlowsList.vue'
import CommandsList from './views/commands/CommandsList.vue'
import SkillsList from './views/skills/SkillsList.vue'
import ClientsList from './views/clients/ClientsList.vue'
import SecretsList from './views/secrets/SecretsList.vue'
import ConversationsView from './views/conversations/ConversationsView.vue'
import SettingsView from './views/settings/SettingsView.vue'

const store = useDataStore()

const showLogin = ref(false)
const appReady = ref(false)

const validTabs = ['backends', 'memory', 'mcps', 'agents', 'flows', 'commands', 'skills', 'clients', 'secrets', 'conversations', 'settings']
const saved = location.hash.slice(1)
const activeTab = ref(validTabs.includes(saved) ? saved : 'backends')
const sidebarCollapsed = ref(localStorage.getItem('sidebar-collapsed') === 'true')
const mobileOpen = ref(false)

function onNavigate(tab) {
  activeTab.value = tab
  mobileOpen.value = false
}

function onAuthenticated() {
  showLogin.value = false
  store.init()
}

watch(sidebarCollapsed, (v) => {
  localStorage.setItem('sidebar-collapsed', v)
})

watch(activeTab, (tab) => {
  location.hash = tab
})

const toastRef = ref(null)
const searchRef = ref(null)
function toast(message, type) {
  toastRef.value?.show(message, type)
}
toast.success = (msg) => toastRef.value?.success(msg)
toast.error = (msg) => toastRef.value?.error(msg)
toast.info = (msg) => toastRef.value?.info(msg)

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
provide('toast', toast)

const newEntityHandler = ref(null)
provide('registerNew', (fn) => { newEntityHandler.value = fn })

function onGlobalKeydown(e) {
  if (showLogin.value) return
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') return
  if (e.target.closest('dialog[open]')) return

  if (e.key === 'n' && !e.metaKey && !e.ctrlKey && !e.altKey) {
    e.preventDefault()
    newEntityHandler.value?.()
  }
  if (e.key === 'r' && !e.metaKey && !e.ctrlKey && !e.altKey) {
    e.preventDefault()
    store.refresh()
  }
}

onMounted(async () => {
  document.addEventListener('keydown', onGlobalKeydown)
  const ok = await checkAuth()
  if (!ok && isAuthRequired()) {
    showLogin.value = true
  } else {
    store.init()
  }
})
onUnmounted(() => document.removeEventListener('keydown', onGlobalKeydown))
</script>
