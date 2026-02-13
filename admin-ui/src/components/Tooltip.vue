<template>
  <div class="relative inline-flex group/tooltip">
    <slot />
    <div v-if="text" class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1.5 px-2.5 py-1.5 bg-piedra-800 border border-piedra-700/60 text-arena-200 text-[10px] leading-snug rounded-lg shadow-lg shadow-black/40 opacity-0 invisible group-hover/tooltip:opacity-100 group-hover/tooltip:visible transition-all duration-150 pointer-events-none z-50 w-max max-w-xs whitespace-normal">
      <span>{{ text }}</span>
      <div class="absolute top-full left-1/2 -translate-x-1/2 -mt-px w-0 h-0 border-x-4 border-x-transparent border-t-4 border-t-piedra-700/60" />
    </div>
  </div>
</template>

<script setup>
defineProps({
  text: { type: String, default: '' },
})
</script>
