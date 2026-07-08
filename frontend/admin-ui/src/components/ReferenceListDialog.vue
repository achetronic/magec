<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <dialog
    ref="dialogRef"
    class="bg-piedra-900 border border-piedra-700/50 rounded-2xl p-0 w-full max-w-md text-arena-100 shadow-2xl"
    @close="$emit('close')"
  >
    <div class="p-5 space-y-4">
      <div>
        <p class="text-sm font-semibold text-arena-100">{{ title }}</p>
        <p class="text-xs text-arena-500 mt-1">{{ subtitle }}</p>
      </div>

      <div class="max-h-64 overflow-y-auto space-y-1.5 pr-1">
        <div
          v-for="(row, i) in rows"
          :key="i"
          class="flex items-center gap-2.5 px-3 py-2 rounded-lg bg-piedra-850"
        >
          <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="row.dot" />
          <div class="flex-1 min-w-0">
            <p class="text-xs text-arena-200 truncate">
              <span class="text-arena-500">{{ typeLabel(row.referrerType) }}</span>
              {{ row.referrerName || row.referrerId }}
            </p>
            <p class="text-[10px] font-mono text-arena-500 truncate">{{ row.detail }}</p>
          </div>
          <span v-if="row.trailing" class="text-[10px] flex-shrink-0" :class="row.trailingClass">
            {{ row.trailing }}
          </span>
        </div>
      </div>

      <div class="flex justify-end gap-3">
        <button @click="close" class="px-4 py-2 text-sm text-arena-400 hover:text-arena-200 hover:bg-piedra-800 rounded-lg transition-colors">
          Cancel
        </button>
        <button
          @click="confirm"
          :disabled="busy"
          class="px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
          :class="confirmClass"
        >
          {{ confirmLabel }}
        </button>
      </div>
    </div>
  </dialog>
</template>

<script setup>
import { ref } from 'vue'

// ReferenceListDialog is the shared shell for "these entities are affected"
// confirmations: the force-delete demolition quote and the dead-reference
// cleanup both render a scrollable list of referrer rows over the same
// ConfirmDialog-style frame. Rows are display-ready: {referrerType,
// referrerName, referrerId, detail, dot, trailing?, trailingClass?}.
defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  rows: { type: Array, default: () => [] },
  confirmLabel: { type: String, default: 'Confirm' },
  confirmClass: { type: String, default: 'bg-lava-500 hover:bg-lava-600 text-white' },
  busy: { type: Boolean, default: false },
})

const emit = defineEmits(['confirm', 'close'])
const dialogRef = ref(null)

const TYPE_LABELS = {
  agent: 'Agent',
  client: 'Client',
  flow: 'Flow',
  memoryProvider: 'Memory provider',
  settings: 'Settings',
}

function typeLabel(type) {
  return TYPE_LABELS[type] || type
}

function open() {
  dialogRef.value?.showModal()
}

function close() {
  dialogRef.value?.close()
}

function confirm() {
  emit('confirm')
}

defineExpose({ open, close })
</script>
