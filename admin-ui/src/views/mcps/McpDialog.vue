<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit MCP Server' : 'New MCP Server'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="home-assistant" :required="true" />
      </div>
      <div>
        <FormLabel label="Type" />
        <FormSelect v-model="form.type">
          <option value="http">HTTP</option>
          <option value="stdio">Stdio</option>
        </FormSelect>
      </div>
      <div v-if="form.type === 'http'">
        <FormLabel label="Endpoint" />
        <FormInput v-model="form.endpoint" placeholder="http://localhost:8080/mcp" />
      </div>
      <template v-if="form.type === 'stdio'">
        <div>
          <FormLabel label="Command" />
          <FormInput v-model="form.command" placeholder="uvx" />
        </div>
        <div>
          <FormLabel label="Args (comma-separated)" />
          <FormInput v-model="form.argsStr" placeholder="mcp-server-sqlite, --db-path, /data/db" />
        </div>
      </template>
      <div>
        <FormLabel label="System Prompt" />
        <textarea
          v-model="form.systemPrompt"
          rows="2"
          class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y"
          placeholder="Instructions for the LLM about this MCP..."
        />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { mcpsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'

const emit = defineEmits(['saved'])
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  type: 'http',
  endpoint: '',
  command: '',
  argsStr: '',
  systemPrompt: '',
})

function open(mcp = null) {
  isEdit.value = !!mcp
  editId.value = mcp?.id || null
  form.name = mcp?.name || ''
  form.type = mcp?.type || 'http'
  form.endpoint = mcp?.endpoint || ''
  form.command = mcp?.command || ''
  form.argsStr = (mcp?.args || []).join(', ')
  form.systemPrompt = mcp?.systemPrompt || ''
  dialogRef.value?.open()
}

async function save() {
  const data = { name: form.name, type: form.type, systemPrompt: form.systemPrompt.trim() }
  if (form.type === 'http') {
    data.endpoint = form.endpoint.trim()
  } else {
    data.command = form.command.trim()
    data.args = form.argsStr ? form.argsStr.split(',').map(s => s.trim()).filter(Boolean) : []
  }
  try {
    if (isEdit.value) {
      await mcpsApi.update(editId.value, data)
    } else {
      await mcpsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

defineExpose({ open })
</script>
