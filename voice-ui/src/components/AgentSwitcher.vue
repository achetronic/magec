<template>
  <div ref="root" class="relative">
    <button
      class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors"
      @click.stop="open = !open"
    >
      <svg class="w-5 h-5 text-arena-400 hover:text-arena-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
      </svg>
    </button>

    <Transition name="dropdown">
      <div
        v-if="open"
        class="absolute right-0 top-full mt-1.5 w-60 bg-piedra-900 border border-piedra-700/50 rounded-xl shadow-2xl overflow-hidden z-50"
      >
        <div class="p-2 border-b border-piedra-700/30">
          <p class="text-[10px] text-arena-500 font-medium uppercase tracking-wider px-2">{{ t('agentSwitcher.agent') }}</p>
        </div>
        <div class="p-1.5 max-h-80 overflow-y-auto">
          <template v-for="agent in store.allowedAgents" :key="agent.id">
            <button
              class="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-left transition-colors"
              :class="agent.id === store.selectedAgent ? 'bg-sol-500/10' : 'hover:bg-piedra-800'"
              @click="onSelect(agent)"
            >
              <div
                class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0"
                :class="agent.id === store.selectedAgent ? 'bg-sol-500/20' : 'bg-piedra-800'"
              >
                <span
                  class="text-[10px] font-bold"
                  :class="agent.id === store.selectedAgent ? 'text-sol-400' : 'text-arena-500'"
                >
                  {{ (agent.name || agent.id).charAt(0).toUpperCase() }}
                </span>
              </div>
              <div class="min-w-0 flex-1">
                <p
                  class="text-xs truncate"
                  :class="agent.id === store.selectedAgent ? 'text-arena-100 font-medium' : 'text-arena-300'"
                >
                  {{ agent.name || agent.id }}
                </p>
                <p v-if="agent.type === 'flow'" class="text-[9px] text-arena-500 leading-tight mt-0.5">
                  {{ (agent.agents || []).length }} {{ t('agentSwitcher.agents') }} · {{ spokespersons(agent).length }} {{ t('agentSwitcher.spokespersons') }}
                </p>
              </div>
              <svg
                v-if="agent.id === store.selectedAgent"
                class="w-3.5 h-3.5 text-sol-400 flex-shrink-0"
                fill="none" stroke="currentColor" viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
            </button>

            <Transition name="spokesperson">
              <div
                v-if="agent.id === store.selectedAgent && agent.type === 'flow' && spokespersons(agent).length > 0"
                class="ml-5 my-1 pl-3 border-l-2 border-atlantico-500/20 space-y-px"
              >
                <p class="text-[9px] text-atlantico-400/70 font-medium uppercase tracking-wider px-2 py-1">{{ t('agentSwitcher.spokesperson') }}</p>
                <button
                  v-for="sub in spokespersons(agent)"
                  :key="sub.id"
                  class="w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-left transition-all duration-150"
                  :class="sub.id === store.spokesperson
                    ? 'bg-atlantico-500/10 shadow-[inset_0_0_0_1px_rgba(56,188,216,0.15)]'
                    : 'hover:bg-piedra-800/80'"
                  @click.stop="onSelectSpokesperson(sub.id)"
                >
                  <div
                    class="w-5 h-5 rounded flex items-center justify-center flex-shrink-0 transition-colors"
                    :class="sub.id === store.spokesperson ? 'bg-atlantico-500/15' : 'bg-piedra-800/60'"
                  >
                    <svg
                      class="w-3 h-3"
                      :class="sub.id === store.spokesperson ? 'text-atlantico-400' : 'text-arena-600'"
                      fill="none" stroke="currentColor" viewBox="0 0 24 24"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"/>
                    </svg>
                  </div>
                  <p
                    class="text-[11px] truncate flex-1 transition-colors"
                    :class="sub.id === store.spokesperson ? 'text-atlantico-300 font-medium' : 'text-arena-400'"
                  >
                    {{ sub.name || sub.id }}
                  </p>
                  <div
                    class="w-3 h-3 rounded-full border flex-shrink-0 flex items-center justify-center transition-all"
                    :class="sub.id === store.spokesperson
                      ? 'border-atlantico-400 bg-atlantico-400'
                      : 'border-piedra-600'"
                  >
                    <div
                      v-if="sub.id === store.spokesperson"
                      class="w-1.5 h-1.5 rounded-full bg-piedra-900"
                    />
                  </div>
                </button>
              </div>
            </Transition>
          </template>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '../lib/stores/app.js'
import { t } from '../lib/i18n/index.js'

const store = useAppStore()
const root = ref(null)
const open = ref(false)

watch(() => store.spokespersonPanelOpen, (val) => {
  if (val) {
    open.value = true
    store.spokespersonPanelOpen = false
  }
})

function spokespersons(agent) {
  if (!agent.agents) return []
  const resp = agent.agents.filter(a => a.responseAgent)
  return resp.length > 0 ? resp : agent.agents
}

function onSelect(agent) {
  if (agent.id !== store.selectedAgent) {
    store.switchAgent(agent.id)
  } else if (agent.type !== 'flow') {
    open.value = false
  }
}

function onSelectSpokesperson(agentId) {
  store.switchSpokesperson(agentId)
  open.value = false
}

function onClickOutside(e) {
  if (root.value && !root.value.contains(e.target)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.dropdown-enter-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.dropdown-leave-active {
  transition: opacity 0.1s ease, transform 0.1s ease;
}
.dropdown-enter-from {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}

.spokesperson-enter-active {
  transition: opacity 0.2s ease, max-height 0.25s ease;
  max-height: 300px;
  overflow: hidden;
}
.spokesperson-leave-active {
  transition: opacity 0.15s ease, max-height 0.2s ease;
  overflow: hidden;
}
.spokesperson-enter-from {
  opacity: 0;
  max-height: 0;
}
.spokesperson-leave-to {
  opacity: 0;
  max-height: 0;
}
</style>
