<template>
  <div
    class="bg-piedra-900 border rounded-xl p-4 transition-colors"
    :class="active ? 'border-sol-500/30' : 'border-piedra-700/50 hover:border-piedra-600/50'"
  >
    <slot />
  </div>
</template>

<script setup>
defineProps({
  active: { type: Boolean, default: false },
})
</script>
