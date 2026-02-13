<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold text-arena-200">Triggers</h2>
      <button @click="openDialog()" class="px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors">
        + New Trigger
      </button>
    </div>

    <SkeletonCard v-if="store.loading && !store.triggers.length" />

    <EmptyState v-else-if="!store.triggers.length" title="No triggers configured" subtitle="Automate command execution with cron schedules or webhooks" icon="trigger" color="teal" actionLabel="+ New Trigger" @action="openDialog()" />

    <div v-else class="grid gap-3 grid-cols-1 sm:grid-cols-2">
      <Card v-for="t in store.triggers" :key="t.id">
        <div class="flex items-start justify-between gap-3 mb-2">
          <div class="flex items-center gap-3 min-w-0">
            <div class="relative w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
              :class="t.enabled ? 'bg-teal-500/15' : 'bg-piedra-800'">
              <Icon name="trigger" size="md" :class="t.enabled ? 'text-teal-400' : 'text-arena-500'" />
              <span class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900"
                :class="t.enabled ? 'bg-green-500' : 'bg-arena-600'"
                :title="t.enabled ? 'Active' : 'Paused'" />
            </div>
            <div class="min-w-0">
              <div class="flex items-center gap-1.5">
                <h3 class="font-medium text-arena-100 text-sm">{{ t.name }}</h3>
                <Badge v-if="!t.enabled" variant="muted">paused</Badge>
              </div>
              <div class="flex items-center gap-1.5 mt-0.5">
                <Badge :variant="t.type === 'cron' ? 'teal' : 'teal'">{{ t.type }}</Badge>
                <span v-if="t.type === 'cron' && t.cron" class="text-[10px] text-arena-500 font-mono">{{ t.cron.schedule }}</span>
                <Badge v-if="t.type === 'webhook' && t.webhook?.passthrough" variant="muted">passthrough</Badge>
              </div>
            </div>
          </div>
          <div class="flex gap-0.5 flex-shrink-0">
            <button @click="openDialog(t)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="handleDelete(t)" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="t.description" class="text-[10px] text-arena-400 mb-2">{{ t.description }}</p>
        <div class="flex flex-wrap gap-1.5">
          <Tooltip v-if="t.commandId" :text="commandTooltip(t.commandId)">
            <Badge variant="indigo">{{ store.commandLabel(t.commandId) }}</Badge>
          </Tooltip>
          <Tooltip v-if="t.agentId" :text="agentTooltip(t.agentId)">
            <Badge variant="sol">{{ store.agentLabel(t.agentId) }}</Badge>
          </Tooltip>
        </div>
      </Card>
    </div>

    <TriggerDialog ref="dialog" @saved="store.refresh()" />
  </div>
</template>

<script setup>
import { inject, ref, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { triggersApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Tooltip from '../../components/Tooltip.vue'
import Icon from '../../components/Icon.vue'
import EmptyState from '../../components/EmptyState.vue'
import SkeletonCard from '../../components/SkeletonCard.vue'
import TriggerDialog from './TriggerDialog.vue'

const store = useDataStore()
const dialog = ref(null)
const requestDelete = inject('requestDelete')
const toast = inject('toast')
const registerNew = inject('registerNew')
onMounted(() => registerNew(() => openDialog()))
onUnmounted(() => registerNew(null))

function commandTooltip(id) {
  const c = store.commands.find(c => c.id === id)
  if (!c?.prompt) return ''
  return c.prompt.slice(0, 100) + (c.prompt.length > 100 ? '...' : '')
}

function agentTooltip(id) {
  const a = store.agents.find(a => a.id === id)
  if (!a?.systemPrompt) return a?.description || ''
  return a.systemPrompt.slice(0, 100) + (a.systemPrompt.length > 100 ? '...' : '')
}

function openDialog(trigger = null) {
  dialog.value?.open(trigger)
}

function handleDelete(t) {
  requestDelete(`Delete trigger "${t.name}"? This cannot be undone.`, async () => {
    try {
      await triggersApi.delete(t.id)
      await store.refresh()
    } catch (e) {
      toast.error(e.message)
    }
  })
}
</script>
