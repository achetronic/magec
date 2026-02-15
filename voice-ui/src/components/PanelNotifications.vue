<template>
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <div class="flex items-center justify-between p-3 border-b border-piedra-700/50 flex-shrink-0">
      <div class="flex items-center gap-2">
        <button class="p-1.5 hover:bg-piedra-800 rounded-lg transition-colors" @click="store.switchPanel('assistant')">
          <svg class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19l-7-7 7-7"/>
          </svg>
        </button>
        <span class="text-sm text-arena-400">{{ t('notifications.title') }}</span>
      </div>
      <button
        class="p-2 hover:bg-piedra-800 text-arena-400 hover:text-arena-200 rounded-lg transition-all"
        :title="t('notifications.clearAll')"
        @click="store.clearAllNotifications()"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M3 6h18"/><path d="M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/>
        </svg>
      </button>
    </div>

    <div class="flex-1 p-4 overflow-y-auto space-y-3">
      <p v-if="!store.notifications.length" class="text-arena-500 text-sm text-center py-8">
        {{ t('notifications.empty') }}
      </p>
      <NotificationItem
        v-for="n in store.notifications"
        :key="n.id"
        :notification="n"
        @delete="store.removeNotification(n.id)"
      />
    </div>
  </div>
</template>

<script setup>
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'
import NotificationItem from './NotificationItem.vue'

const store = useAppStore()
</script>
