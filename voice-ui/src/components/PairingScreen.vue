<template>
  <div class="fixed inset-0 z-50 bg-piedra-950 flex items-center justify-center">
    <div class="w-full max-w-sm px-6">
      <div class="text-center mb-8">
        <img src="/assets/logo.svg" alt="Magec" class="w-16 h-16 mx-auto mb-4">
        <h1 class="text-2xl font-semibold text-arena-50">Magec</h1>
        <p class="text-sm text-arena-500 mt-1">{{ t('pairing.subtitle') }}</p>
      </div>
      <div class="space-y-4">
        <div>
          <input
            ref="inputRef"
            v-model="token"
            type="text"
            class="w-full bg-piedra-800 border border-piedra-700 rounded-xl px-4 py-3 text-sm font-mono text-arena-100 placeholder-arena-500 focus:outline-none focus:ring-2 focus:ring-sol-500 focus:border-sol-500 transition-colors text-center"
            placeholder="mgc_..."
            autocomplete="off"
            spellcheck="false"
            @keydown.enter="onPair"
          >
        </div>
        <button
          :disabled="!token.trim() || connecting"
          class="w-full px-4 py-3 bg-sol-500 hover:bg-sol-600 disabled:bg-piedra-700 disabled:text-arena-500 text-piedra-950 text-sm font-semibold rounded-xl transition-colors"
          @click="onPair"
        >
          {{ connecting ? t('pairing.connecting') : t('pairing.connect') }}
        </button>
        <p v-if="error" class="text-xs text-lava-400 text-center">{{ t('pairing.error') }}</p>
        <p class="text-[10px] text-arena-600 text-center">{{ t('pairing.hint') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'

const store = useAppStore()
const token = ref('')
const connecting = ref(false)
const error = ref(false)
const inputRef = ref(null)

async function onPair() {
  if (!token.value.trim() || connecting.value) return

  connecting.value = true
  error.value = false

  const ok = await store.pair(token.value.trim())
  if (!ok) {
    error.value = true
    connecting.value = false
    token.value = ''
    inputRef.value?.focus()
  }
}

onMounted(() => {
  inputRef.value?.focus()
})
</script>
