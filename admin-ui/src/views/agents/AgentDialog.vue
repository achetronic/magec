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

      <!-- System Prompt -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>System Prompt</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4">
          <textarea v-model="form.systemPrompt" rows="3" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" placeholder="Custom system prompt..." />
        </div>
      </details>

      <!-- LLM -->
      <fieldset class="border border-piedra-700/40 rounded-xl p-4 space-y-3">
        <legend class="text-xs font-medium text-arena-400 px-1.5">LLM</legend>
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
      </fieldset>

      <!-- Memory -->
      <details class="group border border-piedra-700/40 rounded-xl">
        <summary class="flex items-center justify-between px-4 py-3 cursor-pointer select-none text-xs font-medium text-arena-400 hover:text-arena-300">
          <span>Memory</span>
          <Icon name="chevronDown" size="md" class="text-arena-500 transition-transform group-open:rotate-180" />
        </summary>
        <div class="px-4 pb-4 grid grid-cols-2 gap-3">
          <div>
            <FormLabel label="Session" />
            <FormSelect v-model="form.memorySession">
              <option value="">(none)</option>
              <option v-for="m in sessionProviders" :key="m.id" :value="m.id">{{ m.name }}</option>
            </FormSelect>
          </div>
          <div>
            <FormLabel label="Long-term" />
            <FormSelect v-model="form.memoryLongTerm">
              <option value="">(none)</option>
              <option v-for="m in longTermProviders" :key="m.id" :value="m.id">{{ m.name }}</option>
            </FormSelect>
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
          <div v-if="store.mcps.length" class="flex flex-wrap gap-2">
            <label
              v-for="m in store.mcps" :key="m.id"
              class="flex items-center gap-1.5 px-2.5 py-1 bg-piedra-800 rounded-lg cursor-pointer hover:bg-piedra-700 transition-colors"
            >
              <input type="checkbox" :value="m.id" v-model="form.mcpServers" class="rounded border-piedra-600 bg-piedra-800 text-sol-500 focus:ring-sol-500" />
              <span class="text-xs text-arena-300">{{ m.name }}</span>
            </label>
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
import { ref, reactive, computed } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { agentsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'
import Icon from '../../components/Icon.vue'

const emit = defineEmits(['saved'])
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  description: '',
  systemPrompt: '',
  llmBackend: '',
  llmModel: '',
  memorySession: '',
  memoryLongTerm: '',
  mcpServers: [],
  transcriptionBackend: '',
  transcriptionModel: '',
  ttsBackend: '',
  ttsModel: '',
  ttsVoice: '',
  ttsSpeed: '',
})

const sessionProviders = computed(() => store.memory.filter(m => m.category === 'session'))
const longTermProviders = computed(() => store.memory.filter(m => m.category === 'longterm'))

function open(agent = null) {
  isEdit.value = !!agent
  editId.value = agent?.id || null
  form.name = agent?.name || ''
  form.description = agent?.description || ''
  form.systemPrompt = agent?.systemPrompt || ''
  form.llmBackend = agent?.llm?.backend || ''
  form.llmModel = agent?.llm?.model || ''
  form.memorySession = agent?.memory?.session || ''
  form.memoryLongTerm = agent?.memory?.longTerm || ''
  form.mcpServers = [...(agent?.mcpServers || [])]
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
    memory: { session: form.memorySession, longTerm: form.memoryLongTerm },
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
    alert('Error: ' + e.message)
  }
}

defineExpose({ open })
</script>
