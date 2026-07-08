<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <label class="block text-xs text-arena-400 mb-1">
    {{ label }}
    <span v-if="required" class="text-lava-400">*</span>
  </label>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  required: { type: Boolean, default: false },
})
</script>
