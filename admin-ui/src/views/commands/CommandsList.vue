<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Commands</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Command
      </button>
    </div>

    <EmptyState v-if="!store.commands.length" title="No commands configured" subtitle="Create reusable prompts that triggers can invoke against agents" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="c in store.commands" :key="c.id">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 bg-indigo-500/15">
              <Icon name="command" size="md" class="text-indigo-400" />
            </div>
            <div class="min-w-0">
              <h3 class="font-medium text-arena-100 text-sm">{{ c.name }}</h3>
              <p v-if="c.description" class="text-[10px] text-arena-500 truncate">{{ c.description }}</p>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(c)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(c)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p class="text-[10px] text-arena-500 line-clamp-2 italic">"{{ c.prompt }}"</p>
      </Card>
    </div>

    <CommandDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { commandsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import CommandDialog from './CommandDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')

function openDialog(cmd = null) {
  dialog.value?.open(cmd)
}

function handleDelete(c) {
  requestDelete(`Delete command "${c.name}"? This cannot be undone.`, async () => {
    try {
      await commandsApi.delete(c.id)
      await store.refresh()
    } catch (e) {
      alert('Error: ' + e.message)
    }
  })
}
</script>
