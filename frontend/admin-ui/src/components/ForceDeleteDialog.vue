<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <ReferenceListDialog
    ref="listRef"
    :title="`&quot;${label}&quot; is still in use`"
    subtitle="Deleting it will also update everything below. Review the damage before forcing."
    :rows="rows"
    confirm-label="Force delete"
    confirm-class="bg-lava-500 hover:bg-lava-600 text-white"
    @confirm="onConfirm"
    @close="$emit('close')"
  />
</template>

<script setup>
import { computed, ref } from 'vue'
import ReferenceListDialog from './ReferenceListDialog.vue'

const props = defineProps({
  label: { type: String, default: '' },
  references: { type: Array, default: () => [] },
})

const emit = defineEmits(['confirm', 'close'])
const listRef = ref(null)

// Structural references get the lava treatment: forcing the delete breaks or
// mutilates the referrer. Membership references are harmlessly unlinked.
const rows = computed(() => props.references.map((ref) => ({
  referrerType: ref.referrerType,
  referrerName: ref.referrerName,
  referrerId: ref.referrerId,
  detail: ref.field,
  dot: ref.kind === 'structural' ? 'bg-lava-400' : 'bg-arena-500',
  trailing: actionLabel(ref),
  trailingClass: ref.kind === 'structural' ? 'text-lava-400' : 'text-arena-500',
})))

// actionLabel says what the force delete does to this referrer.
function actionLabel(ref) {
  if (ref.kind !== 'structural') return 'unlinked'
  if (ref.field.startsWith('node ')) return 'node removed'
  return 'breaks it'
}

function onConfirm() {
  emit('confirm')
  listRef.value?.close()
}

function open() {
  listRef.value?.open()
}

function close() {
  listRef.value?.close()
}

defineExpose({ open, close })
</script>
