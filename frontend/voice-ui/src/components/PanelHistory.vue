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
        :is-tool="msg.isTool"
        :tool-name="msg.toolName"
        :tool-args="msg.toolArgs"
        :tool-result="msg.toolResult"
        :file-parts="msg.fileParts"
      />
    </div>

    <div class="p-3 border-t border-piedra-700/50 flex-shrink-0">
      <div v-if="attachedFiles.length > 0" class="flex flex-wrap gap-2 mb-2">
        <div v-for="(file, i) in attachedFiles" :key="i" class="relative group rounded-md overflow-hidden bg-piedra-800 border border-piedra-700/50 flex items-center justify-center w-12 h-12">
          <img v-if="file.mimeType.startsWith('image/')" :src="'data:' + file.mimeType + ';base64,' + file.data" class="w-full h-full object-cover" />
          <svg v-else class="w-6 h-6 text-arena-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/>
          </svg>
          <button @click="removeFile(i)" class="absolute -top-1 -right-1 p-0.5 bg-lava-500 rounded-full opacity-0 group-hover:opacity-100 transition-opacity text-white">
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
          </button>
        </div>
      </div>
      <form class="relative flex items-end" @submit.prevent="onSubmit">
        <div class="relative flex-1 flex bg-piedra-800/50 border border-piedra-700/50 rounded-lg transition-colors focus-within:ring-1 focus-within:ring-sol-500 focus-within:border-sol-500">
          <button
            type="button"
            class="p-2.5 text-arena-500 hover:text-sol-400 transition-colors flex-shrink-0 self-end"
            :title="t('actions.attach')"
            @click="fileInputRef.click()"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"/>
            </svg>
          </button>
          <input type="file" ref="fileInputRef" class="hidden" multiple @change="onFileChange" accept="image/*,application/pdf,text/*">
          <textarea
            ref="textareaRef"
            v-model="textInput"
            rows="1"
            :placeholder="t('assistant.textInputPlaceholder')"
            class="w-full bg-transparent border-none pl-1 pr-10 py-2.5 text-sm text-arena-100 placeholder-arena-500 focus:outline-none focus:ring-0 resize-none overflow-y-auto"
            style="max-height: 150px; min-height: 40px;"
            @keydown="handleKeydown"
            @input="autoResize"
          />
          <button
            type="submit"
            class="absolute right-2 bottom-1.5 p-1 text-arena-500 hover:text-sol-400 transition-colors"
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
const textareaRef = ref(null)
const fileInputRef = ref(null)
const attachedFiles = ref([])

function handleKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    onSubmit()
  }
}

function autoResize() {
  if (!textareaRef.value) return
  textareaRef.value.style.height = 'auto'
  textareaRef.value.style.height = Math.min(textareaRef.value.scrollHeight, 150) + 'px'
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const b64 = reader.result.toString().split(',')[1]
      resolve(b64)
    }
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

async function onFileChange(e) {
  const files = e.target.files
  if (!files) return
  
  // Total size validation: < 10MB
  let totalSize = attachedFiles.value.reduce((acc, f) => acc + f.size, 0)
  for (const f of files) {
    if (f.size > 5 * 1024 * 1024) {
      alert(`File ${f.name} is too large (max 5MB)`)
      continue
    }
    if (totalSize + f.size > 10 * 1024 * 1024) {
      alert("Total attachment size exceeds 10MB limit")
      break
    }
    if (attachedFiles.value.length >= 10) {
      alert("Max 10 files allowed")
      break
    }
    
    totalSize += f.size
    try {
      const b64 = await fileToBase64(f)
      attachedFiles.value.push({
        mimeType: f.type || 'application/octet-stream',
        data: b64,
        size: f.size
      })
    } catch (err) {
      console.error('Failed to read file', err)
    }
  }
  
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

function removeFile(index) {
  attachedFiles.value.splice(index, 1)
}

function onSubmit() {
  const text = textInput.value.trim()
  if (!text && attachedFiles.value.length === 0) return
  
  const inlineDataParts = attachedFiles.value.map(f => ({
    inlineData: { mimeType: f.mimeType, data: f.data }
  }))
  
  store.sendTextMessage(text, inlineDataParts)
  
  textInput.value = ''
  attachedFiles.value = []
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto'
  }
}

watch(() => store.messages.length, () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
})
</script>
