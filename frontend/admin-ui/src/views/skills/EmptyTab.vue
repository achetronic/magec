<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <!-- EmptyTab — quiet empty-state slug for the skill viewer tabs.
       Centered, generous vertical padding, muted glyph + single line
       of explanatory text. Matches the spirit of EmptyState.vue but
       without the action button (the viewer is read-only) and without
       the entity-color treatment (we are inside the cyan band; an
       extra colour would scream). -->
  <div class="flex flex-col items-center justify-center py-10 text-center">
    <div class="w-10 h-10 rounded-xl bg-piedra-800/60 border border-piedra-700/40 flex items-center justify-center mb-3">
      <Icon :name="icon" size="md" class="text-arena-600" />
    </div>
    <p class="text-[11px] text-arena-500 max-w-xs leading-relaxed">{{ message }}</p>
  </div>
</template>

<script setup>
import Icon from '../../components/Icon.vue'

defineProps({
  icon: { type: String, default: 'skill' },
  message: { type: String, required: true },
})
</script>
