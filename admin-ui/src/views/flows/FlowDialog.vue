<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Flow' : 'New Flow'" size="2xl" @save="save">
    <div class="space-y-4">
      <div class="grid grid-cols-3 gap-4">
        <div>
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="my-workflow" :required="true" />
        </div>
        <div class="col-span-2">
          <FormLabel label="Description" />
          <FormInput v-model="form.description" placeholder="What this flow does..." />
        </div>
      </div>
      <FlowCanvas
        v-model="form.root"
        :agents="store.agents"
      />
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { flowsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormLabel from '../../components/FormLabel.vue'
import FlowCanvas from './FlowCanvas.vue'

const emit = defineEmits(['saved'])
const store = useDataStore()
const dialogRef = ref(null)
const editId = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  description: '',
  root: null,
})

function open(flow = null) {
  isEdit.value = !!flow
  editId.value = flow?.id || null
  form.name = flow?.name || ''
  form.description = flow?.description || ''
  form.root = flow ? JSON.parse(JSON.stringify(flow.root)) : null
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    root: cleanStep(form.root),
  }
  try {
    if (isEdit.value) {
      await flowsApi.update(editId.value, data)
    } else {
      await flowsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    alert('Error: ' + e.message)
  }
}

function cleanStep(step) {
  const clean = { type: step.type }
  if (step.type === 'agent') {
    clean.agentId = step.agentId
  } else {
    clean.steps = (step.steps || []).map(cleanStep)
    if (step.type === 'loop') {
      clean.maxIterations = step.maxIterations || 0
    }
  }
  return clean
}

defineExpose({ open })
</script>
