<template>
  <div
    class="bg-piedra-900 border rounded-xl p-4 transition-all duration-200"
    :class="active
      ? 'border-piedra-700/50 hover:border-green-500/15 hover:shadow-[0_0_15px_-3px_rgba(74,222,128,0.04)]'
      : 'border-piedra-700/50 hover:border-piedra-600/50'"
  >
    <div class="flex items-start gap-3">
      <!-- Radio toggle -->
      <button
        @click="$emit('activate')"
        class="mt-0.5 flex-shrink-0 group/radio cursor-pointer"
        :title="active ? 'Deactivate' : 'Set as active'"
      >
        <div
          class="w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all duration-200"
          :class="active
            ? 'border-green-400 bg-green-400/10'
            : 'border-piedra-600 bg-piedra-800 group-hover/radio:border-arena-500'"
        >
          <div
            class="w-2.5 h-2.5 rounded-full transition-all duration-200"
            :class="active ? 'bg-green-400 scale-100' : 'bg-transparent scale-0 group-hover/radio:scale-75 group-hover/radio:bg-arena-600'"
          />
        </div>
      </button>

      <!-- Content -->
      <div class="flex-1 min-w-0">
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
            <button @click="testHealth" class="p-1.5 hover:bg-piedra-800 rounded-lg cursor-pointer" title="Test Connection">
              <Icon name="bolt" size="sm" class="text-arena-400" />
            </button>
            <button @click="$emit('edit')" class="p-1.5 hover:bg-piedra-800 rounded-lg cursor-pointer" title="Edit">
              <Icon name="edit" size="sm" class="text-arena-400" />
            </button>
            <button @click="$emit('delete')" class="p-1.5 hover:bg-piedra-800 rounded-lg cursor-pointer" title="Delete">
              <Icon name="trash" size="sm" class="text-arena-400 hover:text-lava-400" />
            </button>
          </div>
        </div>
        <p v-if="provider.embedding?.backend" class="text-[10px] text-arena-400 mb-2">
          Embedding: {{ store.backendLabel(provider.embedding.backend) }} / {{ provider.embedding.model || '?' }}
        </p>
        <p v-if="provider.config?.ttl" class="text-[10px] text-arena-400 mb-2">TTL: {{ provider.config.ttl }}</p>
        <div class="flex flex-wrap gap-1">
          <Badge variant="muted">{{ categoryLabel }}</Badge>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { memoryApi } from '../../lib/api/index.js'
import Badge from '../../components/Badge.vue'
import Icon from '../../components/Icon.vue'

const props = defineProps({
  provider: { type: Object, required: true },
  active: { type: Boolean, default: false },
})

defineEmits(['edit', 'delete', 'activate'])

const store = useDataStore()
const health = ref(null)
const healthLoading = ref(false)

const displayName = computed(() => {
  const t = store.memoryTypes.find(t => t.type === props.provider.type)
  return t?.displayName || props.provider.type
})
const subtitle = computed(() => props.provider.config?.connectionString || 'not configured')
const categoryLabel = computed(() => props.provider.category === 'session' ? 'Session' : 'Long-term')

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
