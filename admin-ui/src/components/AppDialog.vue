<template>
  <dialog
    ref="dialogRef"
    class="bg-piedra-900 border border-piedra-700/50 rounded-2xl p-0 text-arena-100 shadow-2xl"
    :class="sizeClass"
    @close="$emit('close')"
  >
    <div class="flex flex-col max-h-[85vh]">
      <div class="flex items-center justify-between p-5 border-b border-piedra-700/50 flex-shrink-0">
        <h3 class="text-lg font-semibold">{{ title }}</h3>
        <button type="button" @click="close" class="p-1.5 hover:bg-piedra-800 rounded-lg">
          <svg class="w-5 h-5 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="p-5 overflow-y-auto flex-1">
        <slot />
      </div>
      <div class="flex justify-end gap-3 p-5 border-t border-piedra-700/50 flex-shrink-0">
        <slot name="footer">
          <button type="button" @click="close" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
            Cancel
          </button>
          <button type="button" @click="$emit('save')" class="px-4 py-2 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-sm font-medium rounded-lg transition-colors">
            Save
          </button>
        </slot>
      </div>
    </div>
  </dialog>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  size: { type: String, default: 'md' },
})

defineEmits(['close', 'save'])

const dialogRef = ref(null)

const sizeClass = computed(() => {
  const map = {
    sm: 'w-full max-w-sm',
    md: 'w-full max-w-lg',
    lg: 'w-full max-w-2xl',
    xl: 'w-full max-w-4xl',
    '2xl': 'w-full max-w-6xl',
  }
  return map[props.size] || map.md
})

function open() {
  dialogRef.value?.showModal()
}

function close() {
  dialogRef.value?.close()
}

defineExpose({ open, close })
</script>
