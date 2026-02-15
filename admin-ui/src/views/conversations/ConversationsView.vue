<template>
  <Transition name="section" mode="out-in">
    <ConversationDetail
      v-if="selectedId"
      :key="selectedId"
      :conversationId="selectedId"
      @back="selectedId = ''"
      @deleted="selectedId = ''"
      @navigate="selectedId = $event"
    />
    <ConversationsList
      v-else
      ref="listRef"
      @select="selectedId = $event"
    />
  </Transition>
</template>

<script setup>
import { ref } from 'vue'
import ConversationsList from './ConversationsList.vue'
import ConversationDetail from './ConversationDetail.vue'

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
