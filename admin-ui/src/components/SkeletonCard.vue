<template>
  <div :class="grid ? 'grid gap-3 grid-cols-1 sm:grid-cols-2' : 'space-y-3'">
    <div v-for="i in count" :key="i" class="bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 animate-pulse">
      <div class="flex items-start gap-3">
        <div class="w-8 h-8 rounded-lg bg-piedra-800" />
        <div class="flex-1 space-y-2">
          <div class="h-3.5 bg-piedra-800 rounded w-1/3" />
          <div class="h-2.5 bg-piedra-800/60 rounded w-1/2" />
        </div>
      </div>
      <div v-if="badges" class="flex gap-1.5 mt-3">
        <div v-for="j in 2" :key="j" class="h-4 bg-piedra-800/60 rounded w-12" />
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  count: { type: Number, default: 3 },
  badges: { type: Boolean, default: true },
  grid: { type: Boolean, default: true },
})
</script>
