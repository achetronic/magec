<template>
  <Teleport to="body">
    <div class="fixed bottom-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-x-4 opacity-0"
      >
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex items-center gap-2.5 px-3.5 py-2.5 rounded-xl border shadow-lg shadow-black/30 min-w-[240px] max-w-[360px]"
          :class="toastClasses(toast.type)"
        >
          <div class="w-5 h-5 flex-shrink-0 flex items-center justify-center">
            <svg v-if="toast.type === 'success'" class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            <svg v-else-if="toast.type === 'error'" class="w-4 h-4 text-lava-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <svg v-else class="w-4 h-4 text-arena-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <p class="text-xs font-medium leading-snug flex-1">{{ toast.message }}</p>
          <button
            @click="dismiss(toast.id)"
            class="w-5 h-5 flex-shrink-0 flex items-center justify-center rounded-md hover:bg-piedra-700/50 transition-colors"
          >
            <svg class="w-3 h-3 text-arena-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'

const toasts = ref([])
let nextId = 0

function toastClasses(type) {
  if (type === 'success') return 'bg-piedra-900 border-green-500/30 text-green-200'
  if (type === 'error') return 'bg-piedra-900 border-lava-500/30 text-lava-200'
  return 'bg-piedra-900 border-piedra-600/50 text-arena-200'
}

function show(message, type = 'info', duration = 3000) {
  const id = ++nextId
  toasts.value.push({ id, message, type })
  if (duration > 0) {
    setTimeout(() => dismiss(id), duration)
  }
}

function dismiss(id) {
  const i = toasts.value.findIndex(t => t.id === id)
  if (i !== -1) toasts.value.splice(i, 1)
}

function success(message) { show(message, 'success') }
function error(message) { show(message, 'error', 5000) }
function info(message) { show(message, 'info') }

defineExpose({ show, success, error, info })
</script>
