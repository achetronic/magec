<template>
  <aside
    id="sidebar"
    class="fixed lg:static inset-y-0 left-0 z-30 w-72 bg-piedra-900 border-r border-piedra-700/50 flex flex-col transition-all duration-200"
    :class="sidebarClasses"
  >
    <div class="p-4 border-b border-piedra-700/50">
      <div class="flex items-center justify-between">
        <h2 class="font-medium text-arena-100">{{ t('sessions.title') }}</h2>
        <div class="flex items-center gap-1">
          <button
            class="p-2 hover:bg-piedra-800 rounded-lg transition-colors"
            :title="t('sessions.new')"
            @click="store.newSession()"
          >
            <svg class="w-5 h-5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4v16m8-8H4"/>
            </svg>
          </button>
          <button
            class="p-2 hover:bg-piedra-800 rounded-lg transition-colors hidden lg:block"
            @click="store.sidebarOpen = false"
          >
            <svg class="w-5 h-5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11 19l-7-7 7-7m8 14l-7-7 7-7"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-2 space-y-1">
      <p v-if="!store.sessions.length" class="text-arena-500 text-sm text-center py-4">
        {{ t('sessions.empty') }}
      </p>
      <SessionItem
        v-for="session in store.sessions"
        :key="session.id"
        :session="session"
        :active="session.id === store.currentSessionId"
        @select="store.selectSession(session.id)"
        @delete="store.deleteSession(session.id)"
      />
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'
import SessionItem from './SessionItem.vue'

const store = useAppStore()

const isMobile = computed(() => window.innerWidth < 1024)

const sidebarClasses = computed(() => {
  if (isMobile.value) {
    return store.sidebarOpen ? 'sidebar-open' : 'sidebar-closed'
  }
  return store.sidebarOpen ? '' : 'sidebar-collapsed'
})
</script>
