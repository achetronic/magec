<template>
  <div class="relative">
    <button
      class="p-1.5 sm:p-2 hover:bg-piedra-800 rounded-lg transition-colors"
      @click.stop="open = !open"
    >
      <svg class="w-5 h-5 text-arena-400 hover:text-arena-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute right-0 top-full mt-1.5 w-52 bg-piedra-900 border border-piedra-700/50 rounded-xl shadow-2xl overflow-hidden z-50"
    >
      <div class="p-2 border-b border-piedra-700/30">
        <p class="text-[10px] text-arena-500 font-medium uppercase tracking-wider px-2">Agente</p>
      </div>
      <div class="p-1.5 space-y-0.5 max-h-60 overflow-y-auto">
        <button
          v-for="agent in store.allowedAgents"
          :key="agent.id"
          class="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-left transition-colors"
          :class="agent.id === store.selectedAgent ? 'bg-sol-500/10' : 'hover:bg-piedra-800'"
          @click="onSelect(agent.id)"
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
          </div>
          <svg
            v-if="agent.id === store.selectedAgent"
            class="w-3.5 h-3.5 text-sol-400 flex-shrink-0"
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '../lib/stores/app.js'

const store = useAppStore()
const open = ref(false)

function onSelect(agentId) {
  store.switchAgent(agentId)
  open.value = false
}

function onClickOutside(e) {
  if (!e.target.closest('.relative')) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onUnmounted(() => document.removeEventListener('click', onClickOutside))
</script>
