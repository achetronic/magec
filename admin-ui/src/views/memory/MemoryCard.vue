<template>
  <Card>
    <div class="flex items-start justify-between gap-3 mb-2">
      <div class="flex items-center gap-3 min-w-0">
        <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 relative bg-green-500/15">
          <span class="text-[10px] font-mono font-bold text-green-300">
            {{ displayName.substring(0, 3).toUpperCase() }}
          </span>
          <span
            class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border border-piedra-900"
            :class="healthClass"
            :title="healthTitle"
          />
        </div>
        <div class="min-w-0">
          <div class="flex items-center gap-1.5">
            <h3 class="font-medium text-arena-100 text-sm">{{ provider.name }}</h3>
            <Badge variant="muted">{{ displayName }}</Badge>
          </div>
          <p class="text-[10px] text-arena-500 truncate">{{ subtitle }}</p>
        </div>
      </div>
      <div class="flex gap-0.5 flex-shrink-0">
        <button @click="testHealth" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Test Connection">
          <Icon name="bolt" size="sm" class="text-arena-400" />
        </button>
        <button @click="$emit('edit')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Edit">
          <Icon name="edit" size="sm" class="text-arena-400" />
        </button>
        <button @click="$emit('delete')" class="p-1.5 hover:bg-piedra-800 rounded-lg" title="Delete">
          <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
        </button>
      </div>
    </div>
    <p v-if="provider.embedding?.backend" class="text-[10px] text-arena-400 mb-2">
      Embedding: {{ store.backendLabel(provider.embedding.backend) }} / {{ provider.embedding.model || '?' }}
    </p>
    <p v-if="provider.config?.ttl" class="text-[10px] text-arena-400 mb-2">TTL: {{ provider.config.ttl }}</p>
    <div v-if="usedByAgents.length" class="flex flex-wrap gap-1">
      <Badge variant="sol" v-for="name in usedByAgents" :key="name">{{ name }}</Badge>
    </div>
    <p v-else class="text-[10px] text-arena-600">Not used by any agent</p>
  </Card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
import Card from '../../components/Card.vue'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'

const props = defineProps({
  provider: { type: Object, required: true },
})

defineEmits(['edit', 'delete'])

const store = useDataStore()
const health = ref(null)
const healthLoading = ref(false)

const isSession = computed(() => props.provider.category === 'session')
const displayName = computed(() => {
  const t = store.memoryTypes.find(t => t.type === props.provider.type)
  return t?.displayName || props.provider.type
})
const subtitle = computed(() => props.provider.config?.connectionString || 'not configured')

const usedByAgents = computed(() => {
  const names = new Set()
  for (const a of store.agents) {
    const label = a.name || a.id
    if (a.memory?.session === props.provider.id) names.add(label)
    if (a.memory?.longTerm === props.provider.id) names.add(label)
  }
  return [...names]
})

const healthClass = computed(() => {
  if (healthLoading.value) return 'bg-piedra-600 animate-pulse'
  if (health.value === null) return 'bg-piedra-600'
  return health.value.healthy ? 'bg-green-500' : 'bg-lava-500'
})

const healthTitle = computed(() => {
  if (healthLoading.value) return 'Testing...'
  if (health.value === null) return 'Checking...'
  return health.value.detail || (health.value.healthy ? 'Connected' : 'Unreachable')
})

async function testHealth() {
  healthLoading.value = true
  try {
    health.value = await memoryApi.checkHealth(props.provider.id)
  } catch {
    health.value = { healthy: false, detail: 'Health check failed' }
  } finally {
    healthLoading.value = false
  }
}

onMounted(() => testHealth())
</script>
