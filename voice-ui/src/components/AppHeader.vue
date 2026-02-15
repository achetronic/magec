<template>
  <header class="border-b border-piedra-700/50 backdrop-blur-sm bg-piedra-950/90 flex-shrink-0">
    <div class="px-3 sm:px-4 py-3 sm:py-4 flex items-center justify-between gap-2">
      <div class="flex items-center gap-2 sm:gap-3 min-w-0">
        <button class="p-1.5 sm:p-2 -ml-1 sm:-ml-2 hover:bg-piedra-800 rounded-lg flex-shrink-0" @click="toggleSidebar">
          <svg class="w-5 h-5 text-arena-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 6h16M4 12h16M4 18h16"/>
          </svg>
        </button>
        <div class="min-w-0 flex items-center gap-2">
          <img src="/assets/logo.svg" alt="Magec" class="w-7 h-7 flex-shrink-0">
          <div class="hidden sm:block">
            <h1 class="text-lg sm:text-xl font-semibold tracking-tight text-arena-50 truncate">Magec</h1>
            <p class="text-xs text-arena-500">{{ t('app.subtitle') }}</p>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-0.5 sm:gap-1 flex-shrink-0">
        <StatusIndicator />

        <button
          class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors"
          :class="{ 'bg-piedra-800': store.activePanel === 'assistant' }"
          @click="store.switchPanel('assistant')"
        >
          <svg class="w-5 h-5 text-sol-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
          </svg>
        </button>

        <button
          class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors"
          :class="{ 'bg-piedra-800': store.activePanel === 'history' }"
          @click="store.switchPanel('history')"
        >
          <svg class="w-5 h-5 text-arena-400 hover:text-arena-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
          </svg>
        </button>

        <button
          class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors relative"
          :class="{ 'bg-piedra-800': store.activePanel === 'notifications' }"
          @click="store.switchPanel('notifications')"
        >
          <svg class="w-5 h-5 text-arena-400 hover:text-arena-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
          </svg>
          <span
            v-if="store.notificationCount > 0"
            class="absolute -top-0.5 -right-0.5 w-4 h-4 bg-lava-500 rounded-full text-[10px] font-medium text-white flex items-center justify-center"
          >
            {{ store.notificationCount > 99 ? '99+' : store.notificationCount }}
          </span>
        </button>

        <AgentSwitcher v-if="store.allowedAgents.length > 1" />

        <button
          class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors"
          :class="{ 'bg-piedra-800': store.activePanel === 'settings' }"
          @click="store.switchPanel('settings')"
        >
          <svg class="w-5 h-5 text-arena-400 hover:text-arena-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
          </svg>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'
import StatusIndicator from './StatusIndicator.vue'
import AgentSwitcher from './AgentSwitcher.vue'

const store = useAppStore()

function toggleSidebar() {
  store.sidebarOpen = !store.sidebarOpen
}
</script>
