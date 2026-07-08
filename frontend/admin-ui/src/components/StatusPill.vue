<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <span class="inline-flex items-stretch rounded overflow-hidden text-[10px] font-medium">
    <span class="w-[18px] flex-shrink-0 rounded-r" :class="colorClass" />
    <span class="bg-piedra-800 text-arena-500 pl-2 pr-1.5 py-0.5">
      <slot />
    </span>
  </span>
</template>

<script setup>
import { computed } from 'vue'

// StatusPill is a split pill: a colored segment on the left states the
// status at a glance, a muted Badge-like body on the right carries the text.
// Full literal class strings so the Tailwind scanner sees them.
const props = defineProps({
  color: { type: String, default: 'arena' },
})

const colorClass = computed(() => {
  const map = {
    green: 'bg-green-400',
    lava: 'bg-lava-400',
    sol: 'bg-sol-400',
    atlantico: 'bg-atlantico-400',
    purple: 'bg-purple-400',
    teal: 'bg-teal-400',
    rose: 'bg-rose-400',
    indigo: 'bg-indigo-400',
    amber: 'bg-amber-400',
    arena: 'bg-arena-500',
  }
  return map[props.color] || map.arena
})
</script>
