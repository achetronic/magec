<template>
  <div class="text-center py-16 text-arena-500">
    <div v-if="icon" class="mx-auto mb-4 w-14 h-14 rounded-2xl flex items-center justify-center" :class="iconBg">
      <Icon :name="icon" size="lg" :class="iconText" :stroke-width="1.2" />
    </div>
    <div v-else class="w-10 h-10 mx-auto mb-3 text-arena-600">
      <slot name="icon" />
    </div>
    <p class="text-sm font-medium text-arena-300">{{ title }}</p>
    <p v-if="subtitle" class="text-xs mt-1 text-arena-500 max-w-xs mx-auto">{{ subtitle }}</p>
    <button
      v-if="actionLabel"
      @click="$emit('action')"
      class="mt-4 px-4 py-2 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-semibold rounded-lg transition-colors"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  icon: { type: String, default: '' },
  color: { type: String, default: '' },
  actionLabel: { type: String, default: '' },
})

defineEmits(['action'])

const iconBg = computed(() => {
  const map = {
    purple: 'bg-purple-500/10',
    green: 'bg-green-500/10',
    atlantico: 'bg-atlantico-500/10',
    sol: 'bg-sol-500/10',
    rose: 'bg-rose-500/10',
    indigo: 'bg-indigo-500/10',
    teal: 'bg-teal-500/10',
    lava: 'bg-lava-500/10',
  }
  return map[props.color] || 'bg-piedra-800/60'
})

const iconText = computed(() => {
  const map = {
    purple: 'text-purple-400',
    green: 'text-green-400',
    atlantico: 'text-atlantico-400',
    sol: 'text-sol-400',
    rose: 'text-rose-400',
    indigo: 'text-indigo-400',
    teal: 'text-teal-400',
    lava: 'text-lava-400',
  }
  return map[props.color] || 'text-arena-500'
})
</script>
