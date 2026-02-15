<template>
  <Teleport to="body">
    <Transition name="search">
      <div v-if="open" class="fixed inset-0 z-[9998] flex items-start justify-center pt-[15vh]" @mousedown.self="close">
        <div class="absolute inset-0 bg-black/50" />
        <div class="relative w-full max-w-md bg-piedra-900 border border-piedra-700/60 rounded-xl shadow-2xl shadow-black/50 overflow-hidden">
          <div class="flex items-center gap-2.5 px-4 py-3 border-b border-piedra-700/50">
            <svg class="w-4 h-4 text-arena-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              ref="inputRef"
              v-model="query"
              type="text"
              class="flex-1 bg-transparent text-sm text-arena-100 placeholder-arena-500 outline-none"
              placeholder="Search entities..."
              @keydown.escape="close"
              @keydown.down.prevent="moveSelection(1)"
              @keydown.up.prevent="moveSelection(-1)"
              @keydown.enter.prevent="selectCurrent"
            />
            <kbd class="hidden sm:inline-flex px-1.5 py-0.5 text-[9px] font-mono text-arena-500 bg-piedra-800 border border-piedra-700/50 rounded">ESC</kbd>
          </div>

          <div v-if="results.length" class="max-h-[320px] overflow-y-auto py-1.5">
            <button
              v-for="(item, i) in results"
              :key="item.id + item.section"
              @click="go(item)"
              @mouseenter="selectedIndex = i"
              class="flex items-center gap-3 w-full px-4 py-2 text-left transition-colors"
              :class="selectedIndex === i ? 'bg-piedra-800/80' : 'hover:bg-piedra-800/40'"
            >
              <div
                class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0"
                :class="iconBg(item.color)"
              >
                <Icon :name="item.icon" size="xs" :class="iconText(item.color)" />
              </div>
              <div class="min-w-0 flex-1">
                <p class="text-xs font-medium text-arena-100 truncate">{{ item.name }}</p>
                <p class="text-[10px] text-arena-500 truncate">{{ item.sectionLabel }}</p>
              </div>
              <Icon v-if="selectedIndex === i" name="chevronRight" size="xs" class="text-arena-500 flex-shrink-0" />
            </button>
          </div>
          <div v-else-if="query.length > 0" class="px-4 py-8 text-center">
            <p class="text-xs text-arena-500">No results for "{{ query }}"</p>
          </div>
          <div v-else class="px-4 py-6 text-center">
            <p class="text-[10px] text-arena-500">Type to search across all entities</p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useDataStore } from '../lib/stores/data.js'
import Icon from './Icon.vue'

const emit = defineEmits(['navigate'])
const store = useDataStore()

const open = ref(false)
const query = ref('')
const selectedIndex = ref(0)
const inputRef = ref(null)

const sectionDefs = [
  { key: 'backends',  label: 'Backends',     icon: 'server',  color: 'purple',    data: () => store.backends },
  { key: 'memory',    label: 'Memory',       icon: 'database', color: 'green',     data: () => store.memory },
  { key: 'mcps',      label: 'MCP Servers',  icon: 'bolt',     color: 'atlantico', data: () => store.mcps },
  { key: 'agents',    label: 'Agents',       icon: 'users',    color: 'sol',       data: () => store.agents },
  { key: 'flows',     label: 'Flows',        icon: 'flow',     color: 'rose',      data: () => store.flows },
  { key: 'commands',  label: 'Commands',     icon: 'command',  color: 'indigo',    data: () => store.commands },
  { key: 'clients',   label: 'Clients',      icon: 'phone',    color: 'lava',      data: () => store.clients },
  { key: 'conversations', label: 'Conversations', icon: 'chat', color: 'teal', data: () => [] },
]

const results = computed(() => {
  const q = query.value.toLowerCase().trim()
  if (!q) return []
  const items = []
  for (const s of sectionDefs) {
    for (const item of s.data()) {
      const name = item.name || item.id || ''
      if (name.toLowerCase().includes(q) || (item.description || '').toLowerCase().includes(q)) {
        items.push({
          id: item.id,
          name,
          section: s.key,
          sectionLabel: s.label,
          icon: s.icon,
          color: s.color,
        })
      }
    }
    if (items.length >= 20) break
  }
  return items
})

watch(results, () => { selectedIndex.value = 0 })

function show() {
  query.value = ''
  selectedIndex.value = 0
  open.value = true
  nextTick(() => inputRef.value?.focus())
}

function close() {
  open.value = false
}

function moveSelection(delta) {
  if (!results.value.length) return
  selectedIndex.value = (selectedIndex.value + delta + results.value.length) % results.value.length
}

function selectCurrent() {
  if (results.value[selectedIndex.value]) {
    go(results.value[selectedIndex.value])
  }
}

function go(item) {
  emit('navigate', item.section)
  close()
}

function iconBg(color) {
  const map = {
    purple: 'bg-purple-500/15', green: 'bg-green-500/15', atlantico: 'bg-atlantico-500/15',
    sol: 'bg-sol-500/15', rose: 'bg-rose-500/15', indigo: 'bg-indigo-500/15',
    teal: 'bg-teal-500/15', lava: 'bg-lava-500/15',
  }
  return map[color] || map.sol
}

function iconText(color) {
  const map = {
    purple: 'text-purple-400', green: 'text-green-400', atlantico: 'text-atlantico-400',
    sol: 'text-sol-400', rose: 'text-rose-400', indigo: 'text-indigo-400',
    teal: 'text-teal-400', lava: 'text-lava-400',
  }
  return map[color] || map.sol
}

function onKeydown(e) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    open.value ? close() : show()
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))

defineExpose({ show })
</script>

<style scoped>
.search-enter-active { transition: opacity 0.15s ease; }
.search-leave-active { transition: opacity 0.1s ease; }
.search-enter-from, .search-leave-to { opacity: 0; }
</style>
