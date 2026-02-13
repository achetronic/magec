<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Command' : 'New Command'" @save="save">
    <div class="space-y-4">
      <div>
        <FormLabel label="Name" :required="true" />
        <FormInput v-model="form.name" placeholder="daily-summary" :required="true" />
      </div>
      <div>
        <FormLabel label="Description" />
        <FormInput v-model="form.description" placeholder="What this command does..." />
      </div>
      <div>
        <FormLabel label="Prompt" :required="true" />
        <textarea v-model="form.prompt" rows="4" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" placeholder="The prompt to send to the agent..." required />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { commandsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'

const emit = defineEmits(['saved'])
const toast = inject('toast')
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  description: '',
  prompt: '',
})

function open(cmd = null) {
  isEdit.value = !!cmd
  editId.value = cmd?.id || null
  form.name = cmd?.name || ''
  form.description = cmd?.description || ''
  form.prompt = cmd?.prompt || ''
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    prompt: form.prompt.trim(),
  }
  try {
    if (isEdit.value) {
      await commandsApi.update(editId.value, data)
    } else {
      await commandsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
