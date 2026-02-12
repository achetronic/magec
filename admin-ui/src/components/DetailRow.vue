<template>
  <div class="detail-row flex justify-between items-baseline gap-4">
    <span class="text-[11px] text-arena-500 flex-shrink-0">{{ label }}</span>
    <span class="text-[11px] text-arena-200 text-right break-all">{{ value || '—' }}</span>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], default: '' },
})
</script>
