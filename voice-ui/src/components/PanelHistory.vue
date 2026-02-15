<template>
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <div class="flex items-center justify-between p-3 border-b border-piedra-700/50 flex-shrink-0">
      <div class="flex items-center gap-2">
        <button class="p-1.5 hover:bg-piedra-800 rounded-lg transition-colors" @click="store.switchPanel('assistant')">
          <svg class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19l-7-7 7-7"/>
          </svg>
        </button>
        <span class="text-sm text-arena-400">{{ t('assistant.currentConversation') }}</span>
      </div>
      <div class="flex items-center gap-1">
        <button
          class="p-2 hover:bg-piedra-800 text-arena-400 hover:text-arena-200 rounded-lg transition-all"
          :title="t('sessions.new')"
          @click="store.newSession()"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4v16m8-8H4"/>
          </svg>
        </button>
        <button
          :disabled="!store.hasMessages"
          class="p-2 hover:bg-piedra-800 text-arena-400 hover:text-arena-200 disabled:text-piedra-600 rounded-lg transition-all disabled:cursor-not-allowed"
          :title="t('actions.copy')"
          @click="store.copyMessages()"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
          </svg>
        </button>
        <button
          :disabled="!store.hasMessages"
          class="p-2 hover:bg-piedra-800 text-arena-400 hover:text-arena-200 disabled:text-piedra-600 rounded-lg transition-all disabled:cursor-not-allowed"
          :title="t('actions.clear')"
          @click="store.clearMessages()"
        >
          <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 6h18"/><path d="M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2"/><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6"/>
          </svg>
        </button>
      </div>
    </div>

    <div ref="messagesRef" class="flex-1 p-4 overflow-y-auto space-y-4">
      <p v-if="!store.hasMessages" class="text-arena-500 text-sm text-center py-8">
        {{ t('assistant.placeholder') }}
      </p>
      <ChatMessage
        v-for="(msg, i) in store.messages"
        :key="i"
        :role="msg.role"
        :text="msg.text"
      />
    </div>

    <div class="p-3 border-t border-piedra-700/50 flex-shrink-0">
      <form class="relative flex items-center" @submit.prevent="onSubmit">
        <div class="relative flex-1">
          <input
            v-model="textInput"
            type="text"
            :placeholder="t('assistant.textInputPlaceholder')"
            class="w-full bg-piedra-800/50 border border-piedra-700/50 rounded-lg pl-3 pr-10 py-2.5 text-sm text-arena-100 placeholder-arena-500 focus:outline-none focus:ring-1 focus:ring-sol-500 focus:border-sol-500 transition-colors"
          >
          <button
            type="submit"
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-arena-500 hover:text-sol-400 transition-colors"
            :title="t('actions.send')"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 7l3-3m0 0l3 3m-3-3v14"/>
            </svg>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'
import ChatMessage from './ChatMessage.vue'

const store = useAppStore()
const textInput = ref('')
const messagesRef = ref(null)

function onSubmit() {
  const text = textInput.value.trim()
  if (!text) return
  store.sendTextMessage(text)
  textInput.value = ''
}

watch(() => store.messages.length, () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
})
</script>
