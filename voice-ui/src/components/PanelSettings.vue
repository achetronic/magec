<template>
  <div class="flex-1 flex flex-col min-h-0 overflow-hidden">
    <div class="flex items-center gap-2 p-3 border-b border-piedra-700/50 flex-shrink-0">
      <button class="p-1.5 hover:bg-piedra-800 rounded-lg transition-colors" @click="store.switchPanel('assistant')">
        <svg class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 19l-7-7 7-7"/>
        </svg>
      </button>
      <span class="text-sm text-arena-400">{{ t('settings.title') }}</span>
    </div>
    <div class="flex-1 p-5 overflow-y-auto">
      <div class="space-y-4">

        <div class="group">
          <div class="flex items-center justify-between p-4 bg-piedra-800/40 rounded-xl border border-piedra-700/30 hover:border-piedra-600/50 transition-colors">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-lg bg-atlantico-500/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-atlantico-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
                </svg>
              </div>
              <div>
                <span class="text-sm font-medium text-arena-100">{{ t('settings.wakeWord.title') }}</span>
                <p class="text-xs text-arena-500" :class="{ 'text-arena-600': !store.wakeWordEnabled }">
                  {{ store.wakeWordEnabled ? t('settings.wakeWord.description', { phrase: store.wakeWordPhrase }) : t('settings.wakeWord.disabled') }}
                </p>
              </div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                class="sr-only peer"
                :checked="store.wakeWordEnabled"
                :disabled="!store.wakeWordAvailable"
                @change="store.toggleWakeWord($event.target.checked)"
              >
              <div class="w-11 h-6 bg-piedra-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-arena-400 after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-sol-500 peer-checked:after:bg-white" />
            </label>
          </div>
          <div v-if="store.wakeWordModels.length > 0" class="mt-3 pt-3 border-t border-piedra-700/30">
            <span class="text-xs text-arena-400 mb-2 block">{{ t('settings.wakeWord.model') }}</span>
            <div class="grid grid-cols-3 gap-2">
              <label v-for="model in store.wakeWordModels" :key="model.id" class="relative">
                <input
                  type="radio"
                  name="wakeWordModel"
                  :value="model.id"
                  :checked="model.id === store.activeWakeWordModel"
                  class="peer sr-only"
                  @change="store.setWakeWordModel(model.id)"
                >
                <div class="p-3 rounded-lg border border-piedra-600/50 bg-piedra-700/30 cursor-pointer transition-all peer-checked:border-atlantico-500 peer-checked:bg-atlantico-500/10 hover:border-piedra-500">
                  <div class="text-sm font-medium text-arena-100 mb-0.5">{{ model.name }}</div>
                  <div class="text-xs text-arena-500 italic">{{ model.phrase }}</div>
                </div>
              </label>
            </div>
          </div>
        </div>

        <div class="group">
          <div class="p-4 bg-piedra-800/40 rounded-xl border border-piedra-700/30">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-lg bg-lava-500/20 flex items-center justify-center">
                  <svg class="w-5 h-5 text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z"/>
                  </svg>
                </div>
                <span class="text-sm font-medium text-arena-100">{{ t('settings.tts.title') }}</span>
              </div>
              <label class="relative inline-flex items-center cursor-pointer" :class="{ 'opacity-50 cursor-not-allowed': !store.ttsAvailable }">
                <input
                  type="checkbox"
                  class="sr-only peer"
                  :checked="store.ttsEnabled"
                  :disabled="!store.ttsAvailable"
                  @change="store.toggleTTS($event.target.checked)"
                >
                <div class="w-11 h-6 bg-piedra-700 rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-arena-300 after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-sol-500" />
              </label>
            </div>
          </div>
        </div>

        <div class="group">
          <div class="p-4 bg-piedra-800/40 rounded-xl border border-piedra-700/30">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-9 h-9 rounded-lg bg-arena-500/20 flex items-center justify-center">
                <svg class="w-5 h-5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"/>
                </svg>
              </div>
              <span class="text-sm font-medium text-arena-100">{{ t('settings.language.title') }}</span>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <label class="relative">
                <input
                  type="radio"
                  name="language"
                  value="es"
                  :checked="currentLang === 'es'"
                  class="peer sr-only"
                  @change="store.changeLanguage('es')"
                >
                <div class="p-3 rounded-lg border border-piedra-600/50 bg-piedra-700/30 cursor-pointer transition-all peer-checked:border-sol-500 peer-checked:bg-sol-500/10 hover:border-piedra-500 text-center">
                  <div class="text-sm font-medium text-arena-100">Español</div>
                </div>
              </label>
              <label class="relative">
                <input
                  type="radio"
                  name="language"
                  value="en"
                  :checked="currentLang === 'en'"
                  class="peer sr-only"
                  @change="store.changeLanguage('en')"
                >
                <div class="p-3 rounded-lg border border-piedra-600/50 bg-piedra-700/30 cursor-pointer transition-all peer-checked:border-sol-500 peer-checked:bg-sol-500/10 hover:border-piedra-500 text-center">
                  <div class="text-sm font-medium text-arena-100">English</div>
                </div>
              </label>
            </div>
          </div>
        </div>

      </div>

      <div class="mt-6 pt-4 border-t border-piedra-700/30">
        <p class="text-xs text-arena-600 text-center">
          Magec · {{ t('settings.savedAutomatically') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAppStore } from '../lib/stores/app.js'
import { t, getLanguage } from '../lib/i18n/index.js'

const store = useAppStore()
const currentLang = computed(() => getLanguage())
</script>
