<template>
  <span
    class="px-1.5 py-0.5 text-[10px] rounded font-medium"
    :class="variantClasses"
  >
    <slot />
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'default' },
})

const variantClasses = computed(() => {
  const map = {
    default: 'bg-piedra-800 text-arena-300',
    sol: 'bg-sol-500/10 text-sol-300',
    atlantico: 'bg-atlantico-500/15 text-atlantico-300',
    lava: 'bg-lava-500/15 text-lava-300',
    purple: 'bg-purple-500/15 text-purple-300',
    green: 'bg-green-500/15 text-green-300',
    teal: 'bg-teal-500/15 text-teal-300',
    rose: 'bg-rose-500/15 text-rose-300',
    indigo: 'bg-indigo-500/15 text-indigo-300',
    muted: 'bg-piedra-800 text-arena-500',
  }
  return map[props.variant] || map.default
})
</script>
