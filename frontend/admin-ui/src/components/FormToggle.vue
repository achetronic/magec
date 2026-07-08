<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    @click="$emit('update:modelValue', !modelValue)"
    class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border border-piedra-700 transition-colors duration-200 ease-in-out focus:outline-none focus:ring-1 focus:ring-sol-500"
    :class="modelValue ? 'bg-sol-500/30' : 'bg-piedra-800'"
  >
    <span
      class="pointer-events-none inline-block h-4 w-4 rounded-full bg-arena-300 shadow-sm ring-0 transition-transform duration-200 ease-in-out"
      :class="modelValue ? 'translate-x-4' : 'translate-x-0'"
    />
  </button>
</template>

<script setup>
defineProps({
  modelValue: { type: Boolean, default: false },
})

defineEmits(['update:modelValue'])
</script>
