<template>
  <AppDialog ref="dialogRef" :title="isEdit ? 'Edit Cron Job' : 'New Cron Job'" @save="save">
    <div class="space-y-4">
      <div class="flex items-center gap-4">
        <div class="flex-1">
          <FormLabel label="Name" :required="true" />
          <FormInput v-model="form.name" placeholder="daily-summary" :required="true" />
        </div>
        <label class="flex flex-col items-center gap-1 cursor-pointer flex-shrink-0 pt-1">
          <span class="text-[10px] text-arena-500">Enabled</span>
          <div class="relative">
            <input type="checkbox" v-model="form.enabled" class="sr-only peer" />
            <div class="w-9 h-5 bg-piedra-700 rounded-full peer-checked:bg-sol-500/60 transition-colors" />
            <div class="absolute left-0.5 top-0.5 w-4 h-4 bg-arena-400 rounded-full peer-checked:translate-x-4 peer-checked:bg-white transition-transform" />
          </div>
        </label>
      </div>
      <div>
        <FormLabel label="Description" />
        <FormInput v-model="form.description" placeholder="What this cron job does..." />
      </div>
      <div>
        <FormLabel label="Schedule" :required="true" />
        <FormInput v-model="form.schedule" placeholder="0 9 * * *" :required="true" mono />
        <p class="text-[10px] text-arena-500 mt-1">Standard cron expression (min hour day month weekday)</p>
      </div>
      <div>
        <FormLabel label="Agent" :required="true" />
        <FormSelect v-model="form.agentId">
          <option value="" disabled>Select an agent</option>
          <option v-for="a in store.agents" :key="a.id" :value="a.id">{{ a.name || a.id }}</option>
        </FormSelect>
      </div>
      <div>
        <FormLabel label="Prompt" :required="true" />
        <textarea v-model="form.prompt" rows="3" class="w-full bg-piedra-800 border border-piedra-700 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-sol-500 focus:border-sol-500 outline-none resize-y" placeholder="The prompt to send to the agent..." required />
      </div>
    </div>
  </AppDialog>
</template>

<script setup>
import { ref, reactive, inject } from 'vue'
import { useDataStore } from '../../lib/stores/data.js'
import { cronsApi } from '../../lib/api/index.js'
import AppDialog from '../../components/AppDialog.vue'
import FormInput from '../../components/FormInput.vue'
import FormSelect from '../../components/FormSelect.vue'
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
  schedule: '',
  agentId: '',
  prompt: '',
  enabled: true,
})

function open(cron = null) {
  isEdit.value = !!cron
  editId.value = cron?.id || null
  form.name = cron?.name || ''
  form.description = cron?.description || ''
  form.schedule = cron?.schedule || ''
  form.agentId = cron?.agentId || ''
  form.prompt = cron?.prompt || ''
  form.enabled = cron?.enabled ?? true
  dialogRef.value?.open()
}

async function save() {
  const data = {
    name: form.name.trim(),
    description: form.description.trim(),
    schedule: form.schedule.trim(),
    agentId: form.agentId,
    prompt: form.prompt.trim(),
    enabled: form.enabled,
  }
  try {
    if (isEdit.value) {
      await cronsApi.update(editId.value, data)
    } else {
      await cronsApi.create(data)
    }
    dialogRef.value?.close()
    emit('saved')
  } catch (e) {
    toast.error(e.message)
  }
}

defineExpose({ open })
</script>
