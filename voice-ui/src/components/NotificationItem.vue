<template>
  <div class="group flex items-start gap-3 p-3 rounded-xl transition-all hover:bg-opacity-30" :class="config.bg">
    <div class="flex-shrink-0 mt-0.5" v-html="config.icon" />
    <div class="flex-1 min-w-0">
      <p class="text-sm text-arena-200 break-words">{{ notification.message }}</p>
      <p class="text-xs text-arena-500 mt-1">{{ date }}</p>
    </div>
    <button
      class="flex-shrink-0 p-1 opacity-0 group-hover:opacity-100 hover:bg-piedra-700/50 rounded transition-all"
      :title="t('notifications.delete')"
      @click="$emit('delete')"
    >
      <svg class="w-4 h-4 text-arena-500 hover:text-arena-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12"/>
      </svg>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { t } from '../lib/i18n/index.js'
import { formatRelativeDate } from '../lib/utils/format.js'

const ICONS = {
  alert: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
  info: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  check: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
  spinner: 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15'
}

const icon = (path, color, extra = '') =>
  `<svg class="w-4 h-4 text-${color} ${extra}" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="${path}"/></svg>`

const CONFIGS = {
  error: { icon: icon(ICONS.alert, 'lava-400'), bg: 'bg-lava-500/20' },
  warning: { icon: icon(ICONS.alert, 'sol-400'), bg: 'bg-sol-500/20' },
  info: { icon: icon(ICONS.info, 'atlantico-400'), bg: 'bg-atlantico-500/20' },
  loading: { icon: icon(ICONS.spinner, 'sol-400', 'animate-spin'), bg: 'bg-sol-500/20' },
  success: { icon: icon(ICONS.check, 'atlantico-400'), bg: 'bg-atlantico-500/20' }
}

const props = defineProps({
  notification: { type: Object, required: true }
})

defineEmits(['delete'])

const config = computed(() => CONFIGS[props.notification.type] || CONFIGS.info)
const date = computed(() => formatRelativeDate(props.notification.timestamp, t))
</script>
