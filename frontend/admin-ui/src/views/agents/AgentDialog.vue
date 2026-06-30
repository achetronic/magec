<!-- SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com> -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Agent' : 'New Agent'" size="lg" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="My Agent" :required="true" />
      </div>
      <div>
        <FormLabel label="Description" />
        <FormInput v-model="form.description" placeholder="What this agent does..." />
      </div>
      <div>
        <FormLabel label="Tags" />
        <div class="flex flex-wrap gap-1.5 mb-2" v-if="form.tags.length">
          <span
            v-for="(tag, i) in form.tags" :key="i"
            class="inline-flex items-center gap-1 px-2 py-0.5 text-[11px] font-medium rounded-lg bg-sol-500/10 text-sol-300 border border-sol-500/20"
          >
            {{ tag }}
            <button type="button" @click="removeTag(i)" class="hover:text-lava-400 transition-colors cursor-pointer">&times;</button>
          </span>
        </div>
        <FormInput v-model="tagInput" placeholder="Type a tag and press Enter" @keydown.enter.prevent="addTag" />
      </div>
      <!-- System Prompt -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>System Prompt</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 space-y-3">
          <textarea v-model="form.systemPrompt" rows="3" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" placeholder="Custom system prompt..." />
          <div>
            <FormLabel label="Output Key (optional)" />
            <FormInput v-model="form.outputKey" placeholder="e.g. analysis_result" />
            <p class="text-[10px] text-arena-500 mt-1">Saves this agent's final output under the given key. Other agents can reference it with <code class="text-arena-300 bg-piedra-800 px-0.5 rounded">&#123;&#123;agent.output:key_name&#125;&#125;</code> in their system prompt.</p>
          </div>
        </div>
      </details>

      <!-- LLM -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>LLM</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 space-y-4">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <FormLabel label="Backend" />
              <FormSelect v-model="form.llmBackend">
                <option v-for="b in store.backends" :key="b.id" :value="b.id">{{ b.name }} ({{ b.type }})</option>
              </FormSelect>
            </div>
            <div>
              <FormLabel label="Model" />
              <FormInput v-model="form.llmModel" placeholder="qwen3:8b" />
            </div>
          </div>

          <!-- LLM Headers -->
          <div>
            <FormLabel label="Headers" />
            <div class="space-y-2">
              <div v-for="(h, i) in form.llmHeaders" :key="i" class="flex gap-2 items-center">
                <input
                  v-model="h.key"
                  placeholder="anthropic-beta"
                  class="flex-1 bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-1.5 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
                />
                <input
                  v-model="h.value"
                  placeholder="context-1m-2025-08-07"
                  class="flex-[2] bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-1.5 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none"
                />
                <button @click="form.llmHeaders.splice(i, 1)" class="p-1.5 hover:bg-piedra-800 rounded-lg text-arena-400 hover:text-lava-400 flex-shrink-0" title="Remove header">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M18 6L6 18M6 6l12 12"/></svg>
                </button>
              </div>
              <button @click="form.llmHeaders.push({ key: '', value: '' })" class="text-xs text-sol-400 hover:text-sol-500 transition-colors">
                + Add header
              </button>
            </div>
            <p class="text-[10px] text-arena-500 mt-1">Extra HTTP headers for this agent's LLM requests. Override backend-level headers.</p>
          </div>

          <!-- Context Guard -->
          <div class="border-t border-piedra-700/30 pt-3">
            <div class="flex items-center justify-between">
              <div>
                <span class="text-xs font-medium text-arena-400">Context Guard <span class="ml-1 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider rounded bg-sol-500/15 text-sol-400 border border-sol-500/20">Experimental</span></span>
                <p class="text-[10px] text-arena-500 mt-0.5">Automatically summarize history to prevent context overflow</p>
              </div>
              <FormToggle v-model="form.contextGuardEnabled" />
            </div>

            <div v-if="form.contextGuardEnabled" class="mt-3 grid grid-cols-2 gap-3">
              <div>
                <FormLabel label="Strategy" />
                <FormSelect v-model="form.contextGuardStrategy">
                  <option value="threshold">Token threshold</option>
                  <option value="sliding_window">Sliding window</option>
                </FormSelect>
                <p class="text-[10px] text-arena-500 mt-1" v-if="form.contextGuardStrategy === 'threshold'">Summarizes when token usage approaches the model's context window limit</p>
                <p class="text-[10px] text-arena-500 mt-1" v-else>Summarizes when conversation exceeds a fixed number of messages</p>
              </div>
              <div v-if="form.contextGuardStrategy === 'sliding_window'">
                <FormLabel label="Max turns" />
                <FormInput v-model="form.contextGuardMaxTurns" type="number" placeholder="20" />
                <p class="text-[10px] text-arena-500 mt-1">Number of messages to keep before summarizing older ones</p>
              </div>
              <div v-if="form.contextGuardStrategy === 'threshold'">
                <FormLabel label="Max tokens" />
                <FormInput v-model="form.contextGuardMaxTokens" type="number" placeholder="Auto (model limit)" />
                <p class="text-[10px] text-arena-500 mt-1">Token limit for triggering summarization. Leave empty to auto-detect from model.</p>
              </div>
            </div>
          </div>
        </div>
      </details>

      <!-- MCPs -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>MCP Servers</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4">
          <div v-if="store.mcps.length" class="flex flex-wrap gap-1.5">
            <button
              v-for="m in store.mcps" :key="m.id"
              type="button"
              @click="toggleMcp(m.id)"
              class="px-2.5 py-1 text-[11px] font-medium rounded-lg border transition-all cursor-pointer"
              :class="form.mcpServers.includes(m.id)
                ? 'bg-atlantico-500/15 text-atlantico-300 border-atlantico-500/30'
                : 'bg-piedra-800 text-arena-500 border-piedra-700/40 hover:border-piedra-600 hover:text-arena-300'"
            >
              {{ m.name }}
            </button>
          </div>
          <p v-else class="text-xs text-arena-500">No MCP servers defined yet</p>
        </div>
      </details>

      <!-- Skills -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>Skills</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4">
          <div v-if="store.skills.length" class="flex flex-wrap gap-1.5">
            <button
              v-for="sk in store.skills" :key="sk.id"
              type="button"
              @click="toggleSkill(sk.id)"
              class="px-2.5 py-1 text-[11px] font-medium rounded-lg border transition-all cursor-pointer"
              :class="form.skills.includes(sk.id)
                ? 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30'
                : 'bg-piedra-800 text-arena-500 border-piedra-700/40 hover:border-piedra-600 hover:text-arena-300'"
            >
              {{ sk.name }}
            </button>
          </div>
          <p v-else class="text-xs text-arena-500">No skills defined yet</p>
        </div>
      </details>

      <!-- A2A -->
      <div class="border border-piedra-700/40 rounded-xl px-4 py-3">
        <div class="flex items-center justify-between">
          <div>
            <span class="text-xs font-medium text-arena-400">A2A Protocol</span>
            <p class="text-[10px] text-arena-500 mt-0.5">Expose this agent via the Agent-to-Agent protocol for external discovery and invocation</p>
          </div>
          <FormToggle v-model="form.a2aEnabled" />
        </div>
      </div>

      <!-- Voice -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>Voice (STT / TTS)</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 space-y-4">
          <div class="space-y-3">
            <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Transcription (STT)</h4>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <FormLabel label="Backend" />
                <FormSelect v-model="form.transcriptionBackend">
                  <option value="">(none)</option>
                  <option v-for="b in sttBackends" :key="b.id" :value="b.id">{{ b.name }} ({{ b.type }})</option>
                </FormSelect>
              </div>
              <div>
                <FormLabel label="Model" />
                <FormInput v-model="form.transcriptionModel" :placeholder="sttModelPlaceholder" />
              </div>
            </div>
            <template v-if="sttExtraProps.length">
              <div class="grid gap-3" :class="sttExtraHasHalf ? 'grid-cols-2' : 'grid-cols-1'">
                <div v-for="{ key, prop } in sttExtraProps" :key="key" :class="prop['x-size'] === 'half' ? '' : 'col-span-full'">
                  <label class="flex items-center gap-1 text-xs text-arena-400 mb-1">
                    {{ prop.title || key }}
                    <a v-if="prop['x-link']" :href="prop['x-link']" target="_blank" class="text-arena-600 hover:text-arena-400 transition-colors" :title="prop.title || key"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" width="13" height="13"><circle cx="8" cy="8" r="6"/><path d="M8 7.5V11M8 5.5V5"/></svg></a>
                  </label>
                  <textarea v-if="prop['x-format'] === 'textarea'" v-model="form.sttProviderConfig[key]" rows="2" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" :placeholder="prop['x-placeholder'] || prop.default || ''" />
                  <FormInput v-else v-model="form.sttProviderConfig[key]" :placeholder="prop['x-placeholder'] || prop.default || ''" :type="prop.type === 'number' ? 'number' : 'text'" />
                  <p v-if="prop.description" class="text-[10px] text-arena-500 mt-1">{{ prop.description }}</p>
                </div>
              </div>
            </template>
          </div>
          <hr class="border-piedra-700/30" />
          <div class="space-y-3">
            <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Text-to-Speech (TTS)</h4>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <FormLabel label="Backend" />
                <FormSelect v-model="form.ttsBackend">
                  <option value="">(none)</option>
                  <option v-for="b in ttsBackends" :key="b.id" :value="b.id">{{ b.name }} ({{ b.type }})</option>
                </FormSelect>
              </div>
              <div>
                <FormLabel label="Model" />
                <FormInput v-model="form.ttsModel" :placeholder="ttsModelPlaceholder" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="flex items-center gap-1 text-xs text-arena-400 mb-1">
                  Voice
                  <a v-if="selectedTtsProviderType === 'gemini'" href="https://cloud.google.com/text-to-speech/docs/gemini-tts#voice_options" target="_blank" class="text-arena-600 hover:text-arena-400 transition-colors" title="Available voices"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" width="13" height="13"><circle cx="8" cy="8" r="6"/><path d="M8 7.5V11M8 5.5V5"/></svg></a>
                </label>
                <FormInput v-model="form.ttsVoice" :placeholder="ttsVoicePlaceholder" />
              </div>
            </div>
            <template v-if="ttsMainProps.length">
              <div class="grid gap-3" :class="ttsMainHasHalf ? 'grid-cols-2' : 'grid-cols-1'">
                <div v-for="{ key, prop } in ttsMainProps" :key="key" :class="prop['x-size'] === 'half' ? '' : 'col-span-full'">
                  <label class="flex items-center gap-1 text-xs text-arena-400 mb-1">
                    {{ prop.title || key }}
                    <a v-if="prop['x-link']" :href="prop['x-link']" target="_blank" class="text-arena-600 hover:text-arena-400 transition-colors" :title="prop.title || key"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" width="13" height="13"><circle cx="8" cy="8" r="6"/><path d="M8 7.5V11M8 5.5V5"/></svg></a>
                  </label>
                  <textarea v-if="prop['x-format'] === 'textarea'" v-model="form.ttsProviderConfig[key]" rows="2" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" :placeholder="prop['x-placeholder'] || prop.default || ''" />
                  <FormInput v-else v-model="form.ttsProviderConfig[key]" :placeholder="prop['x-placeholder'] || prop.default || ''" :type="prop.type === 'number' ? 'number' : 'text'" />
                  <p v-if="prop.description" class="text-[10px] text-arena-500 mt-1">{{ prop.description }}</p>
                </div>
              </div>
            </template>
            <div v-if="ttsAdvancedProps.length" class="border-t border-piedra-700/30 pt-3">
              <div class="flex items-center justify-between mb-3">
                <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Advanced</h4>
                <FormToggle v-model="showTtsAdvanced" />
              </div>
              <div v-if="showTtsAdvanced" class="grid gap-3" :class="ttsAdvancedHasHalf ? 'grid-cols-2' : 'grid-cols-1'">
                <div v-for="{ key, prop } in ttsAdvancedProps" :key="key" :class="prop['x-size'] === 'half' ? '' : 'col-span-full'">
                  <label class="flex items-center gap-1 text-xs text-arena-400 mb-1">
                    {{ prop.title || key }}
                    <a v-if="prop['x-link']" :href="prop['x-link']" target="_blank" class="text-arena-600 hover:text-arena-400 transition-colors" :title="prop.title || key"><svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" width="13" height="13"><circle cx="8" cy="8" r="6"/><path d="M8 7.5V11M8 5.5V5"/></svg></a>
                  </label>
                  <textarea v-if="prop['x-format'] === 'textarea'" v-model="form.ttsProviderConfig[key]" rows="2" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" :placeholder="prop['x-placeholder'] || prop.default || ''" />
                  <FormInput v-else v-model="form.ttsProviderConfig[key]" :placeholder="prop['x-placeholder'] || prop.default || ''" :type="prop.type === 'number' ? 'number' : 'text'" />
                  <p v-if="prop.description" class="text-[10px] text-arena-500 mt-1">{{ prop.description }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </details>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, computed, inject, watch, nextTick } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { agentsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'
import FormToggle from '../../components/FormToggle.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const tagInput = ref('')
const showTtsAdvanced = ref(false)
const isOpening = ref(false)

const form = reactive({
  name: '',
  description: '',
  outputKey: '',
  systemPrompt: '',
  llmBackend: '',
  llmModel: '',
  llmHeaders: [],
  mcpServers: [],
  skills: [],
  tags: [],
  transcriptionBackend: '',
  transcriptionModel: '',
  sttProviderConfig: {},
  ttsBackend: '',
  ttsModel: '',
  ttsVoice: '',
  ttsProviderConfig: {},
  contextGuardEnabled: false,
  contextGuardStrategy: 'threshold',
  contextGuardMaxTurns: '',
  contextGuardMaxTokens: '',
  a2aEnabled: false,
})

function backendType(backendId) {
  if (!backendId) return ''
  const b = store.backends.find(b => b.id === backendId)
  return b?.type || ''
}

const selectedTtsProviderType = computed(() => backendType(form.ttsBackend))
const selectedSttProviderType = computed(() => backendType(form.transcriptionBackend))

function voiceProviderFor(type) {
  return store.voiceTypes.find(v => v.type === type) || null
}

const ttsBackends = computed(() => {
  return store.backends.filter(b => {
    const p = voiceProviderFor(b.type)
    return !p || p.supportsTts
  })
})

const sttBackends = computed(() => {
  return store.backends.filter(b => {
    const p = voiceProviderFor(b.type)
    return !p || p.supportsStt
  })
})

const ttsExtraSchema = computed(() => {
  const p = voiceProviderFor(selectedTtsProviderType.value)
  return p?.ttsConfigSchema || null
})

const sttExtraSchema = computed(() => {
  const p = voiceProviderFor(selectedSttProviderType.value)
  return p?.sttConfigSchema || null
})

function schemaToProps(schema) {
  if (!schema?.properties) return []
  const order = schema.propertyOrder || Object.keys(schema.properties)
  return order
    .filter(key => key in schema.properties)
    .map(key => ({ key, prop: schema.properties[key] }))
}

const ttsExtraProps = computed(() => schemaToProps(ttsExtraSchema.value))
const sttExtraProps = computed(() => schemaToProps(sttExtraSchema.value))
const ttsMainProps = computed(() => ttsExtraProps.value.filter(p => !p.prop['x-advanced']))
const ttsAdvancedProps = computed(() => ttsExtraProps.value.filter(p => p.prop['x-advanced']))
const ttsMainHasHalf = computed(() => ttsMainProps.value.some(p => p.prop['x-size'] === 'half'))
const ttsAdvancedHasHalf = computed(() => ttsAdvancedProps.value.some(p => p.prop['x-size'] === 'half'))
const sttExtraHasHalf = computed(() => sttExtraProps.value.some(p => p.prop['x-size'] === 'half'))

const ttsModelPlaceholder = computed(() => selectedTtsProviderType.value === 'gemini' ? 'gemini-2.5-flash-preview-tts' : 'tts-1')
const ttsVoicePlaceholder = computed(() => selectedTtsProviderType.value === 'gemini' ? 'Kore' : 'alloy')
const sttModelPlaceholder = computed(() => selectedSttProviderType.value === 'gemini' ? 'gemini-2.0-flash' : 'whisper-1')

watch(() => form.ttsBackend, () => { if (!isOpening.value) form.ttsProviderConfig = {} })
watch(() => form.transcriptionBackend, () => { if (!isOpening.value) form.sttProviderConfig = {} })

function headersToList(obj) {
  if (!obj || !Object.keys(obj).length) return []
  return Object.entries(obj).map(([key, value]) => ({ key, value }))
}

function listToHeaders(list) {
  const obj = {}
  for (const h of list) {
    const k = h.key.trim()
    if (k) obj[k] = h.value
  }
  return Object.keys(obj).length ? obj : undefined
}

function toggleMcp(id) {
  const idx = form.mcpServers.indexOf(id)
  if (idx === -1) form.mcpServers.push(id)
  else form.mcpServers.splice(idx, 1)
}

function toggleSkill(id) {
  const idx = form.skills.indexOf(id)
  if (idx === -1) form.skills.push(id)
  else form.skills.splice(idx, 1)
}

function addTag() {
  const tag = tagInput.value.trim().toLowerCase()
  if (tag && !form.tags.includes(tag)) {
    form.tags.push(tag)
  }
  tagInput.value = ''
}

function removeTag(i) {
  form.tags.splice(i, 1)
}

function cleanConfig(cfg) {
  if (!cfg) return undefined
  const clean = {}
  for (const [k, v] of Object.entries(cfg)) {
    if (v !== '' && v !== undefined && v !== null) {
      clean[k] = typeof v === 'string' && !isNaN(Number(v)) && v.includes('.') ? parseFloat(v) : v
    }
  }
  return Object.keys(clean).length ? clean : undefined
}

function buildNamespacedConfig(providerType, providerConfig) {
  if (!providerType) return undefined
  const clean = cleanConfig(providerConfig)
  if (!clean) return undefined
  return { [providerType]: clean }
}

async function open(agent = null) {
  isOpening.value = true
  isEdit.value = !!agent
  editId.value = agent?.id || null
  form.name = agent?.name || ''
  form.description = agent?.description || ''
  form.outputKey = agent?.outputKey || ''
  form.systemPrompt = agent?.systemPrompt || ''
  form.llmBackend = agent?.llm?.backend || ''
  form.llmModel = agent?.llm?.model || ''
  form.llmHeaders = headersToList(agent?.llm?.headers)
  form.mcpServers = [...(agent?.mcpServers || [])]
  form.skills = [...(agent?.skills || [])]
  form.tags = [...(agent?.tags || [])]
  form.transcriptionBackend = agent?.transcription?.backend || ''
  form.transcriptionModel = agent?.transcription?.model || ''
  const sttType = backendType(form.transcriptionBackend)
  form.sttProviderConfig = { ...(agent?.transcription?.config?.[sttType] || {}) }
  form.ttsBackend = agent?.tts?.backend || ''
  form.ttsModel = agent?.tts?.model || ''
  form.ttsVoice = agent?.tts?.voice || ''
  const ttsType = backendType(form.ttsBackend)
  form.ttsProviderConfig = { ...(agent?.tts?.config?.[ttsType] || {}) }
  form.contextGuardEnabled = agent?.contextGuard?.enabled || false
  form.contextGuardStrategy = agent?.contextGuard?.strategy || 'threshold'
  form.contextGuardMaxTurns = agent?.contextGuard?.maxTurns || ''
  form.contextGuardMaxTokens = agent?.contextGuard?.maxTokens || ''
  form.a2aEnabled = agent?.a2a?.enabled || false
  const gcfg = agent?.tts?.config?.gemini || {}
  showTtsAdvanced.value = !!(gcfg.languageCode || gcfg.temperature || gcfg.stylePrompt)
  await nextTick()
  isOpening.value = false
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    outputKey: form.outputKey.trim(),
    systemPrompt: form.systemPrompt.trim(),
    llm: { backend: form.llmBackend, model: form.llmModel.trim(), headers: listToHeaders(form.llmHeaders) },
    transcription: { backend: form.transcriptionBackend, model: form.transcriptionModel.trim(), config: buildNamespacedConfig(selectedSttProviderType.value, form.sttProviderConfig) },
    tts: {
      backend: form.ttsBackend,
      model: form.ttsModel.trim(),
      voice: form.ttsVoice.trim(),
      config: buildNamespacedConfig(selectedTtsProviderType.value, form.ttsProviderConfig),
    },
    mcpServers: form.mcpServers,
    skills: form.skills,
    tags: form.tags.length ? form.tags : undefined,
    contextGuard: form.contextGuardEnabled ? {
      enabled: true,
      strategy: form.contextGuardStrategy,
      maxTurns: parseInt(form.contextGuardMaxTurns) || 0,
      maxTokens: parseInt(form.contextGuardMaxTokens) || 0,
    } : undefined,
    a2a: form.a2aEnabled ? { enabled: true } : undefined,
  }
  try {
    if (isEdit.value) {
      await agentsApi.update(editId.value, data)
    } else {
      await agentsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
