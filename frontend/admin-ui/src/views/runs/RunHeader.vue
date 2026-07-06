<template>
  <div class="flex items-start gap-3 pt-2">
    <button
      @click="$emit('back')"
      class="w-7 h-7 rounded-lg flex items-center justify-center text-arena-400 hover:text-arena-200 hover:bg-piedra-800/80 transition-colors flex-shrink-0"
    >
      <Icon name="back" size="sm" />
    </button>

    <!-- App icon, flow or agent, mirroring the list rows -->
    <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0" :class="appKind === 'flow' ? 'bg-rose-500/15' : 'bg-sol-500/15'">
      <Icon :name="appKind === 'flow' ? 'flow' : 'users'" size="sm" :class="appKind === 'flow' ? 'text-rose-400' : 'text-sol-400'" />
    </div>

    <div class="flex-1 min-w-0 space-y-2">
      <div class="flex items-center gap-2">
        <h2 class="text-sm font-semibold text-arena-200 truncate">{{ appName }}</h2>
      </div>

      <!-- Run facts and client metadata: a label column and a pills column,
           so wrapped pills stay aligned under their own column. -->
      <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 items-baseline">
        <template v-if="run?.status">
          <span class="text-[9px] font-semibold text-arena-600 uppercase tracking-wider">Status</span>
          <div class="flex flex-wrap items-center gap-1.5 min-w-0">
            <StatusPill :color="STATUS_COLORS[run.status] || 'arena'">
              {{ STATUS_TEXT[run.status] || run.status }}
            </StatusPill>
          </div>
        </template>
        <template v-if="metaPills.length">
          <span class="text-[9px] font-semibold text-arena-600 uppercase tracking-wider">Client</span>
          <div class="flex flex-wrap items-center gap-1.5 min-w-0">
            <Badge v-for="pill in metaPills" :key="pill.key" variant="muted">
              <span class="text-arena-600">{{ pill.key }}</span> {{ pill.value }}
            </Badge>
          </div>
        </template>
        <span class="text-[9px] font-semibold text-arena-600 uppercase tracking-wider">Run</span>
        <div class="flex flex-wrap items-center gap-1.5 min-w-0">
          <Badge v-for="pill in factPills" :key="pill.key" variant="muted">
            <span class="text-arena-600">{{ pill.key }}</span> {{ pill.value }}
          </Badge>
        </div>
      </div>
    </div>

    <button
      @click="$emit('delete')"
      class="p-1.5 hover:bg-lava-500/10 rounded-lg transition-colors group/btn flex-shrink-0 self-start"
      title="Delete run"
    >
      <Icon name="trash" size="sm" class="text-arena-500 group-hover/btn:text-lava-400 transition-colors" />
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Badge from '../../components/Badge.vue'
import StatusPill from '../../components/StatusPill.vue'
import Icon from '../../components/Icon.vue'

const props = defineProps({
  run:       { type: Object, default: null },
  appName:   { type: String, default: '' },
  appKind:   { type: String, default: 'agent' },
  metaPills: { type: Array, default: () => [] },
})

defineEmits(['back', 'delete'])

const STATUS_COLORS = {
  completed: 'green',
  failed: 'lava',
  interrupted: 'arena'
}

const STATUS_TEXT = {
  completed: 'Completed',
  failed: 'Failed',
  interrupted: 'Interrupted'
}

// factPills carries the run's own facts; client metadata arrives separately
// through the metaPills prop and renders after a divider.
const factPills = computed(() => {
  const r = props.run
  if (!r) return []
  const facts = []
  if (r.runId) facts.push({ key: 'id', value: r.runId })
  if (r.sessionId) facts.push({ key: 'session', value: r.sessionId })
  if (r.startedAt) facts.push({ key: 'started', value: new Date(r.startedAt).toLocaleString() })
  if (r.startedAt) facts.push({ key: 'duration', value: formatDuration(r.startedAt, r.endedAt) })
  if (r.source) facts.push({ key: 'source', value: r.source })
  return facts
})

function formatDuration(startedAt, endedAt) {
  const start = new Date(startedAt).getTime()
  const end = endedAt ? new Date(endedAt).getTime() : Date.now()
  const ms = end - start
  if (ms < 0) return '0s'
  if (ms < 1000) return '<1s'
  const secs = Math.floor(ms / 1000)
  if (secs < 60) return `${secs}s`
  const mins = Math.floor(secs / 60)
  const remSecs = secs % 60
  return remSecs === 0 ? `${mins}m` : `${mins}m ${remSecs}s`
}
</script>
