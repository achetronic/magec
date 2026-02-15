<template>
  <div class="flex gap-3 items-start">
    <div
      class="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0"
      :class="style.avatar"
    >
      <svg class="w-4 h-4" :class="style.avatarIcon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="style.iconPath"/>
      </svg>
    </div>
    <div class="flex-1 rounded-2xl rounded-tl-md px-4 py-3" :class="style.bubble">
      <div v-if="role === 'user'" class="text-sm text-arena-100">{{ text }}</div>
      <div v-else class="text-sm text-arena-100" v-html="renderedText" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { renderMarkdown } from '../lib/utils/format.js'

const MESSAGE_STYLES = {
  user: {
    avatar: 'bg-piedra-800',
    avatarIcon: 'text-arena-400',
    iconPath: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
    bubble: 'bg-piedra-800'
  },
  ai: {
    avatar: 'bg-sol-500/20',
    avatarIcon: 'text-sol-400',
    iconPath: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6m-6 4h6',
    bubble: 'bg-sol-500/20'
  }
}

const props = defineProps({
  role: { type: String, required: true },
  text: { type: String, required: true }
})

const style = computed(() => MESSAGE_STYLES[props.role] || MESSAGE_STYLES.ai)
const renderedText = computed(() => renderMarkdown(props.text))
</script>
