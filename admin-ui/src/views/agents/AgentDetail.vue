<template>
  <div class="border-t border-piedra-700/30 p-4 grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
    <div class="space-y-4">
      <p v-if="agent.description" class="text-xs text-arena-400">{{ agent.description }}</p>

      <div class="space-y-1.5">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">LLM</h4>
        <DetailRow label="Backend" :value="store.backendLabel(agent.llm?.backend)" />
        <DetailRow label="Model" :value="agent.llm?.model" />
      </div>

      <div class="space-y-1.5">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Transcription (STT)</h4>
        <template v-if="agent.transcription?.backend">
          <DetailRow label="Backend" :value="store.backendLabel(agent.transcription.backend)" />
          <DetailRow label="Model" :value="agent.transcription.model" />
        </template>
        <p v-else class="text-[11px] text-arena-600">Disabled</p>
      </div>

      <div class="space-y-1.5">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">TTS</h4>
        <template v-if="agent.tts?.backend">
          <DetailRow label="Backend" :value="store.backendLabel(agent.tts.backend)" />
          <DetailRow label="Model" :value="agent.tts.model" />
          <DetailRow label="Voice" :value="agent.tts.voice" />
          <DetailRow v-if="agent.tts.speed" label="Speed" :value="agent.tts.speed + 'x'" />
        </template>
        <p v-else class="text-[11px] text-arena-600">Disabled</p>
      </div>

      <div v-if="agent.systemPrompt" class="space-y-1">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">System Prompt</h4>
        <p class="text-[11px] text-arena-300 whitespace-pre-wrap bg-piedra-800/50 rounded-lg p-2 max-h-28 overflow-y-auto">{{ agent.systemPrompt }}</p>
      </div>
    </div>

    <div class="space-y-4">
      <div class="space-y-1.5">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">MCP Servers</h4>
        <div v-if="mcpIds.length" class="flex flex-wrap gap-1.5">
          <Badge variant="atlantico" v-for="id in mcpIds" :key="id">{{ mcpName(id) }}</Badge>
        </div>
        <p v-else class="text-[11px] text-arena-600">None linked</p>
      </div>

      <div class="space-y-1.5">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Memory</h4>
        <DetailRow label="Session" :value="store.memoryLabel(agent.memory?.session) || 'Not configured'" />
        <DetailRow label="Long-term" :value="store.memoryLabel(agent.memory?.longTerm) || 'Not configured'" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import Badge from '../../components/Badge.vue'
import DetailRow from '../../components/DetailRow.vue'

const props = defineProps({ agent: { type: Object, required: true } })
const store = useDataStore()

const mcpIds = computed(() => props.agent.mcpServers || [])

function mcpName(id) {
  const m = store.mcps.find(m => m.id === id)
  return m?.name || id
}
</script>
