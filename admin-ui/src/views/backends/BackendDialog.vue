<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Backend' : 'New Backend'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="ollama" :required="true" />
      </div>
      <div>
        <FormLabel label="Type" />
        <FormSelect v-model="form.type">
          <option value="openai">OpenAI-compatible</option>
          <option value="anthropic">Anthropic</option>
          <option value="gemini">Gemini</option>
        </FormSelect>
      </div>
      <div>
        <FormLabel label="URL" />
        <FormInput v-model="form.url" placeholder="http://localhost:11434/v1" />
      </div>
      <div>
        <FormLabel label="API Key" />
        <FormInput v-model="form.apiKey" type="password" placeholder="sk-..." />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { backendsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
import FormLabel from '../../components/FormLabel.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  type: 'openai',
  url: '',
  apiKey: '',
})

function open(backend = null) {
  isEdit.value = !!backend
  editId.value = backend?.id || null
  form.name = backend?.name || ''
  form.type = backend?.type || 'openai'
  form.url = backend?.url || ''
  form.apiKey = backend?.apiKey || ''
  dialogRef.value?.open()
}

async function save() {
  const data = { name: form.name, type: form.type, url: form.url, apiKey: form.apiKey }
  try {
    if (isEdit.value) {
      await backendsApi.update(editId.value, data)
    } else {
      await backendsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
