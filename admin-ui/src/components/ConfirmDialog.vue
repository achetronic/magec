<template>
  <dialog
    ref="dialogRef"
    class="bg-piedra-900 border border-piedra-700/50 rounded-2xl p-0 w-full max-w-sm text-arena-100 shadow-2xl"
    @close="$emit('close')"
  >
    <div class="p-5 space-y-4">
      <p class="text-sm text-arena-200">{{ message }}</p>
      <div class="flex justify-end gap-3">
        <button @click="close" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
          Cancel
        </button>
        <button @click="confirm" class="px-4 py-2 bg-lava-500 hover:bg-lava-600 text-white text-sm font-medium rounded-lg transition-colors">
          Delete
        </button>
      </div>
    </div>
  </dialog>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  message: { type: String, default: 'Are you sure?' },
})

const emit = defineEmits(['confirm', 'close'])
const dialogRef = ref(null)

function open() {
  dialogRef.value?.showModal()
}

function close() {
  dialogRef.value?.close()
}

function confirm() {
  emit('confirm')
  close()
}

defineExpose({ open, close })
</script>
