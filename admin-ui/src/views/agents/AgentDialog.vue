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
            <p class="text-[10px] text-arena-500 mt-1">Saves this agent's final output under the given key. Other agents can reference it with <code class="text-arena-300 bg-piedra-800 px-0.5 rounded">{key_name}</code> in their system prompt.</p>
          </div>
        </div>
      </details>

      <!-- LLM -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>LLM</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 grid grid-cols-2 gap-3">
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
                  <option v-for="b in store.backends" :key="b.id" :value="b.id">{{ b.name }} ({{ b.type }})</option>
                </FormSelect>
              </div>
              <div>
                <FormLabel label="Model" />
                <FormInput v-model="form.transcriptionModel" placeholder="whisper-1" />
              </div>
            </div>
          </div>
          <hr class="border-piedra-700/30" />
          <div class="space-y-3">
            <h4 class="text-[10px] font-medium text-arena-500 uppercase tracking-wider">Text-to-Speech (TTS)</h4>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <FormLabel label="Backend" />
                <FormSelect v-model="form.ttsBackend">
                  <option value="">(none)</option>
                  <option v-for="b in store.backends" :key="b.id" :value="b.id">{{ b.name }} ({{ b.type }})</option>
                </FormSelect>
              </div>
              <div>
                <FormLabel label="Model" />
                <FormInput v-model="form.ttsModel" placeholder="tts-1" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <FormLabel label="Voice" />
                <FormInput v-model="form.ttsVoice" placeholder="alloy" />
              </div>
              <div>
                <FormLabel label="Speed" />
                <FormInput v-model="form.ttsSpeed" type="number" placeholder="1.0" />
              </div>
            </div>
          </div>
        </div>
      </details>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { agentsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)
const tagInput = ref('')

const form = reactive({
  name: '',
  description: '',
  outputKey: '',
  systemPrompt: '',
  llmBackend: '',
  llmModel: '',
  mcpServers: [],
  tags: [],
  transcriptionBackend: '',
  transcriptionModel: '',
  ttsBackend: '',
  ttsModel: '',
  ttsVoice: '',
  ttsSpeed: '',
})

function toggleMcp(id) {
  const idx = form.mcpServers.indexOf(id)
  if (idx === -1) form.mcpServers.push(id)
  else form.mcpServers.splice(idx, 1)
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

function open(agent = null) {
  isEdit.value = !!agent
  editId.value = agent?.id || null
  form.name = agent?.name || ''
  form.description = agent?.description || ''
  form.outputKey = agent?.outputKey || ''
  form.systemPrompt = agent?.systemPrompt || ''
  form.llmBackend = agent?.llm?.backend || ''
  form.llmModel = agent?.llm?.model || ''
  form.mcpServers = [...(agent?.mcpServers || [])]
  form.tags = [...(agent?.tags || [])]
  form.transcriptionBackend = agent?.transcription?.backend || ''
  form.transcriptionModel = agent?.transcription?.model || ''
  form.ttsBackend = agent?.tts?.backend || ''
  form.ttsModel = agent?.tts?.model || ''
  form.ttsVoice = agent?.tts?.voice || ''
  form.ttsSpeed = agent?.tts?.speed || ''
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    outputKey: form.outputKey.trim(),
    systemPrompt: form.systemPrompt.trim(),
    llm: { backend: form.llmBackend, model: form.llmModel.trim() },
    transcription: { backend: form.transcriptionBackend, model: form.transcriptionModel.trim() },
    tts: {
      backend: form.ttsBackend,
      model: form.ttsModel.trim(),
      voice: form.ttsVoice.trim(),
      speed: parseFloat(form.ttsSpeed) || 0,
    },
    mcpServers: form.mcpServers,
    tags: form.tags.length ? form.tags : undefined,
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
