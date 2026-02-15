<template>
  <div class="flex items-center gap-1.5 sm:gap-2 px-2 sm:px-3 py-1 sm:py-1.5 bg-piedra-800/50 rounded-lg mr-1 sm:mr-2">
    <div class="relative flex items-center justify-center w-3 h-3 sm:w-4 sm:h-4">
      <span
        class="absolute inline-flex h-full w-full rounded-full opacity-75"
        :class="[colors.ping, colors.anim]"
      />
      <span
        class="relative inline-flex rounded-full h-1.5 w-1.5 sm:h-2 sm:w-2"
        :class="colors.dot"
      />
    </div>
    <span class="text-[10px] sm:text-xs text-arena-300 hidden sm:inline">{{ store.status.text }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '../lib/stores/app.js'

const STATUS_COLORS = {
  listening: { ping: 'bg-atlantico-400', dot: 'bg-atlantico-500', anim: '' },
  recording: { ping: 'bg-lava-400', dot: 'bg-lava-500', anim: 'animate-ping' },
  processing: { ping: 'bg-sol-400', dot: 'bg-sol-500', anim: 'animate-ping' },
  loading: { ping: 'bg-sol-400', dot: 'bg-sol-500', anim: 'animate-pulse' },
  default: { ping: 'bg-arena-400', dot: 'bg-arena-500', anim: '' }
}

const store = useAppStore()

const colors = computed(() => STATUS_COLORS[store.status.type] || STATUS_COLORS.default)
</script>
