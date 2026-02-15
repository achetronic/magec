<template>
  <div
    class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 transition-all duration-200"
    :class="hoverClass"
  >
    <slot />
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  active: { type: Boolean, default: false },
  color: { type: String, default: '' },
})

const colorMap = {
  purple: 'hover:border-purple-500/15 hover:shadow-[0_0_15px_-3px_rgba(168,85,247,0.04)]',
  green: 'hover:border-green-500/15 hover:shadow-[0_0_15px_-3px_rgba(74,222,128,0.04)]',
  atlantico: 'hover:border-atlantico-500/15 hover:shadow-[0_0_15px_-3px_rgba(56,188,216,0.04)]',
  sol: 'hover:border-sol-500/15 hover:shadow-[0_0_15px_-3px_rgba(250,204,21,0.04)]',
  rose: 'hover:border-rose-500/15 hover:shadow-[0_0_15px_-3px_rgba(244,63,94,0.04)]',
  indigo: 'hover:border-indigo-500/15 hover:shadow-[0_0_15px_-3px_rgba(99,102,241,0.04)]',
  lava: 'hover:border-lava-500/15 hover:shadow-[0_0_15px_-3px_rgba(239,68,68,0.04)]',
  teal: 'hover:border-teal-500/15 hover:shadow-[0_0_15px_-3px_rgba(20,184,166,0.04)]',
}

const hoverClass = computed(() => {
  if (props.color && colorMap[props.color]) return colorMap[props.color]
  return 'hover:border-piedra-600/50'
})
</script>
