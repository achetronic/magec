<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <Transition name="section" mode="out-in">
    <RunDetail
      v-if="selectedId"
      :key="selectedId"
      :runId="selectedId"
      @back="selectedId = ''"
    />
    <RunsList
      v-else
      ref="listRef"
      @select="selectedId = $event"
    />
  </Transition>
</template>

<script setup>
import { ref } from 'vue'
import RunsList from './RunsList.vue'
import RunDetail from './RunDetail.vue'

const selectedId = ref('')
const listRef = ref(null)

defineExpose({
  refresh() {
    if (selectedId.value) {
      selectedId.value = ''
    }
    listRef.value?.refresh()
  }
})
</script>
