<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

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
      <div v-if="role === 'user'" class="text-sm text-arena-100 whitespace-pre-wrap">
        <div v-if="fileParts && fileParts.length > 0" class="flex flex-wrap gap-2 mb-2">
          <div v-for="(file, i) in fileParts" :key="i" class="w-16 h-16 rounded-md overflow-hidden bg-piedra-900 border border-piedra-700/50 flex items-center justify-center">
            <img v-if="file.inlineData && file.inlineData.mimeType.startsWith('image/')" :src="'data:' + file.inlineData.mimeType + ';base64,' + file.inlineData.data" class="w-full h-full object-cover" />
            <svg v-else class="w-6 h-6 text-arena-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/>
            </svg>
          </div>
        </div>
        {{ text }}
      </div>
      <div v-else-if="isTool" class="text-sm">
        <button
          class="w-full flex items-center justify-between p-2 rounded-md hover:bg-piedra-800/50 transition-colors"
          @click="showToolDetails = !showToolDetails"
        >
          <div class="flex items-center gap-2">
            <svg class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
            <span class="text-arena-200 font-medium font-mono text-xs">{{ toolName }}</span>
          </div>
          <svg class="w-4 h-4 text-arena-500 transition-transform duration-200" :class="showToolDetails ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 9l-7 7-7-7"/>
          </svg>
        </button>
        <div v-if="showToolDetails" class="mt-2 p-3 bg-piedra-950/50 rounded-lg border border-piedra-700/30 overflow-x-auto">
          <div v-if="toolArgs" class="mb-2">
            <span class="text-xs text-arena-500 block mb-1">Input</span>
            <pre class="text-[11px] text-arena-300 font-mono">{{ JSON.stringify(toolArgs, null, 2) }}</pre>
          </div>
          <div v-if="toolResult">
            <span class="text-xs text-arena-500 block mb-1">Output</span>
            <pre class="text-[11px] text-arena-300 font-mono">{{ JSON.stringify(toolResult, null, 2) }}</pre>
          </div>
        </div>
      </div>
      <div v-else class="text-sm text-arena-100" v-html="renderedText" />
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
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
  },
  tool: {
    avatar: 'bg-piedra-800/50 border border-piedra-700/50',
    avatarIcon: 'text-arena-500',
    iconPath: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z',
    bubble: 'bg-transparent border border-piedra-700/50'
  }
}

const props = defineProps({
  role: { type: String, required: true },
  text: { type: String, default: '' },
  isTool: { type: Boolean, default: false },
  toolName: { type: String, default: '' },
  toolArgs: { type: Object, default: null },
  toolResult: { type: Object, default: null },
  fileParts: { type: Array, default: () => [] }
})

const showToolDetails = ref(false)

const style = computed(() => {
  if (props.isTool) return MESSAGE_STYLES.tool
  return MESSAGE_STYLES[props.role] || MESSAGE_STYLES.ai
})
const renderedText = computed(() => renderMarkdown(props.text))
</script>
