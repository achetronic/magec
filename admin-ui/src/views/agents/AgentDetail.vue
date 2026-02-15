<template>
  <div class="border-t border-piedra-700/30 p-4 mt-3 space-y-4">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-4">
        <p v-if="agent.description" class="text-xs text-arena-400">{{ agent.description }}</p>

        <div v-if="(agent.tags || []).length" class="flex flex-wrap gap-1.5">
          <Badge variant="muted" v-for="tag in agent.tags" :key="tag">{{ tag }}</Badge>
        </div>

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
      </div>

      <div class="space-y-4">
        <div class="space-y-1.5">
          <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">MCP Servers</h4>
          <div v-if="mcpIds.length" class="flex flex-wrap gap-1.5">
            <Badge variant="muted" v-for="id in mcpIds" :key="id">{{ mcpName(id) }}</Badge>
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

    <div v-if="agent.systemPrompt" class="space-y-1">
      <div class="flex items-center justify-between">
        <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">System Prompt</h4>
        <button
          v-if="isPromptLong"
          @click="promptExpanded = !promptExpanded"
          class="text-[10px] text-sol-400 hover:text-sol-300 transition-colors cursor-pointer"
        >{{ promptExpanded ? 'Collapse' : 'Expand' }}</button>
      </div>
      <div
        class="text-[11px] text-arena-300 leading-relaxed bg-piedra-800/50 rounded-lg p-3 overflow-hidden transition-[max-height] duration-300 ease-in-out"
        :class="promptExpanded ? 'max-h-[80vh]' : 'max-h-32'"
      >
        <div class="whitespace-pre-wrap" v-html="formatPrompt(agent.systemPrompt)" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import Badge from '../../components/Badge.vue'
import DetailRow from '../../components/DetailRow.vue'

const props = defineProps({ agent: { type: Object, required: true } })
const store = useDataStore()

const mcpIds = computed(() => props.agent.mcpServers || [])
const promptExpanded = ref(false)
const isPromptLong = computed(() => (props.agent.systemPrompt || '').length > 200)

function mcpName(id) {
  const m = store.mcps.find(m => m.id === id)
  return m?.name || id
}

function formatPrompt(text) {
  const esc = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  return esc
    .replace(/\*\*(.+?)\*\*/g, '<strong class="text-arena-200">$1</strong>')
    .replace(/^- (.+)$/gm, '<span class="text-arena-500">•</span> $1')
    .replace(/^(\d+)\. (.+)$/gm, '<span class="text-arena-500">$1.</span> $2')
}
</script>
