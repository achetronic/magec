<template>
  <div class="inline-flex items-center bg-piedra-800 rounded-lg border border-piedra-700/50 p-0.5">
    <button
      v-for="opt in options" :key="opt.value"
      type="button"
      :disabled="opt.disabled"
      @click="!opt.disabled && $emit('update:modelValue', opt.value)"
      class="flex items-center gap-1 px-2.5 py-1 text-[10px] font-medium rounded-md transition-colors"
      :class="modelValue === opt.value
        ? 'bg-piedra-700 text-arena-100'
        : opt.disabled
          ? 'text-arena-600 cursor-not-allowed'
          : 'text-arena-500 hover:text-arena-300'"
    >
      <Icon v-if="opt.icon" :name="opt.icon" size="xs" />
      <span>{{ opt.label }}</span>
    </button>
  </div>
</template>

<script setup>
import Icon from './Icon.vue'

defineProps({
  modelValue: { type: [String, Number, Boolean], required: true },
  options: { type: Array, required: true },
})
defineEmits(['update:modelValue'])
</script>
