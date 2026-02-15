<template>
  <div class="flex-1 flex flex-col p-4 gap-4 overflow-hidden">
    <div class="flex-1 flex items-center justify-center overflow-hidden min-h-[180px]">
      <CentellaOrb />
    </div>
    <p
      class="text-center text-xs text-arena-500 flex-shrink-0 transition-opacity duration-300"
      :style="{ opacity: store.isRecording ? 0 : 1 }"
    >
      {{ t('assistant.hint') }}
    </p>
  </div>
</template>

<script setup>
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'
import CentellaOrb from './CentellaOrb.vue'

const store = useAppStore()
</script>
