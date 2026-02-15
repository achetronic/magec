<template>
  <div class="group relative">
    <button
      class="w-full text-left px-3 py-2.5 rounded-lg transition-colors"
      :class="active ? 'bg-sol-500/20' : 'hover:bg-piedra-800'"
      @click="$emit('select')"
    >
      <p class="text-sm text-arena-200 truncate pr-6">{{ preview }}</p>
      <p class="text-xs text-arena-500 mt-0.5">{{ date }}</p>
    </button>
    <button
      class="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded-md opacity-0 group-hover:opacity-100 hover:bg-piedra-700 transition-all"
      :title="t('sessions.delete')"
      @click.stop="$emit('delete')"
    >
      <svg class="w-4 h-4 text-arena-500 hover:text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
      </svg>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { t } from '../lib/i18n/index.js'
import { formatRelativeDate } from '../lib/utils/format.js'

const props = defineProps({
  session: { type: Object, required: true },
  active: { type: Boolean, default: false }
})

defineEmits(['select', 'delete'])

const preview = computed(() => props.session.preview || t('sessions.emptyPreview'))
const date = computed(() => props.session.createdAt ? formatRelativeDate(props.session.createdAt, t) : '')
</script>
