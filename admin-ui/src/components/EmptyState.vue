<template>
  <div class="text-center py-12 text-arena-500">
    <div class="w-10 h-10 mx-auto mb-3 text-arena-600">
      <slot name="icon" />
    </div>
    <p class="text-sm">{{ title }}</p>
    <p v-if="subtitle" class="text-xs mt-1">{{ subtitle }}</p>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
})
</script>
