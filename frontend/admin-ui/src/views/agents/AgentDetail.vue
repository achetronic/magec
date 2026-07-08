<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <div class="border-t border-piedra-700/30 p-4 mt-3 space-y-3">
    <p v-if="agent.description" class="text-xs text-arena-400 leading-relaxed">{{ agent.description }}</p>

    <div v-if="(agent.tags || []).length" class="flex flex-wrap gap-1.5">
      <Badge variant="muted" v-for="tag in agent.tags" :key="tag">{{ tag }}</Badge>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-6 gap-y-4">
      <div class="space-y-1.5">
        <h4 class="text-[9px] font-medium text-arena-600 uppercase tracking-widest mb-2">LLM</h4>
        <DetailRow label="Backend" :value="store.backendLabel(agent.llm?.backend)" />
        <DetailRow label="Model" :value="agent.llm?.model" />
        <DetailRow label="Context Guard" :value="contextGuardSummary" />
        <DetailRow label="A2A" :value="a2aSummary" />
      </div>

      <div class="space-y-1.5 lg:border-l lg:border-piedra-700/20 lg:pl-6">
        <h4 class="text-[9px] font-medium text-arena-600 uppercase tracking-widest mb-2">Voice</h4>
        <DetailRow label="STT" :value="sttSummary" />
        <DetailRow label="TTS" :value="ttsSummary" />
        <DetailRow v-if="agent.tts?.voice" label="Voice" :value="agent.tts.voice" />
      </div>

      <div class="space-y-1.5 lg:border-l lg:border-piedra-700/20 lg:pl-6">
        <h4 class="text-[9px] font-medium text-arena-600 uppercase tracking-widest mb-2">MCP Servers</h4>
        <div v-if="mcpIds.length" class="flex flex-wrap gap-1.5">
          <Badge variant="muted" v-for="id in mcpIds" :key="id">{{ mcpName(id) }}</Badge>
        </div>
        <p v-else class="text-[11px] text-arena-600">None linked</p>

        <template v-if="skillIds.length">
          <h4 class="text-[9px] font-medium text-arena-600 uppercase tracking-widest mb-2 mt-4">Skills</h4>
          <div class="flex flex-wrap gap-1.5">
            <Badge variant="muted" v-for="id in skillIds" :key="id">{{ skillName(id) }}</Badge>
          </div>
        </template>
      </div>
    </div>

    <div v-if="agent.systemPrompt" class="space-y-1">
      <div class="flex items-center justify-between">
        <h4 class="text-[9px] font-medium text-arena-600 uppercase tracking-widest">System Prompt</h4>
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
const skillIds = computed(() => props.agent.skills || [])
const promptExpanded = ref(false)
const isPromptLong = computed(() => (props.agent.systemPrompt || '').length > 200)

const contextGuardSummary = computed(() => {
  const cg = props.agent.contextGuard
  if (!cg?.enabled) return 'Disabled'
  const labels = { threshold: 'Token threshold', sliding_window: 'Sliding window' }
  const strategy = labels[cg.strategy] || cg.strategy || 'Token threshold'
  let detail = strategy
  if (cg.strategy === 'sliding_window' && cg.maxTurns) {
    detail = `${strategy} (${cg.maxTurns} turns)`
  } else if (cg.strategy === 'threshold' && cg.maxTokens) {
    detail = `${strategy} (${cg.maxTokens.toLocaleString()} tokens)`
  }
  return `${detail} (experimental)`
})

const a2aSummary = computed(() => {
  return props.agent.a2a?.enabled ? 'Enabled' : 'Disabled'
})

const sttSummary = computed(() => {
  const t = props.agent.transcription
  if (!t?.backend) return 'Disabled'
  return [store.backendLabel(t.backend), t.model].filter(Boolean).join(' / ')
})

const ttsSummary = computed(() => {
  const t = props.agent.tts
  if (!t?.backend) return 'Disabled'
  return [store.backendLabel(t.backend), t.model].filter(Boolean).join(' / ')
})

function mcpName(id) {
  const m = store.mcps.find(m => m.id === id)
  return m?.name || id
}

function skillName(id) {
  const sk = store.skills.find(sk => sk.id === id)
  return sk?.name || id
}

function formatPrompt(text) {
  const esc = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  return esc
    .replace(/\*\*(.+?)\*\*/g, '<strong class="text-arena-200">$1</strong>')
    .replace(/^- (.+)$/gm, '<span class="text-arena-500">•</span> $1')
    .replace(/^(\d+)\. (.+)$/gm, '<span class="text-arena-500">$1.</span> $2')
}
</script>
