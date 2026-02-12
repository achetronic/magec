<template>
  <label class="block text-xs text-arena-400 mb-1">
    {{ label }}
    <span v-if="required" class="text-lava-400">*</span>
  </label>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  required: { type: Boolean, default: false },
})
</script>
