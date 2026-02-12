<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Cron Jobs</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Cron
      </button>
    </div>

    <EmptyState v-if="!store.crons.length" title="No cron jobs configured" subtitle="Schedule prompts to run automatically on agents" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="c in store.crons" :key="c.id">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
              :class="c.enabled ? 'bg-sol-500/15' : 'bg-piedra-800'">
              <Icon name="clock" size="md" :class="c.enabled ? 'text-sol-400' : 'text-arena-500'" />
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-1.5">
                <h3 class="font-medium text-arena-100 text-sm">{{ c.name }}</h3>
                <Badge v-if="!c.enabled" variant="muted">paused</Badge>
              </div>
              <p class="text-[10px] text-arena-500 font-mono">{{ c.schedule }}</p>
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
        <p v-if="c.description" class="text-[10px] text-arena-400 mb-2">{{ c.description }}</p>
        <Badge variant="sol">{{ store.agentLabel(c.agentId) }}</Badge>
        <p class="text-[10px] text-arena-500 mt-2 line-clamp-2 italic">"{{ c.prompt }}"</p>
      </Card>
    </div>

    <CronDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { cronsApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import CronDialog from './CronDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')

function openDialog(cron = null) {
  dialog.value?.open(cron)
}

function handleDelete(c) {
  requestDelete(`Delete cron job "${c.name}"? This cannot be undone.`, async () => {
    try {
      await cronsApi.delete(c.id)
      await store.refresh()
    } catch (e) {
      alert('Error: ' + e.message)
    }
  })
}
</script>
